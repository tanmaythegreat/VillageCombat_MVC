package resources

import (
	"Village_combat/models"
	"Village_combat/services"
	"errors"
	"math"
	"time"

	"github.com/gorilla/websocket"
)

func CollectResource(userId string, data struct {
	PlacedBuildingId string `json:"placed_building_id"`
}, conn *websocket.Conn) error {
	isBroken, err := models.IsBuildingBroken(userId, data.PlacedBuildingId)
	if err != nil {
		return services.SendError(conn, err)
	}
	if isBroken {
		errPayload := []byte(`{"status": "error", "message": "Cannot collect from broken building."}`)
		return conn.WriteMessage(websocket.TextMessage, errPayload)
	}
	placedBuilding, err := models.GetPlacedBuilding(userId, data.PlacedBuildingId)
	if err != nil {
		return services.SendError(conn, err)
	}
	now := time.Now()
	var dt = now.Sub(placedBuilding.LastUpdatedAt).Hours()

	leveldat, ok := models.ResourceLevelDetails[struct {
		ID    string
		Level int
	}{ID: placedBuilding.BuildingID, Level: placedBuilding.Level}]
	if !ok {
		return services.SendError(conn, errors.New("unknown resource"))
	}
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
		return services.SendError(conn, err)
	}
	if generationRate != 0 {
		err = models.DecreaseUpdateTime(userId, data.PlacedBuildingId, float64(extra)/generationRate, now)
		if err != nil {
			return services.SendError(conn, err)
		}
	} else {
		// ideally this should never happen
		err = models.DecreaseUpdateTime(userId, data.PlacedBuildingId, 0, now)
		if err != nil {
			return services.SendError(conn, err)
		}
	}
	return conn.WriteJSON(map[string]any{
		"msg_type":  "resource_collected",
		"user_data": user,
	})

}
