package battle

import (
	"Village_combat/models"
	"context"
	"encoding/json"
	"log"
	"math"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Connection struct {
	Conn *websocket.Conn
	Mu   *sync.Mutex
	Ch   chan []byte
}
type ConnManager struct {
	Mu          sync.RWMutex
	Connections map[string]Connection
}

var Manager = &ConnManager{
	Mu:          sync.RWMutex{},
	Connections: make(map[string]Connection),
}

type BattleState struct {
	mu                 sync.Mutex
	TroopSpawns        []models.TroopSpawn
	AliveTroopAttacker []struct {
		TroopIndex      int //index in TroopSpawns slice
		Current_X       float64
		Current_Y       float64
		Config          models.TroopConfig
		LevelStat       models.TroopLevelStats
		HealthRemaining int
	}
	AliveTroopDefender []struct {
		TroopIndex      int //index in TroopSpawns slice
		Current_X       float64
		Current_Y       float64
		Config          models.TroopConfig
		LevelStat       models.TroopLevelStats
		HealthRemaining int
	}
	Buildings []struct {
		Placed_Building models.PlacedBuilding
		Defender        *struct {
			LevelStat models.DefenseBuildingLevelStats
			Stat      models.DefenseBuildingStats
		}
	}
	AliveBuildings []struct {
		BuildingIndex   int `json:"BuildingIndex"` //index in the Buildings slice
		HealthRemaining int `json:"HealthRemaining"`
	}
	DeathMap  []int
	StartTime time.Time
}

// logErr logs a non-fatal error with context, and is a no-op if err is nil.
func logErr(context string, err error) {
	if err != nil {
		log.Printf("%s: %v", context, err)
	}
}

// notifyBattleFailed informs any online participant that the battle could not
// proceed, so a client never sits waiting on a battle that silently died server-side.
func notifyBattleFailed(attackerConn Connection, attackerOnline bool, defenderConn Connection, defenderOnline bool, reason string, attackerID, defenderID string) {
	payload := map[string]interface{}{
		"msg_type": "error",
		"message":  reason,
	}
	if attackerOnline {
		if err := attackerConn.Conn.WriteJSON(payload); err != nil {
			log.Println("notifyBattleFailed: failed to notify attacker:", err)
		}
	}
	if defenderOnline {
		if err := defenderConn.Conn.WriteJSON(payload); err != nil {
			log.Println("notifyBattleFailed: failed to notify defender:", err)
		}
	}
	logErr("StartMatch: could not clear attacker battle status", models.SetUserBattleStatus(attackerID, false))
	logErr("StartMatch: could not clear defender battle status", models.SetUserBattleStatus(defenderID, false))
}

func StartMatch(attackerID string, defenderID string) {
	Manager.Mu.RLock()
	attackerConn, attackerOnline := Manager.Connections[attackerID]
	defenderConn, defenderOnline := Manager.Connections[defenderID]
	Manager.Mu.RUnlock()
	logErr("StartMatch: could not clear attacker battle status", models.SetUserBattleStatus(attackerID, true))
	logErr("StartMatch: could not clear defender battle status", models.SetUserBattleStatus(defenderID, true))
	state := &BattleState{
		mu:          sync.Mutex{},
		TroopSpawns: make([]models.TroopSpawn, 0),
		AliveTroopAttacker: make([]struct {
			TroopIndex      int
			Current_X       float64
			Current_Y       float64
			Config          models.TroopConfig
			LevelStat       models.TroopLevelStats
			HealthRemaining int
		}, 0),
		Buildings: make([]struct {
			Placed_Building models.PlacedBuilding
			Defender        *struct {
				LevelStat models.DefenseBuildingLevelStats
				Stat      models.DefenseBuildingStats
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
	placedBuildings, err := models.GetPlacedBuildings(defenderID)
	if err != nil {
		log.Println("StartMatch: failed to load defender's placed buildings:", err)
		notifyBattleFailed(attackerConn, attackerOnline, defenderConn, defenderOnline, "Could not load village layout. Battle cancelled.", attackerID, defenderID)
		return
	}
	ToSend := make([]models.PlacedBuilding, 0, len(placedBuildings))

	for _, building := range placedBuildings {
		if building.Level == 0 {
			continue
		}
		ToSend = append(ToSend, building)
		health, err := models.GetBuildingHealth(building.BuildingID, building.Level)
		if err != nil {
			log.Printf("StartMatch: failed to get health for building %s (lvl %d): %v — treating as not-alive", building.BuildingID, building.Level, err)
		}
		if models.BuildingID_Category[building.BuildingID] == models.Defense {
			levelStat, stat, defErr := models.GetDefenceBuildingStatAndLevelStat(building.BuildingID, building.Level)
			if defErr != nil {
				log.Printf("StartMatch: failed to get defense stats for building %s (lvl %d): %v", building.BuildingID, building.Level, defErr)
			}
			state.Buildings = append(state.Buildings, struct {
				Placed_Building models.PlacedBuilding
				Defender        *struct {
					LevelStat models.DefenseBuildingLevelStats
					Stat      models.DefenseBuildingStats
				}
			}{Placed_Building: building, Defender: &struct {
				LevelStat models.DefenseBuildingLevelStats
				Stat      models.DefenseBuildingStats
			}{LevelStat: levelStat, Stat: stat}})
		} else {
			state.Buildings = append(state.Buildings, struct {
				Placed_Building models.PlacedBuilding
				Defender        *struct {
					LevelStat models.DefenseBuildingLevelStats
					Stat      models.DefenseBuildingStats
				}
			}{Placed_Building: building, Defender: nil})
		}

		if !building.IsBroken && err == nil {
			state.AliveBuildings = append(state.AliveBuildings, struct {
				BuildingIndex   int `json:"BuildingIndex"`
				HealthRemaining int `json:"HealthRemaining"`
			}{BuildingIndex: len(ToSend) - 1, HealthRemaining: health})
		}
	}
	if attackerOnline {
		if err := attackerConn.Conn.WriteJSON(map[string]interface{}{
			"msg_type":          "battle_start",
			"defender_building": ToSend,
			"defender_id":       defenderID,
			"alive_buildings":   state.AliveBuildings,
		}); err != nil {
			log.Println("StartMatch: failed to send battle_start to attacker, aborting match:", err)
			notifyBattleFailed(attackerConn, attackerOnline, defenderConn, defenderOnline, "Attacker could not send message.", attackerID, defenderID)
			return
		}
	}
	if defenderOnline {
		if err := defenderConn.Conn.WriteJSON(map[string]interface{}{
			"msg_type":          "incoming_attack",
			"defender_building": ToSend,
			"defender_id":       defenderID,
			"alive_buildings":   state.AliveBuildings,
		}); err != nil {
			log.Println("StartMatch: failed to send incoming_attack to defender:", err)
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
	var startTime = time.Now()
	runSimulation(ctx, state, attackerConn.Conn, defenderConn.Conn, attackerOnline, defenderOnline, &WriteMU, 1)
	var duration = int(time.Since(startTime).Seconds())
	logErr("StartMatch: could not clear attacker battle status (post-battle)", models.SetUserBattleStatus(attackerID, false))
	logErr("StartMatch: could not clear defender battle status (post-battle)", models.SetUserBattleStatus(defenderID, false))

	tx := models.DB.Begin()
	for _, troopAtkr := range state.AliveTroopAttacker {
		err := models.AddTroopsToUser(attackerID, state.TroopSpawns[troopAtkr.TroopIndex].TroopID, state.TroopSpawns[troopAtkr.TroopIndex].TroopLevel, 1, tx)
		if err != nil {
			log.Println("StartMatch: error occurred while adding surviving attacker troop:", err)
			tx.Rollback()
			break
		}
	}
	if commitErr := tx.Commit().Error; commitErr != nil {
		log.Println("StartMatch: failed to commit attacker troop return transaction:", commitErr)
	}
	tx = models.DB.Begin()
	for _, troopdfndr := range state.AliveTroopDefender {
		err := models.AddTroopsToUser(defenderID, state.TroopSpawns[troopdfndr.TroopIndex].TroopID, state.TroopSpawns[troopdfndr.TroopIndex].TroopLevel, 1, tx)
		if err != nil {
			log.Println("StartMatch: error occurred while adding surviving defender troop:", err)
			tx.Rollback()
			break
		}
	}
	if commitErr := tx.Commit().Error; commitErr != nil {
		log.Println("StartMatch: failed to commit defender troop return transaction:", commitErr)
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
	if err := models.SetBrokenBuildings(defenderID, DeathIDArray, true); err != nil {
		log.Println("StartMatch: failed to persist broken buildings:", err)
	}
	var totalElixirB = 0
	var totalGoldB = 0
	var totalDarkelixirB = 0
	for _, building := range state.Buildings {
		if building.Placed_Building.BuildingID == models.GoldStorage_ID {
			totalGoldB += 1
		} else if building.Placed_Building.BuildingID == models.ElixirStorage_ID {
			totalElixirB += 1
		} else if building.Placed_Building.BuildingID == models.DarkElixirStorage_ID {
			totalDarkelixirB += 1
		}
	}
	elixir, _ := DeathMap[models.ElixirStorage_ID]
	gold, _ := DeathMap[models.GoldStorage_ID]
	dark_elixir, _ := DeathMap[models.DarkElixirDrill_ID]
	defender, err := models.GetUserData(defenderID)
	if err != nil {
		log.Println("StartMatch: failed to load defender data for loot calculation, treating loot as 0:", err)
		defender = models.UserData{}
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
	defenderName, err := models.GetUsername(defenderID)
	if err != nil {
		log.Println("StartMatch: failed to load defender username:", err)
		defenderName = "Unknown"
	}
	attackerName, err := models.GetUsername(attackerID)
	if err != nil {
		log.Println("StartMatch: failed to load attacker username:", err)
		attackerName = "Unknown"
	}
	_, exist := DeathMap[models.TownHall_ID]
	battleHistory := models.BattleHistory{
		AttackerName:     attackerName,
		DefenderName:     defenderName,
		ElixirLooted:     elixirLooted,
		GoldLooted:       goldLooted,
		DarkElixirLooted: darkElixirLooted,
		FoughtAt:         startTime,
		BattleDuration:   duration,
		DoDefenderKnow:   defenderOnline,
		WinnerAttacker:   exist,
	}

	if exist {
		logErr("StartMatch: failed to adjust attacker power", models.AdjustAttackPower(attackerID, 1))
		logErr("StartMatch: failed to adjust defender power", models.AdjustDefencePower(defenderID, -1))
	} else {
		logErr("StartMatch: failed to adjust attacker power", models.AdjustAttackPower(attackerID, -1))
		logErr("StartMatch: failed to adjust defender power", models.AdjustDefencePower(defenderID, 1))
	}
	battleId, err := models.InsertBattleHistory(battleHistory)
	battleHistorySaved := err == nil
	if err != nil {
		log.Println("StartMatch: failed to insert battle history, skipping dependent records:", err)
	}

	if battleHistorySaved {
		for buildingId, count := range DeathMap {
			if err := models.InsertBrokenBuildingBattleHistory(battleId, buildingId, count); err != nil {
				log.Println("StartMatch: failed to insert broken building history:", err)
			}
		}
	}
	TroopLossAttacker := make(map[string]int)
	TroopLossDefender := make(map[string]int)
	for _, troopSpawn := range state.TroopSpawns {
		if troopSpawn.SpawnedByAttacker {
			_, exist := TroopLossAttacker[troopSpawn.TroopID]
			if exist {
				TroopLossAttacker[troopSpawn.TroopID] += 1
			} else {
				TroopLossAttacker[troopSpawn.TroopID] = 1
			}
		} else {
			_, exist := TroopLossDefender[troopSpawn.TroopID]
			if exist {
				TroopLossDefender[troopSpawn.TroopID] += 1
			} else {
				TroopLossDefender[troopSpawn.TroopID] = 1
			}
		}
	}
	for _, s := range state.AliveTroopAttacker {
		TroopLossAttacker[state.TroopSpawns[s.TroopIndex].TroopID] -= 1
	}
	for _, s := range state.AliveTroopDefender {
		TroopLossDefender[state.TroopSpawns[s.TroopIndex].TroopID] -= 1
	}
	if battleHistorySaved {
		for troopId, count := range TroopLossAttacker {
			if err := models.InsertTroopLoosesBattleHistory(battleId, troopId, count, true); err != nil {
				log.Println("StartMatch: failed to insert attacker troop loss history:", err)
			}
		}
		for troopId, count := range TroopLossDefender {
			if err := models.InsertTroopLoosesBattleHistory(battleId, troopId, count, false); err != nil {
				log.Println("StartMatch: failed to insert defender troop loss history:", err)
			}
		}
	}

	if _, err := models.AddUserGold(attackerID, goldLooted); err != nil {
		log.Println("StartMatch: failed to credit looted gold to attacker:", err)
	}
	if _, err := models.AddUserElixir(attackerID, elixirLooted); err != nil {
		log.Println("StartMatch: failed to credit looted elixir to attacker:", err)
	}
	if _, err := models.AddUserDarkElixir(attackerID, darkElixirLooted); err != nil {
		log.Println("StartMatch: failed to credit looted dark elixir to attacker:", err)
	}
	if _, err := models.AddUserGold(defenderID, -goldLooted); err != nil {
		log.Println("StartMatch: failed to credit looted gold to attacker:", err)
	}
	if _, err := models.AddUserElixir(defenderID, -elixirLooted); err != nil {
		log.Println("StartMatch: failed to credit looted elixir to attacker:", err)
	}
	if _, err := models.AddUserDarkElixir(defenderID, -darkElixirLooted); err != nil {
		log.Println("StartMatch: failed to credit looted dark elixir to attacker:", err)
	}

	if attackerOnline {
		if err := attackerConn.Conn.WriteJSON(map[string]interface{}{
			"msg_type":            "battle_over",
			"battle_id":           battleId,
			"battle_outcome":      battleHistory,
			"attacker_troop_loss": TroopLossAttacker,
			"defender_troop_loss": TroopLossDefender,
			"buildings_broken":    DeathMap,
			"opponent_username":   defenderName,
			"my_username":         attackerName,
		}); err != nil {
			log.Println("StartMatch: failed to send battle_over to attacker:", err)
		}
		attackerConn.Mu.Unlock()
	}
	if defenderOnline {
		if err := defenderConn.Conn.WriteJSON(map[string]interface{}{
			"msg_type":            "battle_over",
			"battle_id":           battleId,
			"battle_outcome":      battleHistory,
			"attacker_troop_loss": TroopLossAttacker,
			"defender_troop_loss": TroopLossDefender,
			"buildings_broken":    DeathMap,
			"opponent_username":   attackerName,
			"my_username":         defenderName,
		}); err != nil {
			log.Println("StartMatch: failed to send battle_over to defender:", err)
		}
		defenderConn.Mu.Unlock()
	}
	var initialBuildingPos models.InitialBuildingArray = make([]models.InitialBattleBuilding, len(state.Buildings))
	for i, building := range state.Buildings {
		initialBuildingPos[i].BuildingID = building.Placed_Building.BuildingID
		initialBuildingPos[i].Level = building.Placed_Building.Level
		initialBuildingPos[i].Grid_X = building.Placed_Building.GridX
		initialBuildingPos[i].Grid_Y = building.Placed_Building.GridY
	}
	if battleHistorySaved {
		err = models.SaveBattleRecord(&models.BattleRecord{
			BattleID:         battleId,
			TroopSpawns:      state.TroopSpawns,
			InitialBuildings: initialBuildingPos,
		})
		if err != nil {
			log.Println("StartMatch: failed to save battle record:", err)
		}
	} else {
		log.Println("StartMatch: skipping battle record save, battle history was not saved")
	}

}

type SpawnMessage struct {
	Action     string `json:"action"`
	TroopID    string `json:"troop_id"`
	TroopLevel int    `json:"troop_level"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
}

func readPlayerMessages(ctx context.Context, conn Connection, state *BattleState, isAttacker bool, userID string, defenderId string, otherOnline bool, otherConn *websocket.Conn, WriteMU *sync.Mutex, cancel context.CancelFunc) {
	conn.Mu.Lock() // unlocking is done outside this function
	var msg SpawnMessage
Loop:
	for {
		select {
		case <-ctx.Done():
			return
		case p := <-conn.Ch:
			err := json.Unmarshal(p, &msg)
			if err != nil {

				log.Println("Connection lost with client:", err)
				break Loop
			}
			if msg.Action == "spawn_troop" {
				elapsed := time.Since(state.StartTime).Seconds()
				nearByBuildings, err := models.GetNearByBuildings(defenderId, msg.X, msg.Y)
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

				tx := models.DB.Begin()
				success, err := models.SubtractTroopsOfUser(userID, msg.TroopID, msg.TroopLevel, 1, tx)
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
				if commitErr := tx.Commit().Error; commitErr != nil {
					log.Println("readPlayerMessages: failed to commit troop spawn transaction:", commitErr)
					continue
				}
				troop := models.TroopSpawn{
					TroopID:           msg.TroopID,
					TroopLevel:        msg.TroopLevel,
					SpawnedByAttacker: isAttacker,
					SpawnedAt_X:       msg.X,
					SpawnedAt_Y:       msg.Y,
					SpawnTime:         elapsed,
				}
				lvlStat := models.TroopLevelDetails[struct {
					ID    string
					Level int
				}{ID: troop.TroopID, Level: troop.TroopLevel}]

				var aliveTrop = struct {
					TroopIndex      int //index in TroopSpawns slice
					Current_X       float64
					Current_Y       float64
					Config          models.TroopConfig //they wont be nil
					LevelStat       models.TroopLevelStats
					HealthRemaining int
				}{
					TroopIndex:      len(state.TroopSpawns),
					Current_X:       float64(troop.SpawnedAt_X),
					Current_Y:       float64(troop.SpawnedAt_Y),
					Config:          models.TroopConfigs[troop.TroopID],
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
					if err := otherConn.WriteJSON(map[string]interface{}{
						"msg_type": "spawn_troop",
						"troop":    troop,
					}); err != nil {
						log.Println("readPlayerMessages: failed to relay spawn_troop to opponent:", err)
					}
				}
				if err := conn.Conn.WriteJSON(map[string]interface{}{
					"msg_type": "spawn_troop",
					"troop":    troop,
				}); err != nil {
					log.Println("readPlayerMessages: failed to echo spawn_troop to sender:", err)
				}
				WriteMU.Unlock()
			} else if (msg.Action == "" || msg.Action == "retreat") && isAttacker {
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
	tickFrequency time.Duration,
) {
	ticker := time.NewTicker(tickFrequency * time.Second)
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
				if err := attackerConn.WriteJSON(update); err != nil {
					log.Println("runSimulation: failed to send battle_update to attacker:", err)
				}
			}
			if defenderOnline {
				if err := defenderConn.WriteJSON(update); err != nil {
					log.Println("runSimulation: failed to send battle_update to defender:", err)
				}
			}
			WriteMU.Unlock()
		}
	}

}
func simulate(state *BattleState) map[string]interface{} {

	buildingDmgDone := make([]int, len(state.AliveBuildings))
	AttackertroopDmgDone := make([]int, len(state.AliveTroopAttacker))
	DefendertroopDmgDone := make([]int, len(state.AliveTroopDefender))

	for i := 0; i < len(state.AliveTroopAttacker); i++ {
		troop := &state.AliveTroopAttacker[i]
		if !state.TroopSpawns[troop.TroopIndex].SpawnedByAttacker {
			continue
		}

		var prefCat models.BuildingCategory
		hasPreferred := false
		if troop.Config.PreferredTarget != nil {
			prefCat = *troop.Config.PreferredTarget
			for _, ab := range state.AliveBuildings {
				if models.BuildingID_Category[state.Buildings[ab.BuildingIndex].Placed_Building.BuildingID] == prefCat {
					hasPreferred = true
					break
				}
			}
		}

		bestBuildingIdx := -1
		bestBuildingi := -1
		bestDefenderTroopi := -1
		minDstSq := math.MaxFloat64
		targetIsTroop := false

		for j, ab := range state.AliveBuildings {
			b := &state.Buildings[ab.BuildingIndex]
			if hasPreferred && models.BuildingID_Category[b.Placed_Building.BuildingID] != prefCat && models.BuildingID_Category[b.Placed_Building.BuildingID] != models.Wall {
				continue
			}
			dx := float64(b.Placed_Building.GridX) - troop.Current_X
			dy := float64(b.Placed_Building.GridY) - troop.Current_Y
			dstSq := dx*dx + dy*dy
			if dstSq < minDstSq {
				minDstSq = dstSq
				bestBuildingIdx = ab.BuildingIndex
				bestBuildingi = j
				targetIsTroop = false
			}
		}

		for tIdx := 0; tIdx < len(state.AliveTroopDefender); tIdx++ {
			t := &state.AliveTroopDefender[tIdx]
			dx := t.Current_X - troop.Current_X
			dy := t.Current_Y - troop.Current_Y
			dstSq := dx*dx + dy*dy
			if dstSq < minDstSq {
				minDstSq = dstSq
				bestDefenderTroopi = tIdx
				targetIsTroop = true
			}
		}

		if targetIsTroop && bestDefenderTroopi != -1 {
			target := &state.AliveTroopDefender[bestDefenderTroopi]
			dist := math.Sqrt(minDstSq)
			if dist <= troop.Config.AttackRange {
				dmg := float64(troop.LevelStat.DamagePerShot) / troop.Config.AttackSpeedSeconds
				intDmg := int(dmg)
				if intDmg <= 0 && troop.LevelStat.DamagePerShot > 0 {
					intDmg = 1
				}
				DefendertroopDmgDone[bestDefenderTroopi] += intDmg
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
		} else if bestBuildingIdx != -1 {
			targetBuilding := &state.Buildings[bestBuildingIdx]
			dist := math.Sqrt(minDstSq)
			if dist <= troop.Config.AttackRange {
				dmg := float64(troop.LevelStat.DamagePerShot) / troop.Config.AttackSpeedSeconds
				intDmg := int(dmg)
				if intDmg <= 0 && troop.LevelStat.DamagePerShot > 0 {
					intDmg = 1
				}
				buildingDmgDone[bestBuildingi] += intDmg
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

			if b.Defender.Stat.DamageType == models.Splash {
				targetTroop := &state.AliveTroopAttacker[bestTroopi]
				const splashRadiusSq = 1.5 * 1.5
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

	for tIdx, dmg := range DefendertroopDmgDone {
		state.AliveTroopDefender[tIdx].HealthRemaining -= dmg
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

	aliveDefTCount := 0
	DefenderTroopDied := make([]int, 0)
	for i, t := range state.AliveTroopDefender {
		if t.HealthRemaining > 0 {
			state.AliveTroopDefender[aliveDefTCount] = t
			aliveDefTCount++
		} else {
			DefenderTroopDied = append(DefenderTroopDied, i)
		}
	}
	state.AliveTroopDefender = state.AliveTroopDefender[:aliveDefTCount]

	return map[string]interface{}{
		"msg_type":              "battle_update",
		"building_damage":       buildingDmgDone,
		"attacker_troop_damage": AttackertroopDmgDone,
		"defender_troop_damage": DefendertroopDmgDone,
		"building_died":         BuildingDied,
		"attacker_troop_died":   AttackerTroopDied,
		"defender_troop_died":   DefenderTroopDied,
	}
}
