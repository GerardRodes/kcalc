package ksqlite

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

// PRAGMA user_version;

//go:embed sql/migrations/*.sql
var migrations embed.FS

func RunMigrations() (outErr error) {
	wl.Lock()
	defer wl.Unlock()

	var schemaVersion int
	{
		stmt, err := w.conn.Prepare("PRAGMA user_version;")
		if err != nil {
			return fmt.Errorf("prepare stmt: %w", err)
		}
		defer stmt.Close()
		hasRow, err := stmt.Step()
		if err != nil {
			return fmt.Errorf("stmt step: %w", err)
		}
		if !hasRow {
			return errors.New("missing user_version")
		}
		if err := stmt.Scan(&schemaVersion); err != nil {
			return fmt.Errorf("scan schema version: %w", err)
		}
	}
	log.Debug().Int("version", schemaVersion).Msg("read schema version")

	entries, err := migrations.ReadDir("sql/migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	for _, entry := range entries {
		versionStr := strings.SplitN(entry.Name(), "_", 2)[0]
		version, err := strconv.Atoi(versionStr)
		if err != nil {
			return fmt.Errorf("parsing migration version: %w", err)
		}

		if version <= schemaVersion {
			continue
		}

		f, err := migrations.Open(filepath.Join("sql/migrations", entry.Name()))
		if err != nil {
			return fmt.Errorf("open migration: %w", err)
		}

		data, err := io.ReadAll(f)
		if err != nil {
			return fmt.Errorf("read migration: %w", err)
		}

		sql := fmt.Sprintf("BEGIN TRANSACTION;%s;PRAGMA user_version=%d;COMMIT;", string(data), version)
		if err := w.conn.Exec(sql); err != nil {
			return fmt.Errorf("apply migration: %w", err)
		}

		log.Debug().Str("migration", entry.Name()).Msg("applied migration")
	}

	return nil
}
