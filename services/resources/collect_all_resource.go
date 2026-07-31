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
	for _, building := range placedBuildings {
		if building.IsBroken {
			continue
		}
		if models.BuildingID_Category[building.BuildingID] == models.Resource {
			var extra int64
			levelDet := models.ResourceLevelDetails[struct {
				ID    string
				Level int
			}{ID: building.BuildingID, Level: building.Level}]
			generationRate := levelDet.GenerationRatePerHour
			toCollect := math.Min(generationRate*now.Sub(building.LastUpdatedAt).Hours(), float64(levelDet.StorageCapacity))
			_, err = models.UpdatePlacedBuilding(userId, building.ID)
			if err != nil {
				return services.SendError(conn, err)
			}
			if building.BuildingID == models.ElixirCollector_ID {
				_, err, extra = models.AddUserElixirGetRemaining(userId, int64(toCollect))
				if err != nil {
					return services.SendError(conn, err)
				}
			} else if building.BuildingID == models.GoldMine_ID {
				_, err, extra = models.AddUserGoldGetRemaining(userId, int64(toCollect))
				if err != nil {
					return services.SendError(conn, err)
				}
			} else if building.BuildingID == models.DarkElixirDrill_ID {
				_, err, extra = models.AddUserDarkElixirGetRemaining(userId, int64(toCollect))
				if err != nil {
					return services.SendError(conn, err)
				}
			}
			if extra > 0 {
				err = models.DecreaseUpdateTime(userId, building.ID, float64(extra)/generationRate)
				if err != nil {
					return services.SendError(conn, err)
				}
			}
		}
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
