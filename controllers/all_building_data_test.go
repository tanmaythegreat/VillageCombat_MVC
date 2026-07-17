package controllers

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestAllBuildingDataLoad_Success(t *testing.T) {
	mock := newMockDB(t)
	resetBuildingStaticState(t) // empty models.BuildingSize -> no per-building level-0 lookups

	server, client := newTestConnPair(t)

	mock.ExpectQuery(`FROM "building_configs_base"`).
		WillReturnRows(sqlmock.NewRows([]string{"building_id", "name"}))
	// GetAllTroopsDataJSON preloads LevelStats/UpgradeCosts, but with zero
	// base rows GORM has nothing to preload for and issues no further
	// queries.
	mock.ExpectQuery(`FROM "troop_configs"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
	mock.ExpectQuery(`FROM "defense_building_stats"`).
		WillReturnRows(sqlmock.NewRows([]string{"building_id"}))
	mock.ExpectQuery(`FROM "army_building_stats"`).
		WillReturnRows(sqlmock.NewRows([]string{"building_id"}))
	mock.ExpectQuery(`FROM "resource_building_stats"`).
		WillReturnRows(sqlmock.NewRows([]string{"building_id"}))
	mock.ExpectQuery(`FROM "placed_buildings"`).
		WillReturnRows(sqlmock.NewRows([]string{"building_id", "level"}))

	errCh := make(chan error, 1)
	go func() {
		errCh <- AllBuildingDataLoad("user-1", server)
	}()

	var msg struct {
		MsgType string `json:"msg_type"`
	}
	if err := client.ReadJSON(&msg); err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("AllBuildingDataLoad returned error: %v", err)
	}
	if msg.MsgType != "building_troop" {
		t.Errorf("expected msg_type building_troop, got %q", msg.MsgType)
	}
	requireMet(t, mock)
}
