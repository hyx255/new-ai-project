package database

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.uber.org/zap"
)

// RunMigrations reads and executes SQL files from the given directory.
// Files are executed in alphabetical order (e.g., 001_xxx.sql, 002_xxx.sql).
func (db *DB) RunMigrations(migrationsDir string) error {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations directory: %w", err)
	}

	// Collect SQL files
	var sqlFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			sqlFiles = append(sqlFiles, entry.Name())
		}
	}
	sort.Strings(sqlFiles)

	if len(sqlFiles) == 0 {
		db.logger.Info("no migration files found", zap.String("dir", migrationsDir))
		return nil
	}

	for _, name := range sqlFiles {
		path := filepath.Join(migrationsDir, name)
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration file %s: %w", name, err)
		}

		// Split by semicolons for multi-statement support
		statements := splitStatements(string(content))
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := db.Exec(stmt); err != nil {
				return fmt.Errorf("execute migration %s: %w (statement: %.100s)", name, err, stmt)
			}
		}

		db.logger.Info("migration applied", zap.String("file", name))
	}

	return nil
}

// splitStatements splits SQL content by semicolons, ignoring those inside comments.
func splitStatements(sql string) []string {
	var statements []string
	lines := strings.Split(sql, "\n")
	var current strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip pure comment lines
		if strings.HasPrefix(trimmed, "--") || trimmed == "" {
			continue
		}
		current.WriteString(line)
		current.WriteString("\n")
		if strings.HasSuffix(trimmed, ";") {
			statements = append(statements, current.String())
			current.Reset()
		}
	}

	if remaining := strings.TrimSpace(current.String()); remaining != "" {
		statements = append(statements, remaining)
	}

	return statements
}
