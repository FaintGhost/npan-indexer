package search

import (
	"fmt"
	"strings"
)

type Backend string

const (
	BackendMeilisearch Backend = "meilisearch"
	BackendTypesense   Backend = "typesense"
)

type BackendConfig struct {
	Backend             string
	MeiliHost           string
	MeiliAPIKey         string
	MeiliIndex          string
	TypesenseHost       string
	TypesenseAPIKey     string
	TypesenseCollection string
}

type BackendInfo struct {
	Backend Backend
	Host    string
	Index   string
}

func ParseBackend(raw string) (Backend, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "meili", string(BackendMeilisearch):
		return BackendMeilisearch, nil
	case string(BackendTypesense):
		return BackendTypesense, nil
	default:
		return "", fmt.Errorf("不支持的搜索后端: %s（可选: meilisearch|typesense）", raw)
	}
}

func SupportsPublicInstantsearch(backend Backend) bool {
	return backend == BackendMeilisearch
}

func NewIndexOperator(cfg BackendConfig) (IndexOperator, BackendInfo, error) {
	backend, err := ParseBackend(cfg.Backend)
	if err != nil {
		return nil, BackendInfo{}, err
	}

	switch backend {
	case BackendMeilisearch:
		return NewMeiliIndex(cfg.MeiliHost, cfg.MeiliAPIKey, cfg.MeiliIndex), BackendInfo{
			Backend: backend,
			Host:    cfg.MeiliHost,
			Index:   cfg.MeiliIndex,
		}, nil
	case BackendTypesense:
		return NewTypesenseIndex(cfg.TypesenseHost, cfg.TypesenseAPIKey, cfg.TypesenseCollection), BackendInfo{
			Backend: backend,
			Host:    cfg.TypesenseHost,
			Index:   cfg.TypesenseCollection,
		}, nil
	default:
		return nil, BackendInfo{}, fmt.Errorf("未实现的搜索后端: %s", backend)
	}
}
