package migrate

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func Run(db *sql.DB, migrationsDir string) error {
	// Create migrations tracking table
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	// Get applied migrations
	applied := make(map[string]bool)
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("query migrations: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var v string
		rows.Scan(&v)
		applied[v] = true
	}

	// Baseline: if schema_migrations is empty but tables exist, mark existing migrations as applied
	if len(applied) == 0 {
		var exists bool
		db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users')").Scan(&exists)
		if exists {
			// DB was set up before auto-migrate was added; baseline all existing migrations
			baselineFiles, _ := os.ReadDir(migrationsDir)
			for _, e := range baselineFiles {
				if strings.HasSuffix(e.Name(), ".up.sql") {
					v := strings.TrimSuffix(e.Name(), ".up.sql")
					// Only baseline migrations before 0007 (chat is the first new one)
					if v < "0007" {
						db.Exec("INSERT INTO schema_migrations (version) VALUES ($1) ON CONFLICT DO NOTHING", v)
						applied[v] = true
						log.Printf("migration baselined: %s", e.Name())
					}
				}
			}
		}
	}

	// Find .up.sql files
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var upFiles []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".up.sql") {
			upFiles = append(upFiles, e.Name())
		}
	}
	sort.Strings(upFiles)

	// Apply pending
	for _, f := range upFiles {
		version := strings.TrimSuffix(f, ".up.sql")
		if applied[version] {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, f))
		if err != nil {
			return fmt.Errorf("read %s: %w", f, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin tx for %s: %w", f, err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("execute %s: %w", f, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			tx.Rollback()
			return fmt.Errorf("record %s: %w", f, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit %s: %w", f, err)
		}

		log.Printf("migration applied: %s", f)
	}

	return nil
}
