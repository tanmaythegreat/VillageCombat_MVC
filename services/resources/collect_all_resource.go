package resources

import (
	"Village_combat/models"
	"Village_combat/services"
	"math"
	"time"

	"github.com/gorilla/websocket"
)

func CollectAllResource(userId string, conn *websocket.Conn) error {
	placedBuildings, err := models.GetPlacedBuildings(userId)
	if err != nil {
		return services.SendError(conn, err)
	}

	now := time.Now()
	ids := make([]string, 0, len(placedBuildings))
	hrs := make([]float64, 0, len(placedBuildings))

	for _, building := range placedBuildings {
		if building.IsBroken {
			continue
		}
		if models.BuildingID_Category[building.BuildingID] == models.Resource {
			levelDet := models.ResourceLevelDetails[struct {
				ID    string
				Level int
			}{ID: building.BuildingID, Level: building.Level}]
			generationRate := levelDet.GenerationRatePerHour
			if generationRate <= 0 {
				continue // avoid divide-by-zero below; skip misconfigured level data
			}
			toCollect := math.Min(generationRate*now.Sub(building.LastUpdatedAt).Hours(), float64(levelDet.StorageCapacity))

			var extra int64
			var addErr error
			switch building.BuildingID {
			case models.ElixirCollector_ID:
				_, addErr, extra = models.AddUserElixirGetRemaining(userId, int64(toCollect))
			case models.GoldMine_ID:
				_, addErr, extra = models.AddUserGoldGetRemaining(userId, int64(toCollect))
			case models.DarkElixirDrill_ID:
				_, addErr, extra = models.AddUserDarkElixirGetRemaining(userId, int64(toCollect))
			}
			if addErr != nil {
				return services.SendError(conn, addErr)
			}

			ids = append(ids, building.ID)
			hrs = append(hrs, float64(extra)/generationRate)
		}
	}

	err = models.DecreaseUpdateTimeALL(userId, ids, hrs, now)
	if err != nil {
		return services.SendError(conn, err)
	}

	newData, err := models.GetPlacedBuildingJSON(userId)
	if err != nil {
		return services.SendError(conn, err)
	}
	userData, err := models.GetUserData(userId)
	if err != nil {
		return services.SendError(conn, err)
	}

	return conn.WriteJSON(map[string]interface{}{
		"msg_type":         "resource_collected",
		"placed_buildings": newData,
		"user_data":        userData,
	})

}
