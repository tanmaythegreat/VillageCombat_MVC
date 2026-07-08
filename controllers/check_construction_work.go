package controllers

import (
	"Village_combat/models"
	"encoding/json"

	"github.com/gorilla/websocket"
)

func CheckConstructionWork(userId string, conn *websocket.Conn) error {
	constructionTasks, buildings_updated, err := models.CheckIsConstructionComplete(userId)
	if err != nil {
		return SendError(conn, err)
	}
	if len(constructionTasks) > 0 {
		var goldCapacityIncrement int64 = 0
		var darkElixirCapacityIncrement int64 = 0
		var elixirCapacityIncrement int64 = 0
		var troopCapacityIncrement int = 0
		var levelDetails = make([]json.RawMessage, 0, len(buildings_updated))
		for _, building := range buildings_updated {
			if models.BuildingID_Category[building.BuildingID] == models.TownHall {
				err = models.IncrementUserTownHallLevel(userId)
				if err != nil {
					return SendError(conn, err)
				}
			} else if models.BuildingID_Category[building.BuildingID] == models.Resource {
				if building.BuildingID == models.ElixirStorage_ID {
					difference, err := models.GetCapacityDifference(building.BuildingID, building.Level-1, building.Level)
					if err != nil {
						return SendError(conn, err)
					}
					elixirCapacityIncrement += difference
				} else if building.BuildingID == models.GoldStorage_ID {
					difference, err := models.GetCapacityDifference(building.BuildingID, building.Level-1, building.Level)
					if err != nil {
						return SendError(conn, err)
					}
					goldCapacityIncrement += difference
				} else if building.BuildingID == models.DarkElixirStorage_ID {
					difference, err := models.GetCapacityDifference(building.BuildingID, building.Level-1, building.Level)
					if err != nil {
						return SendError(conn, err)
					}
					darkElixirCapacityIncrement += difference
				}
			} else if models.BuildingID_Category[building.BuildingID] == models.Army {
				difference, err := models.GetTroopCapacityDifference(building.Level-1, building.Level)
				if err != nil {
					return SendError(conn, err)
				}
				troopCapacityIncrement += difference
			}
			levelJSON, err := models.GetBuildingDataOfLevelJSON(building.BuildingID, building.Level)
			if err != nil {
				return SendError(conn, err)
			}
			levelDetails = append(levelDetails, levelJSON)
		}
		userData, err := models.AddUserCapacity(userId, goldCapacityIncrement, elixirCapacityIncrement, darkElixirCapacityIncrement, troopCapacityIncrement)
		if err != nil {
			return SendError(conn, err)
		}
		return conn.WriteJSON(map[string]interface{}{
			"msg_type":                "construction_completed",
			"particular_level_detail": levelDetails,
			"construction_done":       constructionTasks,
			"user_data":               userData,
		})
	}
	return nil
}
