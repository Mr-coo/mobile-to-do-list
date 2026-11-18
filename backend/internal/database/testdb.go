package database

import (
	"os"
	"testing"

	"gorm.io/gorm"
)

// NewTestDB returns a migrated database with an empty tasks table, for integration tests.
func NewTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("TEST_MYSQL_DSN not set; see README")
	}
	db, err := Open(dsn)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := Migrate(db); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	if err := db.Exec("TRUNCATE TABLE tasks").Error; err != nil {
		t.Fatalf("truncate tasks: %v", err)
	}
	return db
}
