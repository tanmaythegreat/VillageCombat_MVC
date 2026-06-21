package controllers

import (
	"Village_combat/models"

	"github.com/gorilla/websocket"
)

func MoveBuilding(userId string, data struct {
	PlacedBuildingID string `json:"placed_building_id"`
	GridX            int    `json:"grid_x"`
	GridY            int    `json:"grid_y"`
}, conn *websocket.Conn) error {

	nearByBuildings, err := models.GetNearByBuildings(userId, data.GridX, data.GridY)
	if err != nil {
		return SendError(conn, err)
	}
	placedBuilding, err := models.GetPlacedBuilding(userId, data.PlacedBuildingID)
	if err != nil {
		return SendError(conn, err)
	}
	newSize, exists := models.BuildingSize[placedBuilding.BuildingID]
	if !exists {
		errPayload := []byte(`{"status": "error", "message": "Unknown Building."}`)
		return conn.WriteMessage(websocket.TextMessage, errPayload)
	}
	hasCollision := false
	for _, b := range nearByBuildings {
		if b.Id == data.PlacedBuildingID {
			continue
		}
		if data.GridX < b.At_x+b.Size_x && data.GridX+newSize.X > b.At_x &&
			data.GridY < b.At_y+b.Size_y && data.GridY+newSize.Y > b.At_y {
			hasCollision = true
			break
		}
	}
	if hasCollision {
		errPayload := []byte(`{"status": "error", "message": "Can't Place Here."}`)
		err := conn.WriteMessage(websocket.TextMessage, errPayload)
		if err != nil {
			return SendError(conn, err)
		}
		return conn.WriteJSON(map[string]interface{}{
			"msg_type":           "moved",
			"placed_building_id": data.PlacedBuildingID,
			"grid_x":             placedBuilding.GridX,
			"grid_y":             placedBuilding.GridY,
		})

	}
	_, err = models.UpdatePlacedBuildingPosition(userId, data.PlacedBuildingID, data.GridX, data.GridY)
	if err != nil {
		return SendError(conn, err)
	}
	return conn.WriteJSON(map[string]interface{}{
		"msg_type":           "moved",
		"placed_building_id": data.PlacedBuildingID,
		"grid_x":             data.GridX,
		"grid_y":             data.GridY,
	})
}
