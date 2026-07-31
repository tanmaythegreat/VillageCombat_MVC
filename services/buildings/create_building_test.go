package buildings

import (
	"Village_combat/services"
	"testing"

	"Village_combat/models"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func createBuildingData(buildingID string, x, y int, useGems bool) struct {
	BuildingID string `json:"building_id"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
	UseGems    bool   `json:"use_gems"`
} {
	return struct {
		BuildingID string `json:"building_id"`
		X          int    `json:"x"`
		Y          int    `json:"y"`
		UseGems    bool   `json:"use_gems"`
	}{BuildingID: buildingID, X: x, Y: y, UseGems: useGems}
}

func TestCreateBuilding_UnknownBuilding(t *testing.T) {
	mock := services.NewMockDB(t)
	services.ResetBuildingStaticState(t)

	server, client := services.NewTestConnPair(t)

	mock.ExpectQuery(`FROM "placed_buildings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "building_id", "grid_x", "grid_y", "level", "is_broken"}))
	mock.ExpectQuery(`FROM "user_data"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "town_hall_level"}).AddRow("user-1", 3))

	errCh := make(chan error, 1)
	go func() {
		errCh <- CreateBuilding("user-1", createBuildingData("does-not-exist", 1, 1, false), server)
	}()

	var userDataMsg map[string]interface{}
	if err := client.ReadJSON(&userDataMsg); err != nil {
		t.Fatalf("failed to read first message: %v", err)
	}
	_, raw, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read second message: %v", err)
	}
	if len(raw) == 0 {
		t.Error("expected non-empty 'Unknown Building' error payload")
	}
	if err := <-errCh; err != nil {
		t.Fatalf("CreateBuilding returned error: %v", err)
	}
	services.RequireMet(t, mock)
}

func TestCreateBuilding_Collision(t *testing.T) {
	mock := services.NewMockDB(t)
	services.ResetBuildingStaticState(t)
	models.BuildingSize["wall-1"] = struct {
		X int
		Y int
	}{X: 1, Y: 1}

	server, client := services.NewTestConnPair(t)

	mock.ExpectQuery(`FROM "placed_buildings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "building_id", "grid_x", "grid_y", "level", "is_broken"}).
			AddRow("existing", "user-1", "wall-1", 5, 5, 1, false))
	mock.ExpectQuery(`FROM "user_data"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "town_hall_level"}).AddRow("user-1", 3))

	errCh := make(chan error, 1)
	go func() {
		errCh <- CreateBuilding("user-1", createBuildingData("wall-1", 5, 5, false), server)
	}()

	var userDataMsg map[string]interface{}
	if err := client.ReadJSON(&userDataMsg); err != nil {
		t.Fatalf("failed to read first message: %v", err)
	}
	_, raw, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read second message: %v", err)
	}
	if len(raw) == 0 {
		t.Error("expected non-empty 'Can't Place Here' error payload")
	}
	if err := <-errCh; err != nil {
		t.Fatalf("CreateBuilding returned error: %v", err)
	}
	services.RequireMet(t, mock)
}

func TestCreateBuilding_InsufficientTownHallLevel(t *testing.T) {
	mock := services.NewMockDB(t)
	services.ResetBuildingStaticState(t)
	models.BuildingSize["wall-1"] = struct {
		X int
		Y int
	}{X: 1, Y: 1}

	server, client := services.NewTestConnPair(t)

	mock.ExpectQuery(`FROM "placed_buildings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "building_id", "grid_x", "grid_y", "level", "is_broken"}))
	mock.ExpectQuery(`FROM "user_data"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "town_hall_level"}).AddRow("user-1", 1))
	mock.ExpectQuery(`FROM "upgrade_costs"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "building_id", "upgrade_to_level", "town_hall_level_required"}).
			AddRow("cost-1", "wall-1", 1, 5))

	errCh := make(chan error, 1)
	go func() {
		errCh <- CreateBuilding("user-1", createBuildingData("wall-1", 1, 1, false), server)
	}()

	var userDataMsg map[string]interface{}
	if err := client.ReadJSON(&userDataMsg); err != nil {
		t.Fatalf("failed to read first message: %v", err)
	}
	_, raw, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read second message: %v", err)
	}
	if len(raw) == 0 {
		t.Error("expected non-empty 'Town Hall Level' error payload")
	}
	if err := <-errCh; err != nil {
		t.Fatalf("CreateBuilding returned error: %v", err)
	}
	services.RequireMet(t, mock)
}

func TestCreateBuilding_Success(t *testing.T) {
	mock := services.NewMockDB(t)
	services.ResetBuildingStaticState(t)
	models.BuildingSize["wall-1"] = struct {
		X int
		Y int
	}{X: 1, Y: 1}

	server, client := services.NewTestConnPair(t)

	mock.ExpectQuery(`FROM "placed_buildings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "building_id", "grid_x", "grid_y", "level", "is_broken"}))
	mock.ExpectQuery(`FROM "user_data"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "town_hall_level"}).AddRow("user-1", 5))
	mock.ExpectQuery(`FROM "upgrade_costs"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "building_id", "upgrade_to_level", "town_hall_level_required", "gold_required", "time_required_seconds"}).
			AddRow("cost-1", "wall-1", 1, 1, 100, 10))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "user_data"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`INSERT INTO "placed_buildings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("pb-new"))
	mock.ExpectQuery(`INSERT INTO "construction_tasks"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("task-new"))
	mock.ExpectCommit()

	mock.ExpectQuery(`FROM "user_data"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "town_hall_level"}).AddRow("user-1", 5))

	errCh := make(chan error, 1)
	go func() {
		errCh <- CreateBuilding("user-1", createBuildingData("wall-1", 1, 1, false), server)
	}()

	var msg struct {
		MsgType string `json:"msg_type"`
	}
	if err := client.ReadJSON(&msg); err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("CreateBuilding returned error: %v", err)
	}
	if msg.MsgType != "construction_started" {
		t.Errorf("expected msg_type construction_started, got %q", msg.MsgType)
	}
	services.RequireMet(t, mock)
}
