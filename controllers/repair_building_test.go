package controllers

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func repairBuildingData(id string, useGems bool) struct {
	PlacedBuildingID string `json:"placed_building_id"`
	UseGems          bool   `json:"use_gems"`
} {
	return struct {
		PlacedBuildingID string `json:"placed_building_id"`
		UseGems          bool   `json:"use_gems"`
	}{PlacedBuildingID: id, UseGems: useGems}
}

func TestRepairBuilding_ConstructionUnderProgress(t *testing.T) {
	mock := newMockDB(t)

	server, client := newTestConnPair(t)

	mock.ExpectQuery(`FROM "construction_tasks"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	errCh := make(chan error, 1)
	go func() {
		errCh <- RepairBuilding("user-1", repairBuildingData("pb-1", false), server)
	}()

	_, raw, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	if len(raw) == 0 {
		t.Error("expected non-empty 'already under construction' error payload")
	}
	if err := <-errCh; err != nil {
		t.Fatalf("RepairBuilding returned error: %v", err)
	}
	requireMet(t, mock)
}

func TestRepairBuilding_NotBroken(t *testing.T) {
	mock := newMockDB(t)

	server, client := newTestConnPair(t)

	mock.ExpectQuery(`FROM "construction_tasks"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`is_broken`).
		WillReturnRows(sqlmock.NewRows([]string{"is_broken"}).AddRow(false))

	errCh := make(chan error, 1)
	go func() {
		errCh <- RepairBuilding("user-1", repairBuildingData("pb-1", false), server)
	}()

	_, raw, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	if len(raw) == 0 {
		t.Error("expected non-empty 'Cannot Repair unbroken building' error payload")
	}
	if err := <-errCh; err != nil {
		t.Fatalf("RepairBuilding returned error: %v", err)
	}
	requireMet(t, mock)
}
