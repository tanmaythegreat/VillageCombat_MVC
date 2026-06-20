package controllers

import (
	"Village_combat/models"
	"errors"

	"github.com/gorilla/websocket"
)

func RepairBuilding(userId string, data struct {
	PlacedBuildingID string `json:"placed_building_id"`
	UseGems          bool   `json:"use_gems"`
}, conn *websocket.Conn) error {
	constructionUnderProgress, err := models.IsConstructionUnderProgress(userId, data.PlacedBuildingID)
	if err != nil {
		return SendError(conn)
	}
	if constructionUnderProgress {
		errPayload := []byte(`{"status": "error", "message": "Building already under construction."}`)
		return conn.WriteMessage(websocket.TextMessage, errPayload)
	}
	isBroken, err := models.IsBuildingBroken(userId, data.PlacedBuildingID)
	if err != nil {
		return SendError(conn)
	}
	if !isBroken {
		errPayload := []byte(`{"status": "error", "message": "Cannot Repair unbroken building."}`)
		return conn.WriteMessage(websocket.TextMessage, errPayload)
	}
	building, err := models.GetPlacedBuilding(userId, data.PlacedBuildingID)
	if err != nil {
		return SendError(conn)
	}
	cost, err := models.GetConstructionCost(building.BuildingID, building.Level+1)
	if err != nil {
		return SendError(conn)
	}
	cost.TimeRequiredSeconds /= 10
	cost.OrGemRequired /= 10
	cost.OrGemRequired += 1
	cost.GoldRequired /= 10
	cost.ElixirRequired /= 10
	cost.DarkElixirRequired /= 10
	tx, err := models.UserPurchase(userId, cost, data.UseGems)
	if errors.Is(err, errors.New("insufficient resources")) {
		errPayload := []byte(`{"status": "error", "message": "Insufficient Resources."}`)
		return conn.WriteMessage(websocket.TextMessage, errPayload)
	}
	if err != nil {
		return SendError(conn)
	}
	var time_req int = cost.TimeRequiredSeconds
	if data.UseGems {
		time_req = 1
	}
	task, err := models.StartConstruction_Building(userId, models.BauildingRepair, data.PlacedBuildingID, time_req, tx)
	if err != nil {
		tx.Rollback()
		return SendError(conn)
	}
	tx.Commit()
	userData, err := models.GetUserData(userId)
	if err != nil {
		return SendError(conn)
	}
	return conn.WriteJSON(map[string]interface{}{
		"msg_type":  "construction_started",
		"task":      task,
		"user_data": userData,
	})
}
