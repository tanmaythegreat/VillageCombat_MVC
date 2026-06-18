package controllers

import (
	"Village_combat/models"
	"encoding/json"

	"github.com/gorilla/websocket"
)

func BuildingLevelDetails(userId string, placedBuildingId string, conn *websocket.Conn) error {
	level, err := models.GetPlacedBuildingLevel(userId, placedBuildingId)
	if err != nil {
		return SendError(conn)
	}
	dataOfLevel, err := models.GetBuildingDataOfLevelJSON(placedBuildingId, level)
	if err != nil {
		return SendError(conn)
	}
	return conn.WriteJSON(struct {
		MsgType     string          `json:"msg_type"`
		DataOfLevel json.RawMessage `json:"data_of_level"`
	}{
		MsgType:     "building_level_detail",
		DataOfLevel: dataOfLevel,
	})

}
