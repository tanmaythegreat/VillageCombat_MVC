package buildings

import (
	"Village_combat/services"
	"os"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// TestMain makes sure the JWT secret required by the auth package (which
// this package calls into for login/register/refresh) is present before any
// test runs. auth.getJWTSecretKey panics if JWT_SECRET_KEY is unset.
func TestMain(m *testing.M) {
	if os.Getenv("JWT_SECRET_KEY") == "" {
		os.Setenv("JWT_SECRET_KEY", "test-secret-key-for-controllers-package")
	}
	os.Exit(m.Run())
}
func TestAllBuildingDataLoad_Success(t *testing.T) {
	mock := services.NewMockDB(t)
	services.ResetBuildingStaticState(t) // empty models.BuildingSize -> no per-building level-0 lookups

	server, client := services.NewTestConnPair(t)

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
	services.RequireMet(t, mock)
}
