package controllers

import (
	"Village_combat/models"
	"math"
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

	leveldat := models.ResourceLevelDetails[struct {
		ID    string
		Level int
	}{ID: placedBuilding.BuildingID, Level: placedBuilding.Level}]
	generationRate := leveldat.GenerationRatePerHour
	var amount = math.Min(dt*generationRate, float64(leveldat.StorageCapacity))
	var user models.UserData
	var extra int64
	if placedBuilding.BuildingID == models.GoldMine_ID {
		user, err, extra = models.AddUserGoldGetRemaining(userId, int64(amount))
	} else if placedBuilding.BuildingID == models.ElixirCollector_ID {
		user, err, extra = models.AddUserElixirGetRemaining(userId, int64(amount))
	} else if placedBuilding.BuildingID == models.DarkElixirDrill_ID {
		user, err, extra = models.AddUserDarkElixirGetRemaining(userId, int64(amount))
	}
	if err != nil {
		return SendError(conn, err)
	}
	err = models.DecreaseUpdateTime(userId, data.PlacedBuildingId, float64(extra)/generationRate)
	if err != nil {
		return SendError(conn, err)
	}
	return conn.WriteJSON(map[string]interface{}{
		"msg_type":  "resource_collected",
		"user_data": user,
	})

}
