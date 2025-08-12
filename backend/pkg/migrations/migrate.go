package migrations

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
	"time"
)

type Migration struct {
	Version     string
	Description string
	Up          func(*sql.DB) error
	Down        func(*sql.DB) error
}

type MigrationRunner struct {
	db         *sql.DB
	migrations []Migration
}

func NewMigrationRunner(db *sql.DB) *MigrationRunner {
	return &MigrationRunner{
		db:         db,
		migrations: []Migration{},
	}
}

func (mr *MigrationRunner) AddMigration(migration Migration) {
	mr.migrations = append(mr.migrations, migration)
}

func (mr *MigrationRunner) createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			version VARCHAR(255) PRIMARY KEY,
			description TEXT,
			executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err := mr.db.Exec(query)
	return err
}

func (mr *MigrationRunner) isExecuted(version string) (bool, error) {
	var count int
	err := mr.db.QueryRow("SELECT COUNT(*) FROM migrations WHERE version = $1", version).Scan(&count)
	return count > 0, err
}

func (mr *MigrationRunner) markAsExecuted(version, description string) error {
	_, err := mr.db.Exec(
		"INSERT INTO migrations (version, description, executed_at) VALUES ($1, $2, $3)",
		version, description, time.Now(),
	)
	return err
}

func (mr *MigrationRunner) Run() error {
	// Create migrations table if it doesn't exist
	if err := mr.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %v", err)
	}

	// Sort migrations by version
	sort.Slice(mr.migrations, func(i, j int) bool {
		return mr.migrations[i].Version < mr.migrations[j].Version
	})

	// Execute pending migrations
	for _, migration := range mr.migrations {
		executed, err := mr.isExecuted(migration.Version)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %v", err)
		}

		if !executed {
			log.Printf("Running migration %s: %s", migration.Version, migration.Description)

			// Start transaction
			tx, err := mr.db.Begin()
			if err != nil {
				return fmt.Errorf("failed to start transaction: %v", err)
			}

			// Execute migration
			if err := migration.Up(mr.db); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to execute migration %s: %v", migration.Version, err)
			}

			// Mark as executed
			if err := mr.markAsExecuted(migration.Version, migration.Description); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to mark migration as executed: %v", err)
			}

			// Commit transaction
			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit migration: %v", err)
			}

			log.Printf("Migration %s completed successfully", migration.Version)
		}
	}

	return nil
}

func (mr *MigrationRunner) Rollback(targetVersion string) error {
	// Sort migrations by version in descending order
	sort.Slice(mr.migrations, func(i, j int) bool {
		return mr.migrations[i].Version > mr.migrations[j].Version
	})

	for _, migration := range mr.migrations {
		if migration.Version <= targetVersion {
			break
		}

		executed, err := mr.isExecuted(migration.Version)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %v", err)
		}

		if executed {
			log.Printf("Rolling back migration %s: %s", migration.Version, migration.Description)

			if err := migration.Down(mr.db); err != nil {
				return fmt.Errorf("failed to rollback migration %s: %v", migration.Version, err)
			}

			// Remove from migrations table
			_, err = mr.db.Exec("DELETE FROM migrations WHERE version = $1", migration.Version)
			if err != nil {
				return fmt.Errorf("failed to remove migration record: %v", err)
			}

			log.Printf("Migration %s rolled back successfully", migration.Version)
		}
	}

	return nil
}
