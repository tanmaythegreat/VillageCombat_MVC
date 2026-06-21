package controllers

import (
	"Village_combat/models"
	"encoding/json"

	"github.com/gorilla/websocket"
)

func InitialLoad(userId string, conn *websocket.Conn) error {

	troops, err := models.GetUserTrainedTroops(userId)
	if err != nil {
		return SendError(conn, err)
	}

	building, err := models.GetPlacedBuildingJSON(userId)
	if err != nil {
		return SendError(conn, err)
	}

	constructionTasks, err := models.GetConstructionTasks(userId)
	if err != nil {
		return SendError(conn, err)
	}

	userData, err := models.GetUserData(userId)
	if err != nil {
		return SendError(conn, err)
	}

	user, err := models.GetUser(userId)
	data := struct {
		MsgType           string          `json:"msg_type"`
		Building          json.RawMessage `json:"building"`
		Troops            json.RawMessage `json:"troops"`
		ConstructionTasks json.RawMessage `json:"construction_tasks"`
		UserData          models.UserData `json:"user_data"`
		User              models.User     `json:"user"`
	}{
		MsgType:           "building_troop_of_user",
		Building:          building,
		Troops:            troops,
		ConstructionTasks: constructionTasks,
		UserData:          userData,
		User:              user,
	}

	return conn.WriteJSON(data)
}
