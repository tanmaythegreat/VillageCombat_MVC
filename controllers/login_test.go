package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"golang.org/x/crypto/bcrypt"
)

func doLoginRequest(t *testing.T, method string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("failed to encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, "/login", &buf)
	rec := httptest.NewRecorder()
	LoginHandler(rec, req)
	return rec
}

func TestLoginHandler_MethodNotAllowed(t *testing.T) {
	rec := doLoginRequest(t, http.MethodGet, nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestLoginHandler_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("{bad"))
	rec := httptest.NewRecorder()
	LoginHandler(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestLoginHandler_MissingFields(t *testing.T) {
	rec := doLoginRequest(t, http.MethodPost, map[string]string{
		"password_text": "whatever",
		// no username, no email
	})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestLoginHandler_UserNotFound(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "username", "email", "password_hash"}))

	rec := doLoginRequest(t, http.MethodPost, map[string]string{
		"username":      "ghost",
		"password_text": "Str0ng!Passw0rd",
	})

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d, body=%s", rec.Code, rec.Body.String())
	}
	requireMet(t, mock)
}

func TestLoginHandler_WrongPassword(t *testing.T) {
	mock := newMockDB(t)

	correctHash, err := bcrypt.GenerateFromPassword([]byte("Correct!Passw0rd"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	mock.ExpectQuery(`FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "username", "email", "password_hash"}).
			AddRow("user-1", "alice", "alice@example.com", string(correctHash)))

	rec := doLoginRequest(t, http.MethodPost, map[string]string{
		"username":      "alice",
		"password_text": "Wr0ng!Passw0rd",
	})

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d, body=%s", rec.Code, rec.Body.String())
	}
	requireMet(t, mock)
}

func TestLoginHandler_Success(t *testing.T) {
	mock := newMockDB(t)

	password := "Str0ng!Passw0rd"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	mock.ExpectQuery(`FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "username", "email", "password_hash"}).
			AddRow("user-1", "alice", "alice@example.com", string(hash)))

	// auth.GenerateJWT_Token -> models.AddRefreshToken (no existing token -> insert)
	mock.ExpectQuery(`FROM "refresh_tokens" WHERE user_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}))
	mock.ExpectExec(`INSERT INTO "refresh_tokens"`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rec := doLoginRequest(t, http.MethodPost, map[string]string{
		"username":      "alice",
		"password_text": password,
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		AccessToken string    `json:"access_token"`
		ExpiresAt   time.Time `json:"expires_at"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if !resp.ExpiresAt.After(time.Now()) {
		t.Error("expected expires_at to be in the future")
	}
	requireMet(t, mock)
}

func TestLoginHandler_ByEmail(t *testing.T) {
	mock := newMockDB(t)

	password := "Str0ng!Passw0rd"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	mock.ExpectQuery(`FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "username", "email", "password_hash"}).
			AddRow("user-1", "alice", "alice@example.com", string(hash)))
	mock.ExpectQuery(`FROM "refresh_tokens" WHERE user_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}))
	mock.ExpectExec(`INSERT INTO "refresh_tokens"`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rec := doLoginRequest(t, http.MethodPost, map[string]string{
		"email":         "alice@example.com",
		"password_text": password,
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}
	requireMet(t, mock)
}
