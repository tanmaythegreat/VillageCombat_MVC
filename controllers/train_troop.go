package controllers

import (
	"Village_combat/models"

	"github.com/gorilla/websocket"
)

func TrainTroop(userId string, data struct {
	BarrackPlacedBuildingID string `json:"barrack_placed_building_id"`
	TroopId                 string `json:"troop_id"`
	LevelFrom               int    `json:"level_from"`
	Count                   int    `json:"count"`
	UseGems                 bool   `json:"use_gems"`
}, conn *websocket.Conn) error {
	underProgress, err := models.IsConstructionUnderProgress(userId, data.BarrackPlacedBuildingID)
	if err != nil {
		return SendError(conn)
	}
	if underProgress {
		errPayload := []byte(`{"status": "error", "message": "Already Building Construction work or Training going on here."}`)
		return conn.WriteMessage(websocket.TextMessage, errPayload)
	}
	building, err := models.GetPlacedBuilding(userId, data.BarrackPlacedBuildingID)
	if err != nil {
		return SendError(conn)
	}
	if building.BuildingID != models.Barracks_ID {
		errPayload := []byte(`{"status": "error", "message": "Can only Train in Barracks."}`)
		return conn.WriteMessage(websocket.TextMessage, errPayload)
	}
	isBroken, err := models.IsBuildingBroken(userId, building.ID)
	if err != nil {
		return SendError(conn)
	}
	if isBroken {
		errPayload := []byte(`{"status": "error", "message": "Cannot train in broken barracks."}`)
		return conn.WriteMessage(websocket.TextMessage, errPayload)
	}
	upgradeCost, err := models.GetTroopUpgradeCost(data.TroopId, data.LevelFrom+1)
	if err != nil {
		return SendError(conn)
	}
	upgradeCost.TimeRequiredSeconds *= data.Count
	upgradeCost.DarkElixirRequired *= data.Count
	upgradeCost.ElixirRequired *= data.Count
	upgradeCost.GoldRequired *= data.Count
	upgradeCost.OrGemRequired *= data.Count
	tx, err := models.UserPurchase(userId, upgradeCost, data.UseGems)
	if err != nil {
		return SendError(conn)
	}
	if data.LevelFrom != 0 {
		success, err := models.SubtractTroopsOfUser(userId, data.TroopId, data.LevelFrom, data.Count, tx)
		if err != nil {
			tx.Rollback()
			return SendError(conn)
		}
		if !success {
			tx.Rollback()
			errPayload := []byte(`{"status": "error", "message": "Not enough Troops."}`)
			return conn.WriteMessage(websocket.TextMessage, errPayload)
		}
	}
	var time_req = upgradeCost.TimeRequiredSeconds
	if data.UseGems {
		time_req = 1
	}
	constructionTask, err := models.StartTrainingTroops(userId, data.TroopId, data.Count, data.BarrackPlacedBuildingID, data.LevelFrom+1, time_req, tx)
	if err != nil {
		tx.Rollback()
		return SendError(conn)
	}

	if tx.Commit().Error != nil {
		errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
		return conn.WriteMessage(websocket.TextMessage, errPayload)
	}
	return conn.WriteJSON(map[string]interface{}{
		"msg_type": "troop_training_started",
		"task":     constructionTask,
	})
}
