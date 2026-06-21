package controllers

import (
	"Village_combat/models"
	"time"

	"github.com/gorilla/websocket"
)

func CollectResource(userId string, data struct {
	PlacedBuildingId string `json:"placed_building_id"`
}, conn *websocket.Conn) error {
	isBroken, err := models.IsBuildingBroken(userId, data.PlacedBuildingId)
	if err != nil {
		return SendError(conn, err)
	}
	if isBroken {
		errPayload := []byte(`{"status": "error", "message": "Cannot collect from broken building."}`)
		return conn.WriteMessage(websocket.TextMessage, errPayload)

	}
	placedBuilding, err := models.UpdatePlacedBuilding(userId, data.PlacedBuildingId)
	if err != nil {
		return SendError(conn, err)
	}
	var dt = time.Now().Sub(placedBuilding.LastUpdatedAt).Hours()
	generationRate, err := models.GetGenerationRate(placedBuilding.BuildingID, placedBuilding.Level)
	if err != nil {
		return SendError(conn, err)
	}
	var amount = dt * generationRate
	var user models.UserData
	if placedBuilding.BuildingID == models.GoldMine_ID {
		user, err = models.AddUserGold(userId, int(amount))
	} else if placedBuilding.BuildingID == models.ElixirCollector_ID {
		user, err = models.AddUserElixir(userId, int(amount))
	} else if placedBuilding.BuildingID == models.DarkElixirDrill_ID {
		user, err = models.AddUserDarkElixir(userId, int(amount))
	}
	return conn.WriteJSON(map[string]interface{}{
		"msg_type":  "resource_collected",
		"user_data": user,
	})

}
