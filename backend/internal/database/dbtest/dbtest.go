// Package dbtest is separate from database so the API binary does not link the testing package.
package dbtest

import (
	"os"
	"testing"

	"gorm.io/gorm"

	"taskmanager/internal/database"
)

// New returns a migrated database with an empty tasks table, for integration tests.
func New(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("TEST_MYSQL_DSN not set; see README")
	}
	db, err := database.Open(dsn)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	if err := db.Exec("TRUNCATE TABLE tasks").Error; err != nil {
		t.Fatalf("truncate tasks: %v", err)
	}
	return db
}
