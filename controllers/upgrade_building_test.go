package controllers

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func upgradeBuildingData(id string, useGems bool) struct {
	PlacedBuildingID string `json:"placed_building_id"`
	UseGems          bool   `json:"use_gems"`
} {
	return struct {
		PlacedBuildingID string `json:"placed_building_id"`
		UseGems          bool   `json:"use_gems"`
	}{PlacedBuildingID: id, UseGems: useGems}
}

func TestUpgradeBuilding_ConstructionUnderProgress(t *testing.T) {
	mock := newMockDB(t)

	server, client := newTestConnPair(t)

	mock.ExpectQuery(`FROM "construction_tasks"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	errCh := make(chan error, 1)
	go func() {
		errCh <- UpgradeBuilding("user-1", upgradeBuildingData("pb-1", false), server)
	}()

	_, raw, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	if len(raw) == 0 {
		t.Error("expected non-empty 'already under construction' error payload")
	}
	if err := <-errCh; err != nil {
		t.Fatalf("UpgradeBuilding returned error: %v", err)
	}
	requireMet(t, mock)
}

func TestUpgradeBuilding_Broken(t *testing.T) {
	mock := newMockDB(t)

	server, client := newTestConnPair(t)

	mock.ExpectQuery(`FROM "construction_tasks"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`is_broken`).
		WillReturnRows(sqlmock.NewRows([]string{"is_broken"}).AddRow(true))

	errCh := make(chan error, 1)
	go func() {
		errCh <- UpgradeBuilding("user-1", upgradeBuildingData("pb-1", false), server)
	}()

	_, raw, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	if len(raw) == 0 {
		t.Error("expected non-empty 'Cannot upgrade broken building' error payload")
	}
	if err := <-errCh; err != nil {
		t.Fatalf("UpgradeBuilding returned error: %v", err)
	}
	requireMet(t, mock)
}

func TestUpgradeBuilding_InsufficientTownHallLevel(t *testing.T) {
	mock := newMockDB(t)

	server, client := newTestConnPair(t)

	mock.ExpectQuery(`FROM "construction_tasks"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`is_broken`).
		WillReturnRows(sqlmock.NewRows([]string{"is_broken"}).AddRow(false))
	mock.ExpectQuery(`FROM "placed_buildings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "building_id", "level"}).
			AddRow("pb-1", "user-1", "wall-1", 1))
	mock.ExpectQuery(`FROM "upgrade_costs"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "building_id", "upgrade_to_level", "town_hall_level_required"}).
			AddRow("cost-1", "wall-1", 2, 8))
	mock.ExpectQuery(`FROM "user_data"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "town_hall_level"}).AddRow("user-1", 3))

	errCh := make(chan error, 1)
	go func() {
		errCh <- UpgradeBuilding("user-1", upgradeBuildingData("pb-1", false), server)
	}()

	_, raw, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	if len(raw) == 0 {
		t.Error("expected non-empty 'Town Hall Level' error payload")
	}
	if err := <-errCh; err != nil {
		t.Fatalf("UpgradeBuilding returned error: %v", err)
	}
	requireMet(t, mock)
}
