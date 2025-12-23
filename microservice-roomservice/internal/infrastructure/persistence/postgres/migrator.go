package postgres

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"microservice-roomservice/pkg/logger"

	_ "github.com/lib/pq"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// RunMigrations executes all SQL migration files in order
// This is a best-effort approach that relies on SQL files being idempotent (using IF NOT EXISTS)
func RunMigrations(db *sql.DB) error {
	logger.Info().Msg("Running database migrations...")

	// Read all migration files
	migrationDir := "migrations"
	files, err := fs.ReadDir(migrationFiles, migrationDir)
	if err != nil {
		return fmt.Errorf("failed to read migration directory: %w", err)
	}

	// Get and sort SQL files alphabetically
	var sqlFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, file.Name())
		}
	}
	sort.Strings(sqlFiles)

	if len(sqlFiles) == 0 {
		logger.Info().Msg("No migration files found")
		return nil
	}

	// Execute each migration file
	for _, fileName := range sqlFiles {
		if err := executeMigrationFile(db, migrationFiles, migrationDir, fileName); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", fileName, err)
		}
		logger.Info().Msgf("Successfully applied migration: %s", fileName)
	}

	logger.Info().Msgf("Successfully applied %d migrations", len(sqlFiles))
	return nil
}

// executeMigrationFile reads and executes a single SQL migration file
func executeMigrationFile(db *sql.DB, fsys fs.FS, dir, fileName string) error {
	// Read the migration file content
	filePath := path.Join(dir, fileName)
	content, err := fs.ReadFile(fsys, filePath)
	if err != nil {
		return fmt.Errorf("failed to read migration file %s: %w", fileName, err)
	}

	// Execute the migration
	query := string(content)
	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to execute SQL from %s: %w", fileName, err)
	}

	return nil
}
