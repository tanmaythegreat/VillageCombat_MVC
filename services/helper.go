package services

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"Village_combat/models"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/websocket"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestMain makes sure the JWT secret required by the auth package (which
// this package calls into for login/register/refresh) is present before any
// test runs. auth.getJWTSecretKey panics if JWT_SECRET_KEY is unset.
func testMain(m *testing.M) {
	if os.Getenv("JWT_SECRET_KEY") == "" {
		os.Setenv("JWT_SECRET_KEY", "test-secret-key-for-controllers-package")
	}
	os.Exit(m.Run())
}

// NewMockDB installs a sqlmock-backed *gorm.DB as models.DB for the
// duration of the test and restores the previous value afterwards. Mirrors
// the convention already used in models.Util_test.go / auth.JWT_Auth_test.go.
func NewMockDB(t *testing.T) sqlmock.Sqlmock {
	t.Helper()

	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	gdb, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		DisableAutomaticPing:   true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatalf("failed to open gorm with sqlmock conn: %v", err)
	}

	oldDB := models.DB
	models.DB = gdb

	t.Cleanup(func() {
		models.DB = oldDB
		sqlDB.Close()
	})

	return mock
}

// RequireMet fails the test with full detail if any mocked expectation
// wasn't satisfied (missing calls, calls in the wrong order, etc).
func RequireMet(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

// NewTestConnPair spins up a real client/server websocket connection pair
// backed by an httptest.Server. Most controller functions take a concrete
// *websocket.Conn (not an interface), so this is the simplest reliable way
// to exercise them end-to-end: pass `server` into the function under test,
// then read the resulting message(s) off `client`.
func NewTestConnPair(t *testing.T) (server *websocket.Conn, client *websocket.Conn) {
	t.Helper()

	upgrader := websocket.Upgrader{}
	serverConnCh := make(chan *websocket.Conn, 1)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		serverConnCh <- c
	}))

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	cc, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("client dial failed: %v", err)
	}

	sc := <-serverConnCh

	t.Cleanup(func() {
		cc.Close()
		sc.Close()
		srv.Close()
	})

	return sc, cc
}

// ResetBuildingStaticState snapshots the models package-level static caches
// that several controller functions read directly (models.BuildingSize,
// models.BuildingID_Category, the *_ID globals, models.ResourceLevelDetails)
// and restores them after the test. Individual tests populate whichever
// subset their code path touches.
func ResetBuildingStaticState(t *testing.T) {
	t.Helper()
	oldSize := models.BuildingSize
	oldCat := models.BuildingID_Category
	oldBarracks := models.Barracks_ID
	oldElixirStorage := models.ElixirStorage_ID
	oldGoldStorage := models.GoldStorage_ID
	oldDarkElixirStorage := models.DarkElixirStorage_ID
	oldGoldMine := models.GoldMine_ID
	oldElixirCollector := models.ElixirCollector_ID
	oldDarkElixirDrill := models.DarkElixirDrill_ID
	oldResourceLevelDetails := models.ResourceLevelDetails

	t.Cleanup(func() {
		models.BuildingSize = oldSize
		models.BuildingID_Category = oldCat
		models.Barracks_ID = oldBarracks
		models.ElixirStorage_ID = oldElixirStorage
		models.GoldStorage_ID = oldGoldStorage
		models.DarkElixirStorage_ID = oldDarkElixirStorage
		models.GoldMine_ID = oldGoldMine
		models.ElixirCollector_ID = oldElixirCollector
		models.DarkElixirDrill_ID = oldDarkElixirDrill
		models.ResourceLevelDetails = oldResourceLevelDetails
	})

	models.BuildingSize = make(map[string]struct {
		X int
		Y int
	})
	models.BuildingID_Category = make(map[string]models.BuildingCategory)
}
