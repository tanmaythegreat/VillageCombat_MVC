package buildings

import (
	"Village_combat/services"
	"testing"
	"time"

	"Village_combat/models"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestCheckConstructionWork_NothingComplete(t *testing.T) {
	mock := services.NewMockDB(t)

	server, client := services.NewTestConnPair(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`DELETE FROM "construction_tasks"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectCommit()

	if err := CheckConstructionWork("user-1", server); err != nil {
		t.Fatalf("CheckConstructionWork returned error: %v", err)
	}

	// Nothing should have been written to the client when there's no
	// completed work.
	client.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	if _, _, err := client.ReadMessage(); err == nil {
		t.Error("expected no message to be sent when no construction is complete")
	}
	services.RequireMet(t, mock)
}

func TestCheckConstructionWork_BuildingUpgradeCompletes(t *testing.T) {
	mock := services.NewMockDB(t)
	services.ResetBuildingStaticState(t)
	models.BuildingID_Category["wall-1"] = models.Wall

	server, client := services.NewTestConnPair(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`DELETE FROM "construction_tasks"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "task_type", "placed_building_id", "started_at", "duration_seconds"}).
			AddRow("task-1", "user-1", "building_upgrade", "pb-1", time.Now().Add(-time.Minute), 60))
	mock.ExpectQuery(`UPDATE "placed_buildings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "building_id", "grid_x", "grid_y", "level", "is_broken"}).
			AddRow("pb-1", "user-1", "wall-1", 0, 0, 2, false))
	mock.ExpectCommit()

	// GetBuildingDataOfLevelJSON("wall-1", 2)
	mock.ExpectQuery(`FROM "building_level_stats"`).
		WillReturnRows(sqlmock.NewRows([]string{"building_id", "level", "health"}).AddRow("wall-1", 2, 500))
	mock.ExpectQuery(`FROM "upgrade_costs"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	// AddUserCapacity (all deltas zero since category is Wall)
	mock.ExpectQuery(`UPDATE "user_data"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("user-1"))

	errCh := make(chan error, 1)
	go func() {
		errCh <- CheckConstructionWork("user-1", server)
	}()

	var msg struct {
		MsgType string `json:"msg_type"`
	}
	if err := client.ReadJSON(&msg); err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("CheckConstructionWork returned error: %v", err)
	}
	if msg.MsgType != "construction_completed" {
		t.Errorf("expected msg_type construction_completed, got %q", msg.MsgType)
	}
	services.RequireMet(t, mock)
}
