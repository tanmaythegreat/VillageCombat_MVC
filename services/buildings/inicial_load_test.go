package buildings

import (
	"Village_combat/services"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestInitialLoad_Success(t *testing.T) {
	mock := services.NewMockDB(t)

	server, client := services.NewTestConnPair(t)

	mock.ExpectQuery(`FROM "trained_troops"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "troop_id", "level", "count"}))
	mock.ExpectQuery(`FROM "placed_buildings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "building_id", "grid_x", "grid_y", "level", "is_broken"}))
	mock.ExpectQuery(`FROM "construction_tasks"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}))
	mock.ExpectQuery(`FROM "user_data"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "town_hall_level"}).AddRow("user-1", 5))
	mock.ExpectQuery(`FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "username"}).AddRow("user-1", "alice"))

	errCh := make(chan error, 1)
	go func() {
		errCh <- InitialLoad("user-1", server)
	}()

	var msg struct {
		MsgType string `json:"msg_type"`
		User    struct {
			Username string `json:"username"`
		} `json:"user"`
	}
	if err := client.ReadJSON(&msg); err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("InitialLoad returned error: %v", err)
	}
	if msg.MsgType != "building_troop_of_user" {
		t.Errorf("expected msg_type building_troop_of_user, got %q", msg.MsgType)
	}
	if msg.User.Username != "alice" {
		t.Errorf("expected username alice, got %q", msg.User.Username)
	}
	services.RequireMet(t, mock)
}

func TestInitialLoad_PropagatesEarlyError(t *testing.T) {
	mock := services.NewMockDB(t)

	server, client := services.NewTestConnPair(t)

	mock.ExpectQuery(`FROM "trained_troops"`).
		WillReturnError(errors.New("connection lost"))

	errCh := make(chan error, 1)
	go func() {
		errCh <- InitialLoad("user-1", server)
	}()

	var msg map[string]string
	if err := client.ReadJSON(&msg); err != nil {
		t.Fatalf("failed to read error payload: %v", err)
	}
	if msg["status"] != "error" {
		t.Errorf("expected an error status payload, got %v", msg)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("InitialLoad (SendError path) returned error: %v", err)
	}
	services.RequireMet(t, mock)
}
