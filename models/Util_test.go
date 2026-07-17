package models

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// newMockDB opens a sqlmock-backed *gorm.DB, installs it as the package-level
// DB for the duration of the test, and returns the sqlmock handle plus a
// cleanup func (also registered via t.Cleanup, so calling it manually is
// optional).
func newMockDB(t *testing.T) sqlmock.Sqlmock {
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

	oldDB := DB
	DB = gdb

	t.Cleanup(func() {
		DB = oldDB
		sqlDB.Close()
	})

	return mock
}

// requireMet fails the test with full detail if any mocked expectation
// wasn't satisfied (missing calls, calls in the wrong order, etc).
func requireMet(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}
