package controllers

import (
	"Village_combat/models"
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

func AllBuildingDataLoad(userId string, conn *websocket.Conn) error {
	buildings, err := models.GetAllBuildingConfigsJSON()
	if err != nil {
		return SendError(conn)
	}
	troops, err := models.GetAllTroopsDataJSON()
	if err != nil {
		return SendError(conn)
	}
	defence, err := models.GetAllDefenceBuildingConfigsJSON()
	if err != nil {
		return SendError(conn)
	}
	army, err := models.GetAllArmyBuildingConfigsJSON()
	if err != nil {
		return SendError(conn)
	}
	resource, err := models.GetAllResourceBuildingConfigsJSON()
	if err != nil {
		return SendError(conn)
	}
	id_level, err := models.GetPlacedBuilding_ID_Level(userId)

	configMap := make(map[string]json.RawMessage)
	if err != nil {
		return SendError(conn)
	}
	for _, il := range id_level {
		key := fmt.Sprintf("%s:%d", il.BuildingID, il.Level)
		if _, exists := configMap[key]; exists {
			continue
		}
		jsonData, err := models.GetBuildingDataOfLevelJSON(il.BuildingID, il.Level)
		if err != nil {
			return SendError(conn)
		}
		configMap[key] = jsonData
	}
	for id, _ := range models.BuildingSize {
		key := fmt.Sprintf("%s:%d", id, 1)
		if _, exists := configMap[key]; exists {
			continue
		}
		jsonData, err := models.GetBuildingDataOfLevelJSON(id, 1)
		if err != nil {
			return SendError(conn)
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
