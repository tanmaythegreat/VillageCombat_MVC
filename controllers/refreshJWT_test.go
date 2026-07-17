package controllers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"golang.org/x/crypto/bcrypt"
)

func doRefreshRequest(t *testing.T, method string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("failed to encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, "/refresh", &buf)
	rec := httptest.NewRecorder()
	RefreshHandler(rec, req)
	return rec
}

func TestRefreshHandler_MethodNotAllowed(t *testing.T) {
	rec := doRefreshRequest(t, http.MethodGet, nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestRefreshHandler_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBufferString("{bad"))
	rec := httptest.NewRecorder()
	RefreshHandler(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestRefreshHandler_NoStoredToken(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`FROM "refresh_tokens"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}))

	rec := doRefreshRequest(t, http.MethodPost, map[string]string{
		"user_id":       "user-1",
		"refresh_token": "anything",
	})

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d, body=%s", rec.Code, rec.Body.String())
	}
	requireMet(t, mock)
}

func TestRefreshHandler_ExpiredToken(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`FROM "refresh_tokens"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "jwt_token_hash", "ip_address", "user_agent", "expires_at", "created_at", "is_used",
		}).AddRow("rt-1", "user-1", "somehash", "127.0.0.1", "", time.Now().Add(-time.Hour), time.Now(), false))

	rec := doRefreshRequest(t, http.MethodPost, map[string]string{
		"user_id":       "user-1",
		"refresh_token": "anything",
	})

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for expired token, got %d, body=%s", rec.Code, rec.Body.String())
	}
	requireMet(t, mock)
}

func TestRefreshHandler_WrongUserAgent(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`FROM "refresh_tokens"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "jwt_token_hash", "ip_address", "user_agent", "expires_at", "created_at", "is_used",
		}).AddRow("rt-1", "user-1", "somehash", "127.0.0.1", "some-other-agent", time.Now().Add(time.Hour), time.Now(), false))

	rec := doRefreshRequest(t, http.MethodPost, map[string]string{
		"user_id":       "user-1",
		"refresh_token": "anything",
	})

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for user-agent mismatch, got %d, body=%s", rec.Code, rec.Body.String())
	}
	requireMet(t, mock)
}

func TestRefreshHandler_InvalidTokenEncoding(t *testing.T) {
	mock := newMockDB(t)

	mock.ExpectQuery(`FROM "refresh_tokens"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "jwt_token_hash", "ip_address", "user_agent", "expires_at", "created_at", "is_used",
		}).AddRow("rt-1", "user-1", "somehash", "127.0.0.1", "", time.Now().Add(time.Hour), time.Now(), false))

	rec := doRefreshRequest(t, http.MethodPost, map[string]string{
		"user_id":       "user-1",
		"refresh_token": "not-valid-base64!!!",
	})

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for undecodable refresh token, got %d, body=%s", rec.Code, rec.Body.String())
	}
	requireMet(t, mock)
}

func TestRefreshHandler_TokenDoesNotMatchHash(t *testing.T) {
	mock := newMockDB(t)

	storedPlain := []byte("the-real-refresh-token-bytes")
	storedHash, err := bcrypt.GenerateFromPassword(storedPlain, bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash: %v", err)
	}
	wrongPlain := base64.RawURLEncoding.EncodeToString([]byte("a-completely-different-token"))

	mock.ExpectQuery(`FROM "refresh_tokens"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "jwt_token_hash", "ip_address", "user_agent", "expires_at", "created_at", "is_used",
		}).AddRow("rt-1", "user-1", string(storedHash), "127.0.0.1", "", time.Now().Add(time.Hour), time.Now(), false))

	rec := doRefreshRequest(t, http.MethodPost, map[string]string{
		"user_id":       "user-1",
		"refresh_token": wrongPlain,
	})

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for non-matching token, got %d, body=%s", rec.Code, rec.Body.String())
	}
	requireMet(t, mock)
}

func TestRefreshHandler_Success(t *testing.T) {
	mock := newMockDB(t)

	storedPlain := []byte("the-real-refresh-token-bytes")
	storedHash, err := bcrypt.GenerateFromPassword(storedPlain, bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash: %v", err)
	}
	b64Token := base64.RawURLEncoding.EncodeToString(storedPlain)

	// 1) auth.RefreshAccessToken -> models.GetRefreshTokenByUserID
	mock.ExpectQuery(`FROM "refresh_tokens"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "jwt_token_hash", "ip_address", "user_agent", "expires_at", "created_at", "is_used",
		}).AddRow("rt-1", "user-1", string(storedHash), "127.0.0.1", "", time.Now().Add(time.Hour), time.Now(), false))

	// 2) GenerateJWT_Token -> models.AddRefreshToken (existing token found -> update path)
	mock.ExpectQuery(`FROM "refresh_tokens" WHERE user_id`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "jwt_token_hash", "ip_address", "user_agent", "expires_at", "created_at", "is_used",
		}).AddRow("rt-1", "user-1", string(storedHash), "127.0.0.1", "", time.Now().Add(time.Hour), time.Now(), false))
	mock.ExpectExec(`UPDATE "refresh_tokens"`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rec := doRefreshRequest(t, http.MethodPost, map[string]string{
		"user_id":       "user-1",
		"refresh_token": b64Token,
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	requireMet(t, mock)
}
