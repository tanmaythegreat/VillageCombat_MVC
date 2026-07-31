package troops

import (
	"Village_combat/services"
	"testing"

	"Village_combat/models"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func trainTroopData(barrackID, troopID string, levelFrom, count int, useGems bool) struct {
	BarrackPlacedBuildingID string `json:"barrack_placed_building_id"`
	TroopId                 string `json:"troop_id"`
	LevelFrom               int    `json:"level_from"`
	Count                   int    `json:"count"`
	UseGems                 bool   `json:"use_gems"`
} {
	return struct {
		BarrackPlacedBuildingID string `json:"barrack_placed_building_id"`
		TroopId                 string `json:"troop_id"`
		LevelFrom               int    `json:"level_from"`
		Count                   int    `json:"count"`
		UseGems                 bool   `json:"use_gems"`
	}{BarrackPlacedBuildingID: barrackID, TroopId: troopID, LevelFrom: levelFrom, Count: count, UseGems: useGems}
}

func TestTrainTroop_ConstructionUnderProgress(t *testing.T) {
	mock := services.NewMockDB(t)

	server, client := services.NewTestConnPair(t)

	mock.ExpectQuery(`FROM "construction_tasks"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	errCh := make(chan error, 1)
	go func() {
		errCh <- TrainTroop("user-1", trainTroopData("barrack-1", "archer", 0, 1, false), server)
	}()

	_, raw, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	if len(raw) == 0 {
		t.Error("expected non-empty 'already under construction' error payload")
	}
	if err := <-errCh; err != nil {
		t.Fatalf("TrainTroop returned error: %v", err)
	}
	services.RequireMet(t, mock)
}

func TestTrainTroop_AlreadyTraining(t *testing.T) {
	mock := services.NewMockDB(t)

	server, client := services.NewTestConnPair(t)

	mock.ExpectQuery(`FROM "construction_tasks"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`FROM "construction_tasks"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	errCh := make(chan error, 1)
	go func() {
		errCh <- TrainTroop("user-1", trainTroopData("barrack-1", "archer", 0, 1, false), server)
	}()

	_, raw, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	if len(raw) == 0 {
		t.Error("expected non-empty 'already training' error payload")
	}
	if err := <-errCh; err != nil {
		t.Fatalf("TrainTroop returned error: %v", err)
	}
	services.RequireMet(t, mock)
}

func TestTrainTroop_WrongBuildingType(t *testing.T) {
	mock := services.NewMockDB(t)
	services.ResetBuildingStaticState(t)
	models.Barracks_ID = "barracks-config-id"

	server, client := services.NewTestConnPair(t)

	mock.ExpectQuery(`FROM "construction_tasks"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`FROM "construction_tasks"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`FROM "placed_buildings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "building_id"}).
			AddRow("barrack-1", "user-1", "not-a-barracks"))

	errCh := make(chan error, 1)
	go func() {
		errCh <- TrainTroop("user-1", trainTroopData("barrack-1", "archer", 0, 1, false), server)
	}()

	_, raw, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	if len(raw) == 0 {
		t.Error("expected non-empty 'Can only Train in Barracks' error payload")
	}
	if err := <-errCh; err != nil {
		t.Fatalf("TrainTroop returned error: %v", err)
	}
	services.RequireMet(t, mock)
}

func TestTrainTroop_BrokenBarracks(t *testing.T) {
	mock := services.NewMockDB(t)
	services.ResetBuildingStaticState(t)
	models.Barracks_ID = "barracks-config-id"

	server, client := services.NewTestConnPair(t)

	mock.ExpectQuery(`FROM "construction_tasks"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`FROM "construction_tasks"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`FROM "placed_buildings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "building_id"}).
			AddRow("barrack-1", "user-1", "barracks-config-id"))
	mock.ExpectQuery(`is_broken`).
		WillReturnRows(sqlmock.NewRows([]string{"is_broken"}).AddRow(true))

	errCh := make(chan error, 1)
	go func() {
		errCh <- TrainTroop("user-1", trainTroopData("barrack-1", "archer", 0, 1, false), server)
	}()

	_, raw, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	if len(raw) == 0 {
		t.Error("expected non-empty 'broken barracks' error payload")
	}
	if err := <-errCh; err != nil {
		t.Fatalf("TrainTroop returned error: %v", err)
	}
	services.RequireMet(t, mock)
}
