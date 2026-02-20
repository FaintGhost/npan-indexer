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

	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return err
	}

	return writeFileAtomic(s.filePath, data, 0o644)
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

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return writeFileAtomic(s.filePath, data, 0o644)
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

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return writeFileAtomic(s.filePath, data, 0o644)
}
