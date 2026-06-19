package battle

import (
	"Village_combat/models"
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

func Replay(battleId string, conn *websocket.Conn) error {
	record, err := models.GetBattleRecord(battleId)
	if err != nil {
		errPayload := map[string]string{
			"status":  "error",
			"message": "Internal Server Error",
		}
		return conn.WriteJSON(errPayload)
	}
	state := &BattleState{
		mu:          sync.Mutex{},
		TroopSpawns: record.TroopSpawns,
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

	ToSend := make([]models.PlacedBuilding, len(record.InitialBuildings))

	for i, building := range record.InitialBuildings {
		ToSend[i].Level = building.Level
		ToSend[i].GridX = building.Grid_X
		ToSend[i].GridY = building.Grid_Y
		ToSend[i].BuildingID = building.BuildingID
		health, err := models.GetBuildingHealth(building.BuildingID, building.Level)
		if err != nil {
			// TODO : its game over
		}
		if models.BuildingID_Category[building.BuildingID] == models.Defense {
			levelStat, stat, err := models.GetDefenceBuildingStatAndLevelStat(building.BuildingID, building.Level)
			if err != nil {
				// TODO : end the battle, its all over :(
			}
			state.Buildings = append(state.Buildings, struct {
				Placed_Building models.PlacedBuilding
				Defender        *struct {
					LevelStat models.DefenseBuildingLevelStats
					Stat      models.DefenseBuildingStats
				}
			}{Placed_Building: ToSend[i], Defender: &struct {
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
			}{Placed_Building: ToSend[i], Defender: nil})
		}
		if !building.IsBroken {
			state.AliveBuildings = append(state.AliveBuildings, struct {
				BuildingIndex   int `json:"BuildingIndex"`
				HealthRemaining int `json:"HealthRemaining"`
			}{BuildingIndex: len(ToSend) - 1, HealthRemaining: health})
		}
	}

	conn.WriteJSON(map[string]interface{}{
		"msg_type":          "replay",
		"defender_building": ToSend,
		"alive_buildings":   state.AliveBuildings,
	})

	battleHistory, err := models.GetBattleHistory(battleId)
	if err != nil {
		// TODO : ..
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(battleHistory.BattleDuration)*time.Second)
	defer cancel()
	WriteMU := sync.Mutex{} // this is write lock for connections only for these 3 functions
	go SpawnTroopsOverTime(record.TroopSpawns, state, &WriteMU, conn)
	runSimulation(ctx, state, conn, nil, true, false, &WriteMU, 1)

	DeathMap := make(map[string]int)

	for _, building := range battleHistory.BrokenBuildings {
		DeathMap[building.BuildingID] = building.Count
	}
	TroopLossAttacker := make(map[string]int)
	TroopLossDefender := make(map[string]int)
	for _, troopSpawn := range battleHistory.TroopLosses {
		if troopSpawn.IsAttacker {
			TroopLossAttacker[troopSpawn.TroopID] = troopSpawn.LossCount
		} else {
			TroopLossDefender[troopSpawn.TroopID] = troopSpawn.LossCount
		}
	}

	defenderName, err := models.GetUsername(battleHistory.DefenderID)
	if err != nil {
		// TODO : what to do
	}
	attackerName, err := models.GetUsername(battleHistory.AttackerID)
	if err != nil {
		// TODO : what to do
	}
	conn.WriteJSON(map[string]interface{}{
		"msg_type":            "battle_over",
		"battle_id":           battleId,
		"battle_outcome":      battleHistory,
		"attacker_troop_loss": TroopLossAttacker,
		"buildings_broken":    DeathMap,
		"opponent_username":   defenderName,
		"my_username":         attackerName,
	})
	return nil
}

func SpawnTroopsOverTime(troops []models.TroopSpawn, state *BattleState, WriteMU *sync.Mutex, conn *websocket.Conn) {
	startTime := time.Now()

	for _, troop := range troops {

		delay := time.Duration(troop.SpawnTime * float64(time.Second))
		targetTime := startTime.Add(delay)
		time.Sleep(time.Until(targetTime))
		lvlStat := models.TroopLevelDetails[struct {
			ID    string
			Level int
		}{ID: troop.TroopID, Level: troop.TroopLevel}]

		var aliveTrop = struct {
			TroopIndex      int
			Current_X       float64
			Current_Y       float64
			Config          models.TroopConfig
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
		if troop.SpawnedByAttacker {
			state.AliveTroopAttacker = append(state.AliveTroopAttacker, aliveTrop)
		} else {
			state.AliveTroopDefender = append(state.AliveTroopDefender, aliveTrop)
		}
		state.mu.Unlock()
		WriteMU.Lock()
		conn.WriteJSON(map[string]interface{}{
			"msg_type": "spawn_troop",
			"troop":    troop,
		})
		WriteMU.Unlock()
	}
}
