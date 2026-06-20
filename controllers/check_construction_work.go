package controllers

import (
	"Village_combat/models"
	"encoding/json"

	"github.com/gorilla/websocket"
)

func CheckConstructionWork(userId string, conn *websocket.Conn) error {
	constructionTasks, buildings_updated, err := models.CheckIsConstructionComplete(userId)
	if err != nil {
		return SendError(conn)
	}
	if len(constructionTasks) > 0 {
		var goldCapacityIncrement = 0
		var darkElixirCapacityIncrement = 0
		var elixirCapacityIncrement = 0
		var levelDetails = make([]json.RawMessage, 0, len(buildings_updated))
		for _, building := range buildings_updated {
			if models.BuildingID_Category[building.BuildingID] == models.TownHall {
				err = models.IncrementUserTownHallLevel(userId)
				if err != nil {
					return SendError(conn)
				}
			} else if models.BuildingID_Category[building.BuildingID] == models.Resource {
				if building.BuildingID == models.ElixirStorage_ID {
					difference, err := models.GetCapacityDifference(building.BuildingID, building.Level-1, building.Level)
					if err != nil {
						return SendError(conn)
					}
					elixirCapacityIncrement += difference
				} else if building.BuildingID == models.GoldStorage_ID {
					difference, err := models.GetCapacityDifference(building.BuildingID, building.Level-1, building.Level)
					if err != nil {
						return SendError(conn)
					}
					goldCapacityIncrement += difference
				} else if building.BuildingID == models.DarkElixirStorage_ID {
					difference, err := models.GetCapacityDifference(building.BuildingID, building.Level-1, building.Level)
					if err != nil {
						return SendError(conn)
					}
					darkElixirCapacityIncrement += difference
				}
			}
			levelJSON, err := models.GetBuildingDataOfLevelJSON(building.BuildingID, building.Level)
			if err != nil {
				return SendError(conn)
			}
			levelDetails = append(levelDetails, levelJSON)
		}
		userData, err := models.AddUserCapacity(userId, goldCapacityIncrement, elixirCapacityIncrement, darkElixirCapacityIncrement)
		if err != nil {
			return SendError(conn)
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
