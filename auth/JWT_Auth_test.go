package auth

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"Village_combat/models"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	_, thisFile, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(thisFile), "..")
	_ = godotenv.Load(filepath.Join(root, ".env"))
	os.Exit(m.Run())
}

// withMockModelsDB points the package-level models.DB at a sqlmock-backed
// gorm.DB for the duration of the test, and restores the original value
// afterwards.
func withMockModelsDB(t *testing.T) sqlmock.Sqlmock {
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

// expectAddRefreshToken queues the two DB calls that models.AddRefreshToken
// makes on the "no existing token for this user" path: a SELECT that comes
// back empty, followed by a plain INSERT (no RETURNING clause is used in
// AddRefreshToken, so this is Exec, not Query).
func expectAddRefreshToken(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`FROM "refresh_tokens" WHERE user_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}))
	mock.ExpectExec(`INSERT INTO "refresh_tokens"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestGenerateAndVerifyJWT_Token_Success(t *testing.T) {
	mock := withMockModelsDB(t)
	expectAddRefreshToken(mock)

	userID := "user-123"
	ip := "127.0.0.1"
	ua := "test-agent"
	token, err := GenerateJWT_Token(userID, ip, ua)
	if err != nil {
		t.Fatalf("GenerateJWT_Token returned unexpected error: %v", err)
	}
	if token.AccessToken == "" {
		t.Fatal("expected non-empty AccessToken")
	}
	gotUserID, verified := VerifyJWT_Token(token.AccessToken)
	if !verified {
		t.Fatal("expected token to verify successfully")
	}
	if gotUserID != userID {
		t.Fatalf("expected userID %q, got %q", userID, gotUserID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

func TestVerifyJWT_Token_RejectsGarbage(t *testing.T) {
	_, verified := VerifyJWT_Token("this.is.not.a.valid.jwt")
	if verified {
		t.Fatal("expected garbage token to fail verification")
	}
}

func TestVerifyJWT_Token_RejectsEmptyString(t *testing.T) {
	_, verified := VerifyJWT_Token("")
	if verified {
		t.Fatal("expected empty token string to fail verification")
	}
}

func TestVerifyJWT_Token_RejectsTamperedToken(t *testing.T) {
	mock := withMockModelsDB(t)
	expectAddRefreshToken(mock)

	token, err := GenerateJWT_Token("user-456", "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("GenerateJWT_Token returned unexpected error: %v", err)
	}
	// Flip a character near the end of the signature to invalidate it
	tampered := token.AccessToken[:len(token.AccessToken)-1] + "x"
	_, verified := VerifyJWT_Token(tampered)
	if verified {
		t.Fatal("expected tampered token to fail verification")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

func TestCreateAccessToken_ExpiryIsRespected(t *testing.T) {
	duration := 2 * time.Second
	tokenStr, expiresAt, err := createAccessToken("user-789", duration)
	if err != nil {
		t.Fatalf("createAccessToken returned unexpected error: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("expected non-empty token string")
	}
	if time.Until(expiresAt) > duration || time.Until(expiresAt) <= 0 {
		t.Fatalf("expiresAt %v not within expected window for duration %v", expiresAt, duration)
	}
	// Token should verify immediately
	_, verified := VerifyJWT_Token(tokenStr)
	if !verified {
		t.Fatal("expected freshly created token to verify")
	}
}

func TestValidatePasswordStrength_TableDriven(t *testing.T) {
	cases := []struct {
		name      string
		password  string
		expectErr bool
	}{
		{"too short", "Ab1!", true},
		{"no uppercase", "lowercase123!", true},
		{"no lowercase", "UPPERCASE123!", true},
		{"no digit", "NoDigitsHere!", true},
		{"no special char", "NoSpecialChar123", true},
		{"empty string", "", true},
		{"strong password", "Str0ng!Passw0rd", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tc.password)
			if tc.expectErr && err == nil {
				t.Errorf("expected error for password %q, got nil", tc.password)
			}
			if !tc.expectErr && err != nil {
				t.Errorf("expected no error for password %q, got: %v", tc.password, err)
			}
		})
	}
}
