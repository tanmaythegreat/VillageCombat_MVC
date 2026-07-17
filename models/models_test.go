package models

import (
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Users / auth
// ---------------------------------------------------------------------------

func TestGetUserByName_Found(t *testing.T) {
	mock := newMockDB(t)

	rows := sqlmock.NewRows([]string{"user_id", "username", "email", "password_hash", "created_at", "updated_at"}).
		AddRow("u1", "alice", "alice@example.com", "hash", time.Now(), time.Now())
	mock.ExpectQuery(`FROM "users" WHERE username`).WillReturnRows(rows)

	user, err := GetUserByName("alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.UserID != "u1" || user.Username != "alice" {
		t.Errorf("unexpected user: %+v", user)
	}
	requireMet(t, mock)
}

func TestGetUserByName_NotFound(t *testing.T) {
	mock := newMockDB(t)

	rows := sqlmock.NewRows([]string{"user_id", "username", "email", "password_hash", "created_at", "updated_at"})
	mock.ExpectQuery(`FROM "users" WHERE username`).WillReturnRows(rows)

	_, err := GetUserByName("nobody")
	if err == nil || err.Error() != "user not found" {
		t.Errorf("expected 'user not found' error, got %v", err)
	}
	requireMet(t, mock)
}

func TestGetUserByEmail_Found(t *testing.T) {
	mock := newMockDB(t)

	rows := sqlmock.NewRows([]string{"user_id", "username", "email", "password_hash", "created_at", "updated_at"}).
		AddRow("u1", "alice", "alice@example.com", "hash", time.Now(), time.Now())
	mock.ExpectQuery(`FROM "users" WHERE email`).WillReturnRows(rows)

	user, err := GetUserByEmail("alice@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "alice@example.com" {
		t.Errorf("unexpected user: %+v", user)
	}
	requireMet(t, mock)
}

func TestGetUserData(t *testing.T) {
	mock := newMockDB(t)

	rows := sqlmock.NewRows([]string{"user_id", "town_hall_level", "current_gold"}).
		AddRow("u1", 3, int64(500))
	mock.ExpectQuery(`FROM "user_data" WHERE user_id`).WillReturnRows(rows)

	ud, err := GetUserData("u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ud.TownHallLevel != 3 || ud.CurrentGold != 500 {
		t.Errorf("unexpected user data: %+v", ud)
	}
	requireMet(t, mock)
}

func TestAddRefreshToken_CreatesNewWhenNoneExists(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`FROM "refresh_tokens" WHERE user_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}))

	mock.ExpectExec(`INSERT INTO "refresh_tokens"`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := AddRefreshToken("u1", "hash", "1.2.3.4", "chrome", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	requireMet(t, mock)
}

func TestAddRefreshToken_UpdatesExisting(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`FROM "refresh_tokens" WHERE user_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}).AddRow("rt1", "u1"))

	mock.ExpectExec(`UPDATE "refresh_tokens" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := AddRefreshToken("u1", "newhash", "1.2.3.4", "chrome", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	requireMet(t, mock)
}

func TestRemoveRefreshToken(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectExec(`DELETE FROM "refresh_tokens" WHERE user_id`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := RemoveRefreshToken("u1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	requireMet(t, mock)
}

func TestGetRefreshTokenByUserID_Found(t *testing.T) {
	mock := newMockDB(t)

	rows := sqlmock.NewRows([]string{"id", "user_id", "jwt_token_hash"}).
		AddRow("rt1", "u1", "hash")
	mock.ExpectQuery(`FROM "refresh_tokens" WHERE user_id`).WillReturnRows(rows)

	tok, err := GetRefreshTokenByUserID("u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok.ID != "rt1" {
		t.Errorf("unexpected token: %+v", tok)
	}
	requireMet(t, mock)
}

func TestGetRefreshTokenByUserID_NotFound(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`FROM "refresh_tokens" WHERE user_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}))

	_, err := GetRefreshTokenByUserID("ghost")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	requireMet(t, mock)
}

func TestRegisterUser_Success(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("u1"))
	mock.ExpectQuery(`INSERT INTO "user_data"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("u1"))
	mock.ExpectExec(`INSERT INTO "placed_buildings"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO "user_status"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	TownHall_ID = "th-config-id"

	user, userData, err := RegisterUser("alice", "hash", "alice@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil || userData == nil {
		t.Fatalf("expected non-nil user/userData")
	}
	requireMet(t, mock)
}

func TestRegisterUser_RollbackOnUserCreateError(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnError(errors.New("unique_violation: username already exists"))
	mock.ExpectRollback()

	_, _, err := RegisterUser("alice", "hash", "alice@example.com")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	requireMet(t, mock)
}

// ---------------------------------------------------------------------------
// Buildings
// ---------------------------------------------------------------------------

func TestGetPlacedBuildings(t *testing.T) {
	mock := newMockDB(t)

	rows := sqlmock.NewRows([]string{"id", "user_id", "building_id", "grid_x", "grid_y", "level"}).
		AddRow("pb1", "u1", "b1", 0, 0, 1).
		AddRow("pb2", "u1", "b2", 1, 1, 2)
	mock.ExpectQuery(`FROM "placed_buildings" WHERE user_id`).WillReturnRows(rows)

	buildings, err := GetPlacedBuildings("u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(buildings) != 2 {
		t.Errorf("expected 2 buildings, got %d", len(buildings))
	}
	requireMet(t, mock)
}

func TestGetPlacedBuilding(t *testing.T) {
	mock := newMockDB(t)

	rows := sqlmock.NewRows([]string{"id", "user_id", "building_id", "level"}).
		AddRow("pb1", "u1", "b1", 3)
	mock.ExpectQuery(`FROM "placed_buildings" WHERE user_id`).WillReturnRows(rows)

	pb, err := GetPlacedBuilding("u1", "pb1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pb.Level != 3 {
		t.Errorf("unexpected level: %d", pb.Level)
	}
	requireMet(t, mock)
}

func TestUpdatePlacedBuilding(t *testing.T) {
	mock := newMockDB(t)

	rows := sqlmock.NewRows([]string{"id", "user_id", "building_id", "level"}).
		AddRow("pb1", "u1", "b1", 2)
	mock.ExpectQuery(`FROM "placed_buildings" WHERE id`).WillReturnRows(rows)

	mock.ExpectExec(`UPDATE "placed_buildings" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	old, err := UpdatePlacedBuilding("u1", "pb1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if old.Level != 2 {
		t.Errorf("unexpected old building: %+v", old)
	}
	requireMet(t, mock)
}

func TestDecreaseUpdateTime(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`UPDATE "placed_buildings" SET "last_updated_at"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "last_updated_at"}).AddRow("pb1", time.Now()))

	err := DecreaseUpdateTime("u1", "pb1", 2.5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	requireMet(t, mock)
}

func TestGetBuildingHealth(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`FROM "building_level_stats" WHERE building_id`).
		WillReturnRows(sqlmock.NewRows([]string{"health"}).AddRow(int64(1500)))

	health, err := GetBuildingHealth("b1", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if health != 1500 {
		t.Errorf("expected 1500, got %d", health)
	}
	requireMet(t, mock)
}

func TestConstructBuilding(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`INSERT INTO "placed_buildings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("pb-new"))
	mock.ExpectQuery(`INSERT INTO "construction_tasks"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("task-1"))

	newBuilding, task, err := ConstructBuilding("u1", "b1", 5, 6, DB, 3600)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newBuilding.GridX != 5 || newBuilding.GridY != 6 {
		t.Errorf("unexpected building placement: %+v", newBuilding)
	}
	_ = task
	requireMet(t, mock)
}

func TestUpgradeBuilding(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`INSERT INTO "construction_tasks"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("task-2"))

	task, err := UpgradeBuilding("u1", "pb1", DB, 1800)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.TaskType != BuildingUpgrade {
		t.Errorf("expected BuildingUpgrade task type, got %v", task.TaskType)
	}
	requireMet(t, mock)
}

func TestGetNearByBuildings_UsesDefaultSizeWhenMissing(t *testing.T) {
	mock := newMockDB(t)

	// Reset and only register a size for "known-building"; "unknown-building"
	// should fall back to the 1x1 default inside GetNearByBuildings.
	BuildingSize = map[string]struct {
		X int
		Y int
	}{
		"known-building": {X: 3, Y: 2},
	}

	rows := sqlmock.NewRows([]string{"id", "user_id", "building_id", "grid_x", "grid_y"}).
		AddRow("pb1", "u1", "known-building", 10, 10).
		AddRow("pb2", "u1", "unknown-building", 20, 20)
	mock.ExpectQuery(`FROM "placed_buildings" WHERE user_id`).WillReturnRows(rows)

	nearby, err := GetNearByBuildings("u1", 15, 15)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nearby) != 2 {
		t.Fatalf("expected 2 nearby buildings, got %d", len(nearby))
	}
	if nearby[0].Size_x != 3 || nearby[0].Size_y != 2 {
		t.Errorf("expected known size 3x2, got %dx%d", nearby[0].Size_x, nearby[0].Size_y)
	}
	if nearby[1].Size_x != 1 || nearby[1].Size_y != 1 {
		t.Errorf("expected default size 1x1 for unknown building, got %dx%d", nearby[1].Size_x, nearby[1].Size_y)
	}
	requireMet(t, mock)
}

func TestUpdatePlacedBuildingPosition_Success(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`UPDATE "placed_buildings" SET`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "grid_x", "grid_y"}).AddRow("pb1", 5, 5))

	_, err := UpdatePlacedBuildingPosition("u1", "pb1", 5, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	requireMet(t, mock)
}

func TestUpdatePlacedBuildingPosition_NotFound(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`UPDATE "placed_buildings" SET`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "grid_x", "grid_y"}))

	_, err := UpdatePlacedBuildingPosition("u1", "ghost", 5, 5)
	if err == nil || err.Error() != "building not found or user unauthorized" {
		t.Errorf("expected 'building not found or user unauthorized', got %v", err)
	}
	requireMet(t, mock)
}

func TestSetBrokenBuilding(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectExec(`UPDATE "placed_buildings" SET "is_broken"`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := SetBrokenBuilding("u1", "pb1", true, DB)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	requireMet(t, mock)
}

func TestSetBrokenBuildings_EmptyNoOp(t *testing.T) {
	mock := newMockDB(t)
	// No expectations set at all: an empty ID slice should short-circuit
	// before touching the DB.
	if err := SetBrokenBuildings("u1", nil, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	requireMet(t, mock)
}

func TestIsBuildingBroken_Found(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`SELECT "is_broken" FROM "placed_buildings"`).
		WillReturnRows(sqlmock.NewRows([]string{"is_broken"}).AddRow(true))

	broken, err := IsBuildingBroken("u1", "pb1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !broken {
		t.Errorf("expected broken=true")
	}
	requireMet(t, mock)
}

func TestIsBuildingBroken_NotFound(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`SELECT "is_broken" FROM "placed_buildings"`).
		WillReturnRows(sqlmock.NewRows([]string{"is_broken"}))

	_, err := IsBuildingBroken("u1", "ghost")
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("expected gorm.ErrRecordNotFound, got %v", err)
	}
	requireMet(t, mock)
}

// ---------------------------------------------------------------------------
// Construction tasks
// ---------------------------------------------------------------------------

func TestGetConstructionCost(t *testing.T) {
	mock := newMockDB(t)

	rows := sqlmock.NewRows([]string{"id", "building_id", "upgrade_to_level", "gold_required"}).
		AddRow("cost1", "b1", 4, int64(2000))
	mock.ExpectQuery(`FROM "upgrade_costs" WHERE building_id`).WillReturnRows(rows)

	cost, err := GetConstructionCost("b1", 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cost.GoldRequired != 2000 {
		t.Errorf("unexpected cost: %+v", cost)
	}
	requireMet(t, mock)
}

func TestIsConstructionUnderProgress(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`FROM "construction_tasks" WHERE user_id`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	inProgress, err := IsConstructionUnderProgress("u1", "pb1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !inProgress {
		t.Errorf("expected true")
	}
	requireMet(t, mock)
}

func TestCheckIsConstructionComplete_NoTasks(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`DELETE FROM "construction_tasks"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "task_type"}))
	mock.ExpectCommit()
	// No ExpectRollback here: once Commit succeeds, database/sql marks the
	// transaction "done" and the code's deferred tx.Rollback() returns
	// sql.ErrTxDone locally without ever reaching the driver/mock.

	tasks, buildings, err := CheckIsConstructionComplete("u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 0 || len(buildings) != 0 {
		t.Errorf("expected no completed tasks/buildings, got %d/%d", len(tasks), len(buildings))
	}
	requireMet(t, mock)
}

func TestIncrementUserTownHallLevel(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectExec(`UPDATE "user_data" SET "town_hall_level"`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := IncrementUserTownHallLevel("u1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	requireMet(t, mock)
}

// ---------------------------------------------------------------------------
// Troops
// ---------------------------------------------------------------------------

func TestGetUserTrainedTroops(t *testing.T) {
	mock := newMockDB(t)

	rows := sqlmock.NewRows([]string{"id", "user_id", "troop_id", "level", "count"}).
		AddRow("tt1", "u1", "troop1", 2, 10)
	mock.ExpectQuery(`FROM "trained_troops" WHERE user_id`).WillReturnRows(rows)

	raw, err := GetUserTrainedTroops("u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(raw) == 0 {
		t.Errorf("expected non-empty JSON")
	}
	requireMet(t, mock)
}

func TestGetTroopUpgradeCost(t *testing.T) {
	mock := newMockDB(t)

	rows := sqlmock.NewRows([]string{"id", "troop_id", "upgrade_to_level", "elixir_required"}).
		AddRow("cost1", "troop1", 3, int64(500))
	mock.ExpectQuery(`FROM "upgrade_costs" WHERE troop_id`).WillReturnRows(rows)

	cost, err := GetTroopUpgradeCost("troop1", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cost.ElixirRequired != 500 {
		t.Errorf("unexpected cost: %+v", cost)
	}
	requireMet(t, mock)
}

func TestAddTroopsToUser_NewTroop(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`FROM "trained_troops" WHERE user_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "troop_id", "level", "count"}))

	mock.ExpectQuery(`INSERT INTO "trained_troops"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("tt-new"))

	err := AddTroopsToUser("u1", "troop1", 1, 5, DB)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	requireMet(t, mock)
}

func TestAddTroopsToUser_ExistingTroop(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`FROM "trained_troops" WHERE user_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "troop_id", "level", "count"}).
			AddRow("tt1", "u1", "troop1", 1, 5))

	mock.ExpectExec(`UPDATE "trained_troops" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := AddTroopsToUser("u1", "troop1", 1, 5, DB)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	requireMet(t, mock)
}

func TestSubtractTroopsOfUser_NegativeCountError(t *testing.T) {
	newMockDB(t) // no DB calls expected

	_, err := SubtractTroopsOfUser("u1", "troop1", 1, -5, DB)
	if err == nil || err.Error() != "cannot subtract a negative amount of troops" {
		t.Errorf("expected negative-count error, got %v", err)
	}
}

func TestSubtractTroopsOfUser_ZeroCountNoop(t *testing.T) {
	newMockDB(t) // no DB calls expected

	ok, err := SubtractTroopsOfUser("u1", "troop1", 1, 0, DB)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Errorf("expected true for zero-count no-op")
	}
}

func TestSubtractTroopsOfUser_Success(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectExec(`UPDATE "trained_troops" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	ok, err := SubtractTroopsOfUser("u1", "troop1", 1, 5, DB)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Errorf("expected true")
	}
	requireMet(t, mock)
}

func TestSubtractTroopsOfUser_InsufficientCount(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectExec(`UPDATE "trained_troops" SET`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	ok, err := SubtractTroopsOfUser("u1", "troop1", 1, 999, DB)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Errorf("expected false when not enough troops available")
	}
	requireMet(t, mock)
}

func TestGetCapacityDifference(t *testing.T) {
	mock := newMockDB(t)

	rows := sqlmock.NewRows([]string{"level", "storage_capacity"}).
		AddRow(1, int64(1000)).
		AddRow(2, int64(2500))
	mock.ExpectQuery(`FROM "resource_building_level_stats" WHERE building_id`).WillReturnRows(rows)

	diff, err := GetCapacityDifference("gold-storage", 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if diff != 1500 {
		t.Errorf("expected 1500, got %d", diff)
	}
	requireMet(t, mock)
}

func TestGetTroopCapacityDifference(t *testing.T) {
	mock := newMockDB(t)
	Barracks_ID = "barracks-id"

	rows := sqlmock.NewRows([]string{"level", "troop_capacity"}).
		AddRow(1, 20).
		AddRow(2, 40)
	mock.ExpectQuery(`FROM "army_building_level_stats" WHERE building_id`).WillReturnRows(rows)

	diff, err := GetTroopCapacityDifference(1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if diff != 20 {
		t.Errorf("expected 20, got %d", diff)
	}
	requireMet(t, mock)
}

func TestHasTroopTrainingTask(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`FROM "construction_tasks" WHERE user_id`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	has, err := HasTroopTrainingTask("u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !has {
		t.Errorf("expected true")
	}
	requireMet(t, mock)
}

// ---------------------------------------------------------------------------
// Resources
// ---------------------------------------------------------------------------

func TestAddUserGold(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`UPDATE "user_data" SET "current_gold"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "current_gold"}).AddRow("u1", int64(1500)))

	ud, err := AddUserGold("u1", 500)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ud.CurrentGold != 1500 {
		t.Errorf("unexpected user data: %+v", ud)
	}
	requireMet(t, mock)
}

func TestAddUserElixir(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`UPDATE "user_data" SET "current_elixir"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "current_elixir"}).AddRow("u1", int64(800)))

	ud, err := AddUserElixir("u1", 300)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ud.CurrentElixir != 800 {
		t.Errorf("unexpected user data: %+v", ud)
	}
	requireMet(t, mock)
}

func TestAddUserDarkElixir(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`UPDATE "user_data" SET "current_dark_elixir"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "current_dark_elixir"}).AddRow("u1", int64(50)))

	ud, err := AddUserDarkElixir("u1", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ud.CurrentDarkElixir != 50 {
		t.Errorf("unexpected user data: %+v", ud)
	}
	requireMet(t, mock)
}

func TestAddUserGems(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`UPDATE "user_data" SET "current_gems"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "current_gems"}).AddRow("u1", int64(20)))

	ud, err := AddUserGems("u1", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ud.CurrentGems != 20 {
		t.Errorf("unexpected user data: %+v", ud)
	}
	requireMet(t, mock)
}

func TestAddUserCapacity(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`UPDATE "user_data" SET`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "total_gold_capacity", "total_troop_capacity"}).
			AddRow("u1", int64(5000), 60))

	ud, err := AddUserCapacity("u1", 1000, 500, 100, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ud.TotalGoldCapacity != 5000 || ud.TotalTroopCapacity != 60 {
		t.Errorf("unexpected user data: %+v", ud)
	}
	requireMet(t, mock)
}

func TestGetGenerationRate(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`SELECT "generation_rate_per_hour" FROM "resource_building_level_stats"`).
		WillReturnRows(sqlmock.NewRows([]string{"generation_rate_per_hour"}).AddRow(250.5))

	rate, err := GetGenerationRate("gold-mine", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rate != 250.5 {
		t.Errorf("expected 250.5, got %v", rate)
	}
	requireMet(t, mock)
}

func TestAddUserGoldGetRemaining_WithinCapacity(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`FROM "user_data" WHERE user_id`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "current_gold", "total_gold_capacity"}).
			AddRow("u1", int64(100), int64(1000)))
	mock.ExpectExec(`UPDATE "user_data" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	ud, err, extra := AddUserGoldGetRemaining("u1", 200)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ud.CurrentGold != 300 {
		t.Errorf("expected current gold 300, got %d", ud.CurrentGold)
	}
	if extra != 0 {
		t.Errorf("expected no overflow, got %d", extra)
	}
	requireMet(t, mock)
}

func TestAddUserGoldGetRemaining_ExceedsCapacity(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`FROM "user_data" WHERE user_id`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "current_gold", "total_gold_capacity"}).
			AddRow("u1", int64(900), int64(1000)))
	mock.ExpectExec(`UPDATE "user_data" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	ud, err, extra := AddUserGoldGetRemaining("u1", 500)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ud.CurrentGold != 1000 {
		t.Errorf("expected current gold capped at 1000, got %d", ud.CurrentGold)
	}
	if extra != 400 {
		t.Errorf("expected overflow of 400, got %d", extra)
	}
	requireMet(t, mock)
}

// ---------------------------------------------------------------------------
// Battle / user status
// ---------------------------------------------------------------------------

func TestSetUserBattleStatus(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectExec(`UPDATE "user_status" SET "in_battle"`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := SetUserBattleStatus("u1", true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	requireMet(t, mock)
}

func TestAdjustAttackPower(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectExec(`UPDATE "user_status" SET "attack_power"`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := AdjustAttackPower("u1", 10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	requireMet(t, mock)
}

func TestAdjustDefencePower(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectExec(`UPDATE "user_status" SET "defence_power"`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := AdjustDefencePower("u1", -5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	requireMet(t, mock)
}

func TestMarkDefendedNow(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectExec(`UPDATE "user_status" SET "last_defended"`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := MarkDefendedNow("u1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	requireMet(t, mock)
}

func TestInsertBattleHistory(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`INSERT INTO "battle_history"`).
		WillReturnRows(sqlmock.NewRows([]string{"battle_id"}).AddRow("b1"))

	id, err := InsertBattleHistory(BattleHistory{AttackerName: "alice", DefenderName: "bob"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = id
	requireMet(t, mock)
}

func TestGetUsername(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`SELECT "username" FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"username"}).AddRow("alice"))

	name, err := GetUsername("u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "alice" {
		t.Errorf("expected alice, got %s", name)
	}
	requireMet(t, mock)
}

func TestGetUser(t *testing.T) {
	mock := newMockDB(t)

	rows := sqlmock.NewRows([]string{"user_id", "username", "email"}).
		AddRow("u1", "alice", "alice@example.com")
	mock.ExpectQuery(`FROM "users" WHERE user_id`).WillReturnRows(rows)

	user, err := GetUser("u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Username != "alice" {
		t.Errorf("unexpected user: %+v", user)
	}
	requireMet(t, mock)
}

func TestSaveBattleRecord(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectExec(`UPDATE "battle_record" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	record := &BattleRecord{BattleID: "b1"}
	if err := SaveBattleRecord(record); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	requireMet(t, mock)
}

func TestGetBattleRecord_Found(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`FROM "battle_record" WHERE battle_id`).
		WillReturnRows(sqlmock.NewRows([]string{"battle_id"}).AddRow("b1"))

	record, err := GetBattleRecord("b1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if record.BattleID != "b1" {
		t.Errorf("unexpected record: %+v", record)
	}
	requireMet(t, mock)
}

func TestGetBattleRecord_NotFound(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`FROM "battle_record" WHERE battle_id`).
		WillReturnRows(sqlmock.NewRows([]string{"battle_id"}))

	_, err := GetBattleRecord("ghost")
	if err == nil || err.Error() != "battle record not found" {
		t.Errorf("expected 'battle record not found', got %v", err)
	}
	requireMet(t, mock)
}
