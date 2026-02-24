// Package sqlite implements the official SQLite metadata backend for Plinth.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/pointlet/plinth/internal/core/storage"
	"github.com/pointlet/plinth/internal/core/storage/metadata/sqlite/sqlcgen"
)

// Compile-time assertions to ensure Store implements the expected interfaces.
var (
	_ storage.ObjectStore = (*Store)(nil)
	_ storage.UploadStore = (*Store)(nil)
)

// Store is the official SQLite metadata backend.
// Actual query logic is intentionally left as boilerplate for now.
type Store struct {
	database         *sql.DB
	generatedQueries *sqlcgen.Queries
}

// JournalMode defines the SQLite journal mode PRAGMA value.
type JournalMode string

const (
	JournalModeWAL JournalMode = "WAL"
)

// SyncMode defines the SQLite synchronous PRAGMA value.
type SyncMode string

const (
	SyncModeNormal SyncMode = "NORMAL"
	SyncModeFull   SyncMode = "FULL"
)

// RuntimeConfig defines SQLite runtime settings shared by all store operations.
type RuntimeConfig struct {
	JournalMode JournalMode
	SyncMode    SyncMode
	BusyTimeout time.Duration
	ForeignKeys bool
}

// DefaultRuntimeConfig returns production-oriented defaults for Plinth.
func DefaultRuntimeConfig() RuntimeConfig {
	return RuntimeConfig{
		JournalMode: JournalModeWAL,
		SyncMode:    SyncModeNormal,
		BusyTimeout: 5 * time.Second,
		ForeignKeys: true,
	}
}

// New wires a SQLite-backed object/upload store around an existing database handle.
func New(database *sql.DB) *Store {
	return &Store{
		database:         database,
		generatedQueries: sqlcgen.New(database),
	}
}

// Open opens a SQLite database and returns the metadata backend.
// Caller is responsible for importing/registering a SQLite driver and closing the database.
func Open(dataSourceName string) (*Store, error) {
	return OpenWithConfig(context.Background(), dataSourceName, DefaultRuntimeConfig())
}

// OpenWithConfig opens SQLite, applies runtime PRAGMA settings, and returns the store.
func OpenWithConfig(requestContext context.Context, dataSourceName string, runtimeConfig RuntimeConfig) (*Store, error) {
	database, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err := configureDatabase(requestContext, database, runtimeConfig); err != nil {
		_ = database.Close()
		return nil, err
	}

	return &Store{
		database:         database,
		generatedQueries: sqlcgen.New(database),
	}, nil
}

func configureDatabase(requestContext context.Context, database *sql.DB, runtimeConfig RuntimeConfig) error {
	if runtimeConfig.BusyTimeout < 0 {
		return fmt.Errorf("sqlite: busy timeout must be >= 0")
	}
	if runtimeConfig.JournalMode == "" {
		runtimeConfig.JournalMode = JournalModeWAL
	}
	if runtimeConfig.SyncMode == "" {
		runtimeConfig.SyncMode = SyncModeNormal
	}

	// Intentionally serialized to avoid concurrency pitfalls and maximize stability.
	// I will optimize I/O and query performance, not connection parallelism to
	// metadata store until it becomes a bottleneck.
	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)
	database.SetConnMaxLifetime(0)
	database.SetConnMaxIdleTime(0)

	pragmaQueries := []string{
		"PRAGMA journal_mode = " + string(runtimeConfig.JournalMode) + ";",
		"PRAGMA synchronous = " + string(runtimeConfig.SyncMode) + ";",
		"PRAGMA foreign_keys = " + boolToPragma(runtimeConfig.ForeignKeys) + ";",
		"PRAGMA busy_timeout = " + strconv.FormatInt(runtimeConfig.BusyTimeout.Milliseconds(), 10) + ";",
	}

	for _, pragmaQuery := range pragmaQueries {
		if _, err := database.ExecContext(requestContext, pragmaQuery); err != nil {
			return err
		}
	}

	return database.PingContext(requestContext)
}

func boolToPragma(isTrue bool) string {
	if isTrue {
		return "ON"
	}

	return "OFF"
}
