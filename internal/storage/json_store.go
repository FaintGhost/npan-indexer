package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"npan/internal/models"
)

type ProgressStore interface {
	Load() (*models.SyncProgressState, error)
	Save(state *models.SyncProgressState) error
}

type SyncStateStore interface {
	Load() (*models.SyncState, error)
	Save(state *models.SyncState) error
}

type CheckpointStore interface {
	Load() (*models.CrawlCheckpoint, error)
	Save(checkpoint *models.CrawlCheckpoint) error
	Clear() error
}

type CheckpointStoreFactory interface {
	ForKey(key string) CheckpointStore
}

type JSONCheckpointStore struct {
	filePath string
	mu       sync.Mutex
}

func writeFileAtomic(filePath string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(dir, filepath.Base(filePath)+".tmp-*")
	if err != nil {
		return err
	}

	tmpPath := tmpFile.Name()
	keepTmp := true
	defer func() {
		if keepTmp {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return err
	}

	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		return err
	}

	if err := tmpFile.Chmod(perm); err != nil {
		_ = tmpFile.Close()
		return err
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, filePath); err != nil {
		return err
	}
	keepTmp = false

	// 尽量把目录项也刷盘，降低掉电后 rename 丢失风险。
	dirHandle, err := os.Open(dir)
	if err != nil {
		return nil
	}
	defer dirHandle.Close()

	if err := dirHandle.Sync(); err != nil {
		return fmt.Errorf("同步目录元数据失败: %w", err)
	}
	return nil
}

func loadJSONFile[T any](filePath string) (*T, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, err
	}
	return &value, nil
}

func saveJSONFile(filePath string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(filePath, data, 0o644)
}

func normalizeSyncProgressState(state *models.SyncProgressState) *models.SyncProgressState {
	if state == nil {
		return nil
	}
	if state.RootProgress == nil {
		state.RootProgress = map[string]*models.RootSyncProgress{}
	}
	return state
}

func NewJSONCheckpointStore(filePath string) *JSONCheckpointStore {
	return &JSONCheckpointStore{filePath: filePath}
}

func (s *JSONCheckpointStore) Load() (*models.CrawlCheckpoint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return loadJSONFile[models.CrawlCheckpoint](s.filePath)
}

func (s *JSONCheckpointStore) Save(checkpoint *models.CrawlCheckpoint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return saveJSONFile(s.filePath, checkpoint)
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

	state, err := loadJSONFile[models.SyncProgressState](s.filePath)
	if err != nil {
		return nil, err
	}
	return normalizeSyncProgressState(state), nil
}

func (s *JSONProgressStore) Save(state *models.SyncProgressState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return saveJSONFile(s.filePath, state)
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

	return loadJSONFile[models.SyncState](s.filePath)
}

func (s *JSONSyncStateStore) Save(state *models.SyncState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return saveJSONFile(s.filePath, state)
}
