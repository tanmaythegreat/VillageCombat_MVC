package buildings

import (
	"Village_combat/models"
	"Village_combat/services"
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

func AllBuildingDataLoad(userId string, conn *websocket.Conn) error {
	buildings, err := models.GetAllBuildingConfigsJSON()
	if err != nil {
		return services.SendError(conn, err)
	}
	troops, err := models.GetAllTroopsDataJSON()
	if err != nil {
		return services.SendError(conn, err)
	}
	defence, err := models.GetAllDefenceBuildingConfigsJSON()
	if err != nil {
		return services.SendError(conn, err)
	}
	army, err := models.GetAllArmyBuildingConfigsJSON()
	if err != nil {
		return services.SendError(conn, err)
	}
	resource, err := models.GetAllResourceBuildingConfigsJSON()
	if err != nil {
		return services.SendError(conn, err)
	}
	id_level, err := models.GetPlacedBuilding_ID_Level(userId)

	configMap := make(map[string]json.RawMessage)
	if err != nil {
		return services.SendError(conn, err)
	}
	for _, il := range id_level {
		key := fmt.Sprintf("%s:%d", il.BuildingID, il.Level)
		if _, exists := configMap[key]; exists {
			continue
		}
		jsonData, err := models.GetBuildingDataOfLevelJSON(il.BuildingID, il.Level)
		if err != nil {
			return services.SendError(conn, err)
		}
		configMap[key] = jsonData
	}
	for id, _ := range models.BuildingSize {
		key := fmt.Sprintf("%s:%d", id, 0)
		if _, exists := configMap[key]; exists {
			continue
		}
		jsonData, err := models.GetBuildingDataOfLevelJSON(id, 0)
		if err != nil {
			return services.SendError(conn, err)
		}
		configMap[key] = jsonData
	}
	data := struct {
		MsgType             string                     `json:"msg_type"`
		Building            json.RawMessage            `json:"building"`
		Troops              json.RawMessage            `json:"troops"`
		Defence             json.RawMessage            `json:"defence"`
		Army                json.RawMessage            `json:"army"`
		Resource            json.RawMessage            `json:"resource"`
		ParticularLevelData map[string]json.RawMessage `json:"particular_level_data"`
	}{
		MsgType:             "building_troop",
		Building:            buildings,
		Troops:              troops,
		Defence:             defence,
		Army:                army,
		Resource:            resource,
		ParticularLevelData: configMap,
	}
	return conn.WriteJSON(data)
}
