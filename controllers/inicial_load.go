package controllers

import (
	"Village_combat/models"
	"encoding/json"

	"github.com/gorilla/websocket"
)

func InitialLoad(userId string, conn *websocket.Conn) error {

	troops, err := models.GetUserTrainedTroops(userId)
	if err != nil {
		return SendError(conn)
	}

	building, err := models.GetPlacedBuildingJSON(userId)
	if err != nil {
		return SendError(conn)
	}

	constructionTasks, err := models.GetConstructionTasks(userId)
	if err != nil {
		return SendError(conn)
	}

	userData, err := models.GetUserData(userId)
	if err != nil {
		return SendError(conn)
	}

	data := struct {
		MsgType           string          `json:"msg_type"`
		Building          json.RawMessage `json:"building"`
		Troops            json.RawMessage `json:"troops"`
		ConstructionTasks json.RawMessage `json:"construction_tasks"`
		UserData          models.UserData `json:"user_data"`
	}{
		MsgType:           "building_troop_of_user",
		Building:          building,
		Troops:            troops,
		ConstructionTasks: constructionTasks,
		UserData:          userData,
	}

	return conn.WriteJSON(data)
}
