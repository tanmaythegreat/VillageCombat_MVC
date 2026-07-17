package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func doRegisterRequest(t *testing.T, method string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("failed to encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, "/register", &buf)
	rec := httptest.NewRecorder()
	RegisterHandler(rec, req)
	return rec
}

func TestRegisterHandler_MethodNotAllowed(t *testing.T) {
	rec := doRegisterRequest(t, http.MethodGet, nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestRegisterHandler_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("{not valid json"))
	rec := httptest.NewRecorder()
	RegisterHandler(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestRegisterHandler_MissingFields(t *testing.T) {
	rec := doRegisterRequest(t, http.MethodPost, map[string]string{
		"username": "alice",
		// email and password_text missing
	})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestRegisterHandler_WeakPassword(t *testing.T) {
	rec := doRegisterRequest(t, http.MethodPost, map[string]string{
		"username":      "alice",
		"email":         "alice@example.com",
		"password_text": "weak",
	})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for weak password, got %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestRegisterHandler_DuplicateUser(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnError(errors.New(`duplicate key value violates unique constraint "users_username_key"`))
	mock.ExpectRollback()

	rec := doRegisterRequest(t, http.MethodPost, map[string]string{
		"username":      "alice",
		"email":         "alice@example.com",
		"password_text": "Str0ng!Passw0rd",
	})

	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d, body=%s", rec.Code, rec.Body.String())
	}
	requireMet(t, mock)
}

func TestRegisterHandler_Success(t *testing.T) {
	mock := newMockDB(t)

	// RegisterUser's transaction: user -> user_data -> townhall placed_building -> user_status
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("user-1"))
	mock.ExpectQuery(`INSERT INTO "user_data"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("user-1"))
	mock.ExpectExec(`INSERT INTO "placed_buildings"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO "user_status"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// auth.GenerateJWT_Token -> models.AddRefreshToken (no existing token -> insert)
	mock.ExpectQuery(`FROM "refresh_tokens" WHERE user_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}))
	mock.ExpectExec(`INSERT INTO "refresh_tokens"`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rec := doRegisterRequest(t, http.MethodPost, map[string]string{
		"username":      "alice",
		"email":         "alice@example.com",
		"password_text": "Str0ng!Passw0rd",
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected non-empty access token in response")
	}
	requireMet(t, mock)
}
