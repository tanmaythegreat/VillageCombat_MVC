package controllers

import (
	"testing"
	"time"

	"Village_combat/models"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func collectResourceData(placedBuildingID string) struct {
	PlacedBuildingId string `json:"placed_building_id"`
} {
	return struct {
		PlacedBuildingId string `json:"placed_building_id"`
	}{PlacedBuildingId: placedBuildingID}
}

func TestCollectResource_BrokenBuilding(t *testing.T) {
	mock := newMockDB(t)

	server, client := newTestConnPair(t)

	mock.ExpectQuery(`is_broken`).
		WillReturnRows(sqlmock.NewRows([]string{"is_broken"}).AddRow(true))

	errCh := make(chan error, 1)
	go func() {
		errCh <- CollectResource("user-1", collectResourceData("pb-1"), server)
	}()

	_, raw, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	if len(raw) == 0 {
		t.Error("expected a non-empty error payload for broken building")
	}
	if err := <-errCh; err != nil {
		t.Fatalf("CollectResource returned error: %v", err)
	}
	requireMet(t, mock)
}

func TestCollectResource_Success_GoldMine(t *testing.T) {
	mock := newMockDB(t)
	resetBuildingStaticState(t)
	models.GoldMine_ID = "gold-mine"
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

	// IsBuildingBroken
	mock.ExpectQuery(`is_broken`).
		WillReturnRows(sqlmock.NewRows([]string{"is_broken"}).AddRow(false))

	// UpdatePlacedBuilding: select old row, then bump last_updated_at
	mock.ExpectQuery(`FROM "placed_buildings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "building_id", "level", "last_updated_at"}).
			AddRow("pb-1", "user-1", "gold-mine", 1, time.Now().Add(-time.Hour)))
	mock.ExpectExec(`UPDATE "placed_buildings"`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// AddUserGoldGetRemaining: transaction
	mock.ExpectBegin()
	mock.ExpectQuery(`FROM "user_data"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "current_gold", "total_gold_capacity"}).
			AddRow("user-1", int64(0), int64(1000000)))
	mock.ExpectExec(`UPDATE "user_data"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// DecreaseUpdateTime: UPDATE ... RETURNING
	mock.ExpectQuery(`UPDATE "placed_buildings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("pb-1"))

	errCh := make(chan error, 1)
	go func() {
		errCh <- CollectResource("user-1", collectResourceData("pb-1"), server)
	}()

	var msg struct {
		MsgType string `json:"msg_type"`
	}
	if err := client.ReadJSON(&msg); err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("CollectResource returned error: %v", err)
	}
	if msg.MsgType != "resource_collected" {
		t.Errorf("expected msg_type resource_collected, got %q", msg.MsgType)
	}
	requireMet(t, mock)
}
