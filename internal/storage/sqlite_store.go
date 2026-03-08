package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"npan/internal/models"
)

const (
	sqliteDriverName         = "sqlite"
	stateNamespaceProgress   = "progress"
	stateNamespaceSyncState  = "sync_state"
	stateNamespaceCheckpoint = "checkpoint"
	stateDefaultKey          = "default"
)

type SQLiteStateStoresConfig struct {
	StateDBFile         string
	LegacyProgressFile  string
	LegacySyncStateFile string
}

type SQLiteStateStores struct {
	DB                     *sql.DB
	ProgressStore          ProgressStore
	SyncStateStore         SyncStateStore
	CheckpointStoreFactory CheckpointStoreFactory
}

type sqliteStateStore struct {
	db *sql.DB
}

type SQLiteProgressStore struct {
	stateStore *sqliteStateStore
	legacyFile string
}

type SQLiteSyncStateStore struct {
	stateStore *sqliteStateStore
	legacyFile string
}

type SQLiteCheckpointStoreFactory struct {
	stateStore *sqliteStateStore
}

type SQLiteCheckpointStore struct {
	stateStore *sqliteStateStore
	key        string
}

func NewSQLiteStateStores(cfg SQLiteStateStoresConfig) (*SQLiteStateStores, error) {
	if cfg.StateDBFile == "" {
		return nil, fmt.Errorf("state db file is required")
	}
	if err := os.MkdirAll(filepath.Dir(cfg.StateDBFile), 0o755); err != nil {
		return nil, err
	}

	db, err := sql.Open(sqliteDriverName, cfg.StateDBFile)
	if err != nil {
		return nil, err
	}
	if err := configureSQLiteDB(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	stateStore := &sqliteStateStore{db: db}
	return &SQLiteStateStores{
		DB: db,
		ProgressStore: &SQLiteProgressStore{
			stateStore: stateStore,
			legacyFile: cfg.LegacyProgressFile,
		},
		SyncStateStore: &SQLiteSyncStateStore{
			stateStore: stateStore,
			legacyFile: cfg.LegacySyncStateFile,
		},
		CheckpointStoreFactory: &SQLiteCheckpointStoreFactory{stateStore: stateStore},
	}, nil
}

func configureSQLiteDB(db *sql.DB) error {
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA synchronous=FULL;",
		"PRAGMA busy_timeout=5000;",
	}
	for _, stmt := range pragmas {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}

	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS state_entries (
  namespace TEXT NOT NULL,
  key TEXT NOT NULL,
  payload_json TEXT NOT NULL,
  updated_at_ms INTEGER NOT NULL,
  PRIMARY KEY(namespace, key)
)`)
	return err
}

func (s *sqliteStateStore) loadEntry(namespace string, key string) ([]byte, bool, error) {
	var payload string
	err := s.db.QueryRow(
		`SELECT payload_json FROM state_entries WHERE namespace = ? AND key = ?`,
		namespace,
		key,
	).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return []byte(payload), true, nil
}

func (s *sqliteStateStore) upsertEntry(namespace string, key string, payload []byte) error {
	_, err := s.db.Exec(
		`INSERT INTO state_entries(namespace, key, payload_json, updated_at_ms)
VALUES (?, ?, ?, ?)
ON CONFLICT(namespace, key) DO UPDATE SET
  payload_json = excluded.payload_json,
  updated_at_ms = excluded.updated_at_ms`,
		namespace,
		key,
		string(payload),
		time.Now().UnixMilli(),
	)
	return err
}

func (s *sqliteStateStore) insertEntryIfAbsent(namespace string, key string, payload []byte) error {
	_, err := s.db.Exec(
		`INSERT INTO state_entries(namespace, key, payload_json, updated_at_ms)
VALUES (?, ?, ?, ?)
ON CONFLICT(namespace, key) DO NOTHING`,
		namespace,
		key,
		string(payload),
		time.Now().UnixMilli(),
	)
	return err
}

func (s *sqliteStateStore) deleteEntry(namespace string, key string) error {
	_, err := s.db.Exec(`DELETE FROM state_entries WHERE namespace = ? AND key = ?`, namespace, key)
	return err
}

func (s *SQLiteProgressStore) Load() (*models.SyncProgressState, error) {
	state, err := loadStateWithFallback(
		s.stateStore,
		stateNamespaceProgress,
		stateDefaultKey,
		s.legacyFile,
		func(filePath string) (*models.SyncProgressState, error) {
			return NewJSONProgressStore(filePath).Load()
		},
	)
	if err != nil {
		return nil, err
	}
	return normalizeSyncProgressState(state), nil
}

func (s *SQLiteProgressStore) Save(state *models.SyncProgressState) error {
	return saveStateEntry(s.stateStore, stateNamespaceProgress, stateDefaultKey, state)
}

func (s *SQLiteSyncStateStore) Load() (*models.SyncState, error) {
	return loadStateWithFallback(
		s.stateStore,
		stateNamespaceSyncState,
		stateDefaultKey,
		s.legacyFile,
		func(filePath string) (*models.SyncState, error) {
			return NewJSONSyncStateStore(filePath).Load()
		},
	)
}

func (s *SQLiteSyncStateStore) Save(state *models.SyncState) error {
	return saveStateEntry(s.stateStore, stateNamespaceSyncState, stateDefaultKey, state)
}

func (f *SQLiteCheckpointStoreFactory) ForKey(key string) CheckpointStore {
	return &SQLiteCheckpointStore{stateStore: f.stateStore, key: key}
}

func (s *SQLiteCheckpointStore) Load() (*models.CrawlCheckpoint, error) {
	return loadStateWithFallback(
		s.stateStore,
		stateNamespaceCheckpoint,
		s.key,
		s.key,
		func(filePath string) (*models.CrawlCheckpoint, error) {
			return NewJSONCheckpointStore(filePath).Load()
		},
	)
}

func (s *SQLiteCheckpointStore) Save(checkpoint *models.CrawlCheckpoint) error {
	return saveStateEntry(s.stateStore, stateNamespaceCheckpoint, s.key, checkpoint)
}

func (s *SQLiteCheckpointStore) Clear() error {
	return s.stateStore.deleteEntry(stateNamespaceCheckpoint, s.key)
}

func loadStateWithFallback[T any](stateStore *sqliteStateStore, namespace string, key string, legacyFile string, loadLegacy func(string) (*T, error)) (*T, error) {
	value, ok, err := loadStateEntry[T](stateStore, namespace, key)
	if err != nil {
		return nil, err
	}
	if ok {
		return value, nil
	}
	if legacyFile == "" {
		return nil, nil
	}

	legacyValue, err := loadLegacy(legacyFile)
	if err != nil {
		return nil, err
	}
	if legacyValue == nil {
		return nil, nil
	}
	payload, err := json.Marshal(legacyValue)
	if err != nil {
		return nil, err
	}
	if err := stateStore.insertEntryIfAbsent(namespace, key, payload); err != nil {
		return nil, err
	}
	return loadStateAfterMigration[T](stateStore, namespace, key)
}

func loadStateEntry[T any](stateStore *sqliteStateStore, namespace string, key string) (*T, bool, error) {
	payload, ok, err := stateStore.loadEntry(namespace, key)
	if err != nil || !ok {
		return nil, ok, err
	}

	var value T
	if err := json.Unmarshal(payload, &value); err != nil {
		return nil, false, err
	}
	return &value, true, nil
}

func loadStateAfterMigration[T any](stateStore *sqliteStateStore, namespace string, key string) (*T, error) {
	value, ok, err := loadStateEntry[T](stateStore, namespace, key)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return value, nil
}

func saveStateEntry(stateStore *sqliteStateStore, namespace string, key string, value any) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return stateStore.upsertEntry(namespace, key, payload)
}
