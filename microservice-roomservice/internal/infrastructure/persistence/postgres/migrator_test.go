package postgres

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func TestRunMigrations(t *testing.T) {
	// Skip if no test database is available
	dbDSN := os.Getenv("TEST_DB_DSN")
	if dbDSN == "" {
		t.Skip("TEST_DB_DSN not set, skipping migration test")
	}

	// Connect to test database
	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	// Test migrations
	err = RunMigrations(db)
	if err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	// Verify tables exist
	var exists bool
	err = db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'menu_items')").Scan(&exists)
	if err != nil {
		t.Fatalf("Failed to check menu_items table: %v", err)
	}
	if !exists {
		t.Error("menu_items table was not created")
	}

	err = db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'orders')").Scan(&exists)
	if err != nil {
		t.Fatalf("Failed to check orders table: %v", err)
	}
	if !exists {
		t.Error("orders table was not created")
	}

	// Test idempotency by running migrations again
	err = RunMigrations(db)
	if err != nil {
		t.Fatalf("RunMigrations failed on second run: %v", err)
	}
}
