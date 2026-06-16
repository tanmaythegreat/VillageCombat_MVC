package Battle

import (
	"Village_combat/GO/Database"
	"Village_combat/GO/Models"
	"context"
	"log"
	"math"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type ConnManager struct {
	Mu          sync.RWMutex
	Connections map[string]struct {
		Conn *websocket.Conn
		Mu   *sync.Mutex
	}
}

var Manager = &ConnManager{
	Mu: sync.RWMutex{},
	Connections: make(map[string]struct {
		Conn *websocket.Conn
		Mu   *sync.Mutex
	}),
}

type TroopSpawn struct {
	TroopID           string  `json:"troop_id"`
	TroopLevel        int     `json:"troop_level"`
	SpawnedByAttacker bool    `json:"spawned_by_attacker"`
	SpawnedAt_X       int     `json:"spawnedAt_X"`
	SpawnedAt_Y       int     `json:"spawnedAt_Y"`
	SpawnTime         float64 `json:"spawn_time"`
}

type BattleState struct {
	mu                 sync.Mutex
	TroopSpawns        []TroopSpawn
	AliveTroopAttacker []struct {
		TroopIndex      int //index in TroopSpawns slice
		Current_X       float64
		Current_Y       float64
		Config          Models.TroopConfig
		LevelStat       Models.TroopLevelStats
		HealthRemaining int
	}
	AliveTroopDefender []struct {
		TroopIndex      int //index in TroopSpawns slice
		Current_X       float64
		Current_Y       float64
		Config          Models.TroopConfig
		LevelStat       Models.TroopLevelStats
		HealthRemaining int
	}
	Buildings []struct {
		Placed_Building Models.PlacedBuilding
		Defender        *struct {
			LevelStat Models.DefenseBuildingLevelStats
			Stat      Models.DefenseBuildingStats
		}
	}
	AliveBuildings []struct {
		BuildingIndex   int `json:"BuildingIndex"` //index in the Buildings slice
		HealthRemaining int `json:"HealthRemaining"`
	}
	DeathMap  []int
	StartTime time.Time
}

func StartMatch(attackerID string, defenderID string) {
	Manager.Mu.RLock()
	attackerConn, attackerOnline := Manager.Connections[attackerID]
	defenderConn, defenderOnline := Manager.Connections[defenderID]
	Manager.Mu.RUnlock()

	state := &BattleState{
		mu:          sync.Mutex{},
		TroopSpawns: make([]TroopSpawn, 0),
		AliveTroopAttacker: make([]struct {
			TroopIndex      int
			Current_X       float64
			Current_Y       float64
			Config          Models.TroopConfig
			LevelStat       Models.TroopLevelStats
			HealthRemaining int
		}, 0),
		Buildings: make([]struct {
			Placed_Building Models.PlacedBuilding
			Defender        *struct {
				LevelStat Models.DefenseBuildingLevelStats
				Stat      Models.DefenseBuildingStats
			}
		}, 0),
		AliveBuildings: []struct {
			BuildingIndex   int `json:"BuildingIndex"`
			HealthRemaining int `json:"HealthRemaining"`
		}(make([]struct {
			BuildingIndex   int
			HealthRemaining int
		}, 0)),
		StartTime: time.Now(),
		DeathMap:  make([]int, 0),
	}
	placedBuildings, err := Database.GetPlacedBuildings(defenderID)
	if err != nil {
		// TODO : stop the battle
	}
	ToSend := make([]Models.PlacedBuilding, 0, len(placedBuildings))
	for _, building := range placedBuildings {
		if building.Level == 0 {
			continue
		}
		ToSend = append(ToSend, building)
		health, err := Database.GetBuildingHealth(building.BuildingID, building.Level)
		if err != nil {
			// TODO : its game over
		}
		if Database.BuildingID_Category[building.BuildingID] == Models.Defense {
			levelStat, stat, err := Database.GetDefenceBuildingStatAndLevelStat(building.BuildingID, building.Level)
			if err != nil {
				// TODO : end the battle, its all over :(
			}
			state.Buildings = append(state.Buildings, struct {
				Placed_Building Models.PlacedBuilding
				Defender        *struct {
					LevelStat Models.DefenseBuildingLevelStats
					Stat      Models.DefenseBuildingStats
				}
			}{Placed_Building: building, Defender: &struct {
				LevelStat Models.DefenseBuildingLevelStats
				Stat      Models.DefenseBuildingStats
			}{LevelStat: levelStat, Stat: stat}})
		} else {
			state.Buildings = append(state.Buildings, struct {
				Placed_Building Models.PlacedBuilding
				Defender        *struct {
					LevelStat Models.DefenseBuildingLevelStats
					Stat      Models.DefenseBuildingStats
				}
			}{Placed_Building: building, Defender: nil})
		}
		if !building.IsBroken {
			state.AliveBuildings = append(state.AliveBuildings, struct {
				BuildingIndex   int `json:"BuildingIndex"`
				HealthRemaining int `json:"HealthRemaining"`
			}{BuildingIndex: len(ToSend) - 1, HealthRemaining: health})
		}
	}
	attackerConn.Conn.WriteJSON(map[string]interface{}{
		"msg_type":          "battle_start",
		"defender_building": ToSend,
		"defender_id":       defenderID,
		"alive_buildings":   state.AliveBuildings,
	})
	if defenderOnline {
		err := defenderConn.Conn.WriteJSON(map[string]interface{}{
			"msg_type":          "incoming_attack",
			"defender_building": ToSend,
			"defender_id":       defenderID,
			"alive_buildings":   state.AliveBuildings,
		})
		if err != nil {

		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	WriteMU := sync.Mutex{} // this is write lock for connections only for these 3 functions
	if attackerOnline {
		go readPlayerMessages(ctx, attackerConn, state, true, attackerID, defenderID, defenderOnline, defenderConn.Conn, &WriteMU, cancel)
	}
	if defenderOnline {
		go readPlayerMessages(ctx, defenderConn, state, false, defenderID, defenderID, attackerOnline, attackerConn.Conn, &WriteMU, cancel)
	}
	runSimulation(ctx, state, attackerConn.Conn, defenderConn.Conn, attackerOnline, defenderOnline, &WriteMU)

	Database.SetUserBattleStatus(attackerID, false)
	Database.SetUserBattleStatus(defenderID, false)
	if attackerOnline {
		attackerConn.Conn.SetReadDeadline(time.Now())
	}
	if defenderOnline {
		attackerConn.Conn.SetReadDeadline(time.Now())
	}
	tx := Database.DB.Begin()
	for _, troopAtkr := range state.AliveTroopAttacker {
		err := Database.AddTroopsToUser(attackerID, state.TroopSpawns[troopAtkr.TroopIndex].TroopID, state.TroopSpawns[troopAtkr.TroopIndex].TroopLevel, 1, tx)
		if err != nil {
			log.Println("Error occurred while Adding troop after battle")
			tx.Rollback()
			break
		}
	}
	tx = Database.DB.Begin()
	for _, troopdfndr := range state.AliveTroopDefender {
		err := Database.AddTroopsToUser(defenderID, state.TroopSpawns[troopdfndr.TroopIndex].TroopID, state.TroopSpawns[troopdfndr.TroopIndex].TroopLevel, 1, tx)
		if err != nil {
			log.Println("Error occurred while Adding troop after battle")
			tx.Rollback()
			break
		}
	}

	DeathMap := make(map[string]int)
	DeathIDArray := make([]string, 0)
	for _, i := range state.DeathMap {
		DeathIDArray = append(DeathIDArray, state.Buildings[i].Placed_Building.ID)
		_, exist := DeathMap[state.Buildings[i].Placed_Building.BuildingID]
		if exist {
			DeathMap[state.Buildings[i].Placed_Building.BuildingID] += 1
		} else {
			DeathMap[state.Buildings[i].Placed_Building.BuildingID] = 1
		}
	}
	err = Database.SetBrokenBuildings(defenderID, DeathIDArray, true)
	if err != nil {
		// TODO : building dint break
	}
	var totalElixirB = 0
	var totalGoldB = 0
	var totalDarkelixirB = 0
	for _, building := range state.Buildings {
		if building.Placed_Building.BuildingID == Database.GoldStorage_ID {
			totalGoldB += 1
		} else if building.Placed_Building.BuildingID == Database.ElixirStorage_ID {
			totalElixirB += 1
		} else if building.Placed_Building.BuildingID == Database.DarkElixirStorage_ID {
			totalDarkelixirB += 1
		}
	}
	elixir, _ := DeathMap[Database.ElixirStorage_ID]
	gold, _ := DeathMap[Database.GoldStorage_ID]
	dark_elixir, _ := DeathMap[Database.DarkElixirDrill_ID]
	defender, err := Database.GetUserData(defenderID)
	if err != nil {
		// TODO : what if defender deleted the account during battle !!
	}

	var goldLooted int
	if totalGoldB != 0 {
		goldLooted = (gold * defender.CurrentGold) / totalGoldB
	}
	var elixirLooted int
	if totalElixirB != 0 {
		elixirLooted = (elixir * defender.CurrentElixir) / totalElixirB
	}
	var darkElixirLooted int
	if totalDarkelixirB != 0 {
		darkElixirLooted = (dark_elixir * defender.CurrentDarkElixir) / totalDarkelixirB
	}
	battleHistory := Models.BattleHistory{
		AttackerID:       attackerID,
		DefenderID:       defenderID,
		ElixirLooted:     elixirLooted,
		GoldLooted:       goldLooted,
		DarkElixirLooted: darkElixirLooted,
		FoughtAt:         time.Now(),
		DoDefenderKnow:   defenderOnline,
	}
	battleId, err := Database.InsertBattleHistory(battleHistory)
	if err != nil {
		// TODO : Couldn't insert
	}
	for buildingId, count := range DeathMap {
		err := Database.InsertBrokenBuildingBattleHistory(battleId, buildingId, count)
		if err != nil {
			// TODO :
		}
	}
	TroopLossAttacker := make(map[string]int)
	for _, troopSpawn := range state.TroopSpawns {
		if troopSpawn.SpawnedByAttacker {
			_, exist := TroopLossAttacker[troopSpawn.TroopID]
			if exist {
				TroopLossAttacker[troopSpawn.TroopID] += 1
			} else {
				TroopLossAttacker[troopSpawn.TroopID] = 1
			}
		}
	}
	for _, s := range state.AliveTroopAttacker {
		TroopLossAttacker[state.TroopSpawns[s.TroopIndex].TroopID] -= 1
	}
	for troopId, count := range TroopLossAttacker {
		err := Database.InsertTroopLoosesBattleHistory(battleId, troopId, count)
		if err != nil {
			// TODO : handle it
		}
	}
	defenderName, err := Database.GetUsername(defenderID)
	if err != nil {
		// TODO : what to do
	}
	attackerName, err := Database.GetUsername(attackerID)
	if err != nil {
		// TODO : what to do
	}
	attackerConn.Mu.Lock()
	attackerConn.Conn.WriteJSON(map[string]interface{}{
		"msg_type":            "battle_over",
		"battle_outcome":      battleHistory,
		"attacker_troop_loss": TroopLossAttacker,
		"buildings_broken":    DeathMap,
		"opponent_username":   defenderName,
	})
	attackerConn.Mu.Unlock()
	if defenderOnline {
		defenderConn.Mu.Lock()
		defenderConn.Conn.WriteJSON(map[string]interface{}{
			"msg_type":            "battle_over",
			"battle_outcome":      battleHistory,
			"attacker_troop_loss": TroopLossAttacker,
			"buildings_broken":    DeathMap,
			"opponent_username":   attackerName,
		})
		defenderConn.Mu.Unlock()
	}
}

type SpawnMessage struct {
	Action     string `json:"action"`
	TroopID    string `json:"troop_id"`
	TroopLevel int    `json:"troop_level"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
}

func readPlayerMessages(ctx context.Context, conn struct {
	Conn *websocket.Conn
	Mu   *sync.Mutex
}, state *BattleState, isAttacker bool, userID string, defenderId string, otherOnline bool, otherConn *websocket.Conn, WriteMU *sync.Mutex, cancel context.CancelFunc) {
	conn.Mu.Lock()
	defer conn.Mu.Unlock()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			var msg SpawnMessage
			err := conn.Conn.ReadJSON(&msg)
			if err != nil {
				cancel()
				return
			}
			if msg.Action == "spawn_troop" {
				elapsed := time.Since(state.StartTime).Seconds()
				nearByBuildings, err := Database.GetNearByBuildings(defenderId, msg.X, msg.Y)
				if err != nil {
					errPayload := []byte(`{"status": "error", "message": "Could not spawn troop here."}`)
					WriteMU.Lock()
					err = conn.Conn.WriteMessage(websocket.TextMessage, errPayload)
					WriteMU.Unlock()
					if err != nil {
						log.Println("Failed to send message to client:", err)
					}
					continue
				}
				housing_space := 1 // TODO : housing space
				hasCollision := false
				for _, b := range nearByBuildings {
					if msg.X < b.At_x+b.Size_x && msg.X+housing_space > b.At_x &&
						msg.Y < b.At_y+b.Size_y && msg.Y+housing_space > b.At_y {
						hasCollision = true
						break
					}
				}
				if hasCollision {
					errPayload := []byte(`{"status": "error", "message": "Could not spawn troop here."}`)
					WriteMU.Lock()
					err = conn.Conn.WriteMessage(websocket.TextMessage, errPayload)
					WriteMU.Unlock()
					if err != nil {
						log.Println("Failed to send message to client:", err)
					}
					continue
				}

				tx := Database.DB.Begin()
				success, err := Database.SubtractTroopsOfUser(userID, msg.TroopID, msg.TroopLevel, 1, tx)
				if err != nil || !success {
					tx.Rollback()
					errPayload := []byte(`{"status": "error", "message": "Could not spawn troop here."}`)
					WriteMU.Lock()
					err = conn.Conn.WriteMessage(websocket.TextMessage, errPayload)
					WriteMU.Unlock()
					if err != nil {
						log.Println("Failed to send message to client:", err)
					}
					continue
				}
				tx.Commit()
				troop := TroopSpawn{
					TroopID:           msg.TroopID,
					TroopLevel:        msg.TroopLevel,
					SpawnedByAttacker: isAttacker,
					SpawnedAt_X:       msg.X,
					SpawnedAt_Y:       msg.Y,
					SpawnTime:         elapsed,
				}
				lvlStat := Database.TroopLevelDetails[struct {
					ID    string
					Level int
				}{ID: troop.TroopID, Level: troop.TroopLevel}]

				var aliveTrop = struct {
					TroopIndex      int //index in TroopSpawns slice
					Current_X       float64
					Current_Y       float64
					Config          Models.TroopConfig //they wont be nil
					LevelStat       Models.TroopLevelStats
					HealthRemaining int
				}{
					TroopIndex:      len(state.TroopSpawns),
					Current_X:       float64(troop.SpawnedAt_X),
					Current_Y:       float64(troop.SpawnedAt_Y),
					Config:          Database.TroopConfigs[troop.TroopID],
					LevelStat:       lvlStat,
					HealthRemaining: lvlStat.Health,
				}
				state.mu.Lock()
				state.TroopSpawns = append(state.TroopSpawns, troop)
				if isAttacker {
					state.AliveTroopAttacker = append(state.AliveTroopAttacker, aliveTrop)
				} else {
					state.AliveTroopDefender = append(state.AliveTroopDefender, aliveTrop)
				}
				state.mu.Unlock()
				WriteMU.Lock()
				if otherOnline {
					otherConn.WriteJSON(map[string]interface{}{
						"msg_type": "spawn_troop",
						"troop":    troop,
					})
				}
				conn.Conn.WriteJSON(map[string]interface{}{
					"msg_type": "spawn_troop",
					"troop":    troop,
				})
				WriteMU.Unlock()
			} else if msg.Action == "retreat" && isAttacker {
				cancel()
				return
			}
		}
	}
}
func runSimulation(
	ctx context.Context,
	state *BattleState,
	attackerConn *websocket.Conn,
	defenderConn *websocket.Conn,
	attackerOnline bool,
	defenderOnline bool,
	WriteMU *sync.Mutex,
) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			state.mu.Lock()
			update := simulate(state)
			state.mu.Unlock()

			WriteMU.Lock()
			if attackerOnline {
				attackerConn.WriteJSON(update)
			}
			if defenderOnline {
				defenderConn.WriteJSON(update)
			}
			WriteMU.Unlock()
		}
	}

}

func simulate(state *BattleState) map[string]interface{} {

	buildingDmgDone := make([]int, len(state.AliveBuildings))
	AttackertroopDmgDone := make([]int, len(state.AliveTroopAttacker))

	for i := 0; i < len(state.AliveTroopAttacker); i++ {
		troop := &state.AliveTroopAttacker[i]
		if !state.TroopSpawns[troop.TroopIndex].SpawnedByAttacker {
			continue
		}
		var bestBuildingIdx = -1
		var minDstSq = math.MaxFloat64
		var BestBuildingi = -1
		var prefCat Models.BuildingCategory
		hasPreferred := false
		if troop.Config.PreferredTarget != nil {
			prefCat = *troop.Config.PreferredTarget
			for _, ab := range state.AliveBuildings {
				if Database.BuildingID_Category[state.Buildings[ab.BuildingIndex].Placed_Building.ID] == prefCat {
					hasPreferred = true
					break
				}
			}
		}

		for j, ab := range state.AliveBuildings {
			b := &state.Buildings[ab.BuildingIndex]
			if hasPreferred && Database.BuildingID_Category[b.Placed_Building.BuildingID] != prefCat {
				continue
			}
			// TODO : account for building size
			dx := float64(b.Placed_Building.GridX) - troop.Current_X
			dy := float64(b.Placed_Building.GridY) - troop.Current_Y
			dstSq := dx*dx + dy*dy

			if dstSq < minDstSq {
				minDstSq = dstSq
				bestBuildingIdx = ab.BuildingIndex
				BestBuildingi = j
			}
		}

		if bestBuildingIdx == -1 {
			// TODO : what to do if nothing to attack
			continue
		}

		targetBuilding := &state.Buildings[bestBuildingIdx]
		dist := math.Sqrt(minDstSq)

		if dist <= troop.Config.AttackRange {
			dmg := float64(troop.LevelStat.DamagePerShot) / troop.Config.AttackSpeedSeconds // TODO : think how can i quantise it
			intDmg := int(dmg)
			if intDmg <= 0 && troop.LevelStat.DamagePerShot > 0 {
				intDmg = 1 // guarantee minimum damage
			}
			buildingDmgDone[BestBuildingi] += intDmg
		} else {
			moveDist := troop.Config.MovementSpeed
			if moveDist > dist {
				moveDist = dist
			}
			ratio := moveDist / dist
			dx := float64(targetBuilding.Placed_Building.GridX) - troop.Current_X
			dy := float64(targetBuilding.Placed_Building.GridY) - troop.Current_Y
			troop.Current_X += dx * ratio
			troop.Current_Y += dy * ratio
		}
	}

	for _, ab := range state.AliveBuildings {
		b := &state.Buildings[ab.BuildingIndex]
		if b.Defender == nil {
			continue
		}

		var bestTroopi = -1
		var minDstSq = math.MaxFloat64
		rangeSq := b.Defender.Stat.AttackRange * b.Defender.Stat.AttackRange

		for tIdx := 0; tIdx < len(state.AliveTroopAttacker); tIdx++ {
			t := &state.AliveTroopAttacker[tIdx]

			dx := t.Current_X - float64(b.Placed_Building.GridX)
			dy := t.Current_Y - float64(b.Placed_Building.GridY)
			dstSq := dx*dx + dy*dy

			if dstSq <= rangeSq && dstSq < minDstSq {
				minDstSq = dstSq
				bestTroopi = tIdx
			}
		}

		if bestTroopi != -1 {
			dmg := float64(b.Defender.LevelStat.DamagePerShot) / b.Defender.Stat.AttackSpeedSeconds
			intDmg := int(dmg)
			if intDmg <= 0 && b.Defender.LevelStat.DamagePerShot > 0 {
				intDmg = 1
			}

			if b.Defender.Stat.DamageType == Models.Splash {
				targetTroop := &state.AliveTroopAttacker[bestTroopi]
				const splashRadiusSq = 1.5 * 1.5 // currently there is no splash defence building , i will think about the radius if i add a splash one in future

				for tIdx := 0; tIdx < len(state.AliveTroopAttacker); tIdx++ {
					t := &state.AliveTroopAttacker[tIdx]
					dx := t.Current_X - targetTroop.Current_X
					dy := t.Current_Y - targetTroop.Current_Y
					dstSq := dx*dx + dy*dy
					if dstSq <= splashRadiusSq {
						AttackertroopDmgDone[tIdx] += intDmg
					}
				}
			} else {
				AttackertroopDmgDone[bestTroopi] += intDmg
			}
		}
	}

	for i := 0; i < len(state.AliveTroopDefender); i++ {
		troop := &state.AliveTroopDefender[i]

		var bestTargetIdx = -1
		var minDstSq = math.MaxFloat64

		for tIdx := 0; tIdx < len(state.AliveTroopAttacker); tIdx++ {
			t := &state.AliveTroopAttacker[tIdx]
			dx := t.Current_X - troop.Current_X
			dy := t.Current_Y - troop.Current_Y
			dstSq := dx*dx + dy*dy
			if dstSq < minDstSq {
				minDstSq = dstSq
				bestTargetIdx = tIdx
			}
		}

		if bestTargetIdx == -1 {
			continue
		}

		target := &state.AliveTroopAttacker[bestTargetIdx]
		dist := math.Sqrt(minDstSq)

		if dist <= troop.Config.AttackRange {
			dmg := float64(troop.LevelStat.DamagePerShot) / troop.Config.AttackSpeedSeconds
			intDmg := int(dmg)
			if intDmg <= 0 && troop.LevelStat.DamagePerShot > 0 {
				intDmg = 1
			}
			AttackertroopDmgDone[bestTargetIdx] += intDmg
		} else {
			moveDist := troop.Config.MovementSpeed
			if moveDist > dist {
				moveDist = dist
			}
			ratio := moveDist / dist
			dx := target.Current_X - troop.Current_X
			dy := target.Current_Y - troop.Current_Y
			troop.Current_X += dx * ratio
			troop.Current_Y += dy * ratio
		}
	}

	for bIdx, dmg := range buildingDmgDone {
		state.AliveBuildings[bIdx].HealthRemaining -= dmg
	}

	for tIdx, dmg := range AttackertroopDmgDone {
		state.AliveTroopAttacker[tIdx].HealthRemaining -= dmg
	}
	aliveBCount := 0
	BuildingDied := make([]int, 0)
	for i, ab := range state.AliveBuildings {
		if ab.HealthRemaining > 0 {
			state.AliveBuildings[aliveBCount] = ab
			aliveBCount++
		} else {
			BuildingDied = append(BuildingDied, i)
			state.DeathMap = append(state.DeathMap, ab.BuildingIndex)
		}
	}
	state.AliveBuildings = state.AliveBuildings[:aliveBCount]

	aliveTCount := 0
	AttackerTroopDied := make([]int, 0)
	for i, t := range state.AliveTroopAttacker {
		if t.HealthRemaining > 0 {
			state.AliveTroopAttacker[aliveTCount] = t
			aliveTCount++
		} else {
			AttackerTroopDied = append(AttackerTroopDied, i)
		}
	}
	state.AliveTroopAttacker = state.AliveTroopAttacker[:aliveTCount]
	return map[string]interface{}{
		"msg_type":            "battle_update",
		"building_damage":     buildingDmgDone,
		"troop_damage":        AttackertroopDmgDone,
		"building_died":       BuildingDied,
		"attacker_troop_died": AttackerTroopDied,
	}
}
