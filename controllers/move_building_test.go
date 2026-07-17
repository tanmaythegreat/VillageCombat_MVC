package controllers

import (
	"testing"

	"Village_combat/models"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func moveBuildingData(id string, x, y int) struct {
	PlacedBuildingID string `json:"placed_building_id"`
	GridX            int    `json:"grid_x"`
	GridY            int    `json:"grid_y"`
} {
	return struct {
		PlacedBuildingID string `json:"placed_building_id"`
		GridX            int    `json:"grid_x"`
		GridY            int    `json:"grid_y"`
	}{PlacedBuildingID: id, GridX: x, GridY: y}
}

func TestMoveBuilding_Success(t *testing.T) {
	mock := newMockDB(t)
	resetBuildingStaticState(t)
	models.BuildingSize["cannon"] = struct {
		X int
		Y int
	}{X: 3, Y: 3}

	server, client := newTestConnPair(t)

	mock.ExpectQuery(`FROM "placed_buildings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "building_id", "grid_x", "grid_y", "level", "is_broken"}))

	mock.ExpectQuery(`FROM "placed_buildings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "building_id", "grid_x", "grid_y", "level", "is_broken"}).
			AddRow("pb-1", "user-1", "cannon", 0, 0, 3, false))

	mock.ExpectQuery(`UPDATE "placed_buildings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "grid_x", "grid_y"}).AddRow("pb-1", 5, 5))

	errCh := make(chan error, 1)
	go func() {
		errCh <- MoveBuilding("user-1", moveBuildingData("pb-1", 5, 5), server)
	}()

	var msg struct {
		MsgType          string `json:"msg_type"`
		PlacedBuildingID string `json:"placed_building_id"`
		GridX            int    `json:"grid_x"`
		GridY            int    `json:"grid_y"`
	}
	if err := client.ReadJSON(&msg); err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("MoveBuilding returned error: %v", err)
	}

	if msg.MsgType != "moved" || msg.GridX != 5 || msg.GridY != 5 {
		t.Errorf("unexpected moved payload: %+v", msg)
	}
	requireMet(t, mock)
}

func TestMoveBuilding_Collision(t *testing.T) {
	mock := newMockDB(t)
	resetBuildingStaticState(t)
	models.BuildingSize["cannon"] = struct {
		X int
		Y int
	}{X: 3, Y: 3}

	server, client := newTestConnPair(t)

	mock.ExpectQuery(`FROM "placed_buildings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "building_id", "grid_x", "grid_y", "level", "is_broken"}).
			AddRow("other-building", "user-1", "cannon", 5, 5, 3, false))

	mock.ExpectQuery(`FROM "placed_buildings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "building_id", "grid_x", "grid_y", "level", "is_broken"}).
			AddRow("pb-1", "user-1", "cannon", 0, 0, 3, false))

	errCh := make(chan error, 1)
	go func() {
		errCh <- MoveBuilding("user-1", moveBuildingData("pb-1", 5, 5), server)
	}()

	_, raw, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read first message: %v", err)
	}
	if len(raw) == 0 {
		t.Error("expected non-empty collision error payload")
	}

	var msg struct {
		MsgType string `json:"msg_type"`
		GridX   int    `json:"grid_x"`
		GridY   int    `json:"grid_y"`
	}
	if err := client.ReadJSON(&msg); err != nil {
		t.Fatalf("failed to read second message: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("MoveBuilding returned error: %v", err)
	}
	if msg.GridX != 0 || msg.GridY != 0 {
		t.Errorf("expected building to report its original position on collision, got (%d,%d)", msg.GridX, msg.GridY)
	}
	requireMet(t, mock)
}
