package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"npan/internal/models"
)

type JSONCheckpointStore struct {
	filePath string
	mu       sync.Mutex
}

func NewJSONCheckpointStore(filePath string) *JSONCheckpointStore {
	return &JSONCheckpointStore{filePath: filePath}
}

func (s *JSONCheckpointStore) Load() (*models.CrawlCheckpoint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var checkpoint models.CrawlCheckpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, err
	}
	return &checkpoint, nil
}

func (s *JSONCheckpointStore) Save(checkpoint *models.CrawlCheckpoint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(s.filePath), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0o644)
}

func (s *JSONCheckpointStore) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.Remove(s.filePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

type JSONProgressStore struct {
	filePath string
	mu       sync.Mutex
}

func NewJSONProgressStore(filePath string) *JSONProgressStore {
	return &JSONProgressStore{filePath: filePath}
}

func (s *JSONProgressStore) Load() (*models.SyncProgressState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var state models.SyncProgressState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	if state.RootProgress == nil {
		state.RootProgress = map[string]*models.RootSyncProgress{}
	}

	return &state, nil
}

func (s *JSONProgressStore) Save(state *models.SyncProgressState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(s.filePath), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0o644)
}

type JSONSyncStateStore struct {
	filePath string
	mu       sync.Mutex
}

func NewJSONSyncStateStore(filePath string) *JSONSyncStateStore {
	return &JSONSyncStateStore{filePath: filePath}
}

func (s *JSONSyncStateStore) Load() (*models.SyncState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var state models.SyncState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func (s *JSONSyncStateStore) Save(state *models.SyncState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(s.filePath), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0o644)
}
