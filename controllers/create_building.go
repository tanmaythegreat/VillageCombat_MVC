package controllers

import (
	"Village_combat/models"
	"errors"

	"github.com/gorilla/websocket"
)

func CreateBuilding(userId string, data struct {
	BuildingID string `json:"building_id"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
	UseGems    bool   `json:"use_gems"`
}, conn *websocket.Conn) error {
	nearByBuildings, err := models.GetNearByBuildings(userId, data.X, data.Y)
	if err != nil {
		return SendError(conn, err)
	}
	newSize, exists := models.BuildingSize[data.BuildingID]
	if !exists {
		errPayload := []byte(`{"status": "error", "message": "Unknown Building."}`)
		return conn.WriteMessage(websocket.TextMessage, errPayload)
	}
	hasCollision := false
	for _, b := range nearByBuildings {
		if data.X < b.At_x+b.Size_x && data.X+newSize.X > b.At_x &&
			data.Y < b.At_y+b.Size_y && data.Y+newSize.Y > b.At_y {
			hasCollision = true
			break
		}
	}
	if hasCollision {
		errPayload := []byte(`{"status": "error", "message": "Can't Place Here."}`)
		return conn.WriteMessage(websocket.TextMessage, errPayload)
	}
	cost, err := models.GetConstructionCost(data.BuildingID, 1)
	if err != nil {
		return SendError(conn, err)
	}
	userData, err := models.GetUserData(userId)
	if err != nil {
		return SendError(conn, err)
	}
	if userData.TownHallLevel < cost.TownHallLevelRequired {
		errPayload := []byte(`{"status": "error", "message": "Town Hall Level_from not sufficient."}`)
		return conn.WriteMessage(websocket.TextMessage, errPayload)
	}
	tx, err := models.UserPurchase(userId, cost, data.UseGems)
	if errors.Is(err, errors.New("insufficient resources")) {
		errPayload := []byte(`{"status": "error", "message": "Insufficient Resources."}`)
		return conn.WriteMessage(websocket.TextMessage, errPayload)
	}
	if err != nil {
		return SendError(conn, err)
	}
	var timeReq int = cost.TimeRequiredSeconds
	if data.UseGems {
		timeReq = 1
	}
	placedBuilding, task, err := models.ConstructBuilding(userId, data.BuildingID, data.X, data.Y, tx, timeReq)
	if err != nil {
		tx.Rollback()
		return SendError(conn, err)
	}
	tx.Commit()
	userData, err = models.GetUserData(userId)
	if err != nil {
		return SendError(conn, err)
	}
	return conn.WriteJSON(map[string]interface{}{
		"msg_type":        "construction_started",
		"placed_building": placedBuilding,
		"task":            task,
		"user_data":       userData,
	})
}
