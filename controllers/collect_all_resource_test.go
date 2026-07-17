package controllers

import (
	"testing"
	"time"

	"Village_combat/models"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestCollectAllResource_NoBuildings(t *testing.T) {
	mock := newMockDB(t)

	server, client := newTestConnPair(t)

	mock.ExpectQuery(`FROM "placed_buildings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "building_id", "grid_x", "grid_y", "level", "is_broken"}))
	mock.ExpectQuery(`FROM "placed_buildings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "building_id", "grid_x", "grid_y", "level", "is_broken"}))
	mock.ExpectQuery(`FROM "user_data"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("user-1"))

	errCh := make(chan error, 1)
	go func() {
		errCh <- CollectAllResource("user-1", server)
	}()

	var msg struct {
		MsgType string `json:"msg_type"`
	}
	if err := client.ReadJSON(&msg); err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("CollectAllResource returned error: %v", err)
	}
	if msg.MsgType != "resource_collected" {
		t.Errorf("expected msg_type resource_collected, got %q", msg.MsgType)
	}
	requireMet(t, mock)
}

func TestCollectAllResource_OneGoldMine(t *testing.T) {
	mock := newMockDB(t)
	resetBuildingStaticState(t)
	models.GoldMine_ID = "gold-mine"
	models.BuildingID_Category["gold-mine"] = models.Resource
	models.ResourceLevelDetails = map[struct {
		ID    string
		Level int
	}]models.ResourceBuildingLevelStats{
		{ID: "gold-mine", Level: 1}: {
			BuildingID:            "gold-mine",
			Level:                 1,
			GenerationRatePerHour: 100,
			StorageCapacity:       10000,
		},
	}

	server, client := newTestConnPair(t)

	// GetPlacedBuildings
	mock.ExpectQuery(`FROM "placed_buildings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "building_id", "grid_x", "grid_y", "level", "is_broken", "last_updated_at"}).
			AddRow("pb-1", "user-1", "gold-mine", 0, 0, 1, false, time.Now().Add(-time.Hour)))

	// UpdatePlacedBuilding: select then update
	mock.ExpectQuery(`FROM "placed_buildings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "building_id", "level", "last_updated_at"}).
			AddRow("pb-1", "user-1", "gold-mine", 1, time.Now().Add(-time.Hour)))
	mock.ExpectExec(`UPDATE "placed_buildings"`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// AddUserGoldGetRemaining
	mock.ExpectBegin()
	mock.ExpectQuery(`FROM "user_data"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "current_gold", "total_gold_capacity"}).
			AddRow("user-1", int64(0), int64(1000000)))
	mock.ExpectExec(`UPDATE "user_data"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// GetPlacedBuildingJSON
	mock.ExpectQuery(`FROM "placed_buildings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "building_id", "grid_x", "grid_y", "level", "is_broken"}).
			AddRow("pb-1", "user-1", "gold-mine", 0, 0, 1, false))

	// GetUserData (final)
	mock.ExpectQuery(`FROM "user_data"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "current_gold"}).AddRow("user-1", int64(100)))

	errCh := make(chan error, 1)
	go func() {
		errCh <- CollectAllResource("user-1", server)
	}()

	var msg struct {
		MsgType string `json:"msg_type"`
	}
	if err := client.ReadJSON(&msg); err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("CollectAllResource returned error: %v", err)
	}
	if msg.MsgType != "resource_collected" {
		t.Errorf("expected msg_type resource_collected, got %q", msg.MsgType)
	}
	requireMet(t, mock)
}
