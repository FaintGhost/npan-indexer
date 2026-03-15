package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"npan/internal/models"
	"npan/internal/npan"
	"npan/internal/search"
)

type Config struct {
	ServerAddr              string
	ServerReadHeaderTimeout time.Duration
	ServerReadTimeout       time.Duration
	ServerWriteTimeout      time.Duration
	ServerIdleTimeout       time.Duration
	MetricsAddr             string

	AdminAPIKey             string
	AllowConfigAuthFallback bool

	BaseURL   string
	OAuthHost string

	Token        string
	ClientID     string
	ClientSecret string
	SubID        int64
	SubType      npan.TokenSubjectType

	SearchBackend               string
	MeiliHost                   string
	MeiliAPIKey                 string
	MeiliIndex                  string
	TypesenseHost               string
	TypesenseAPIKey             string
	TypesenseCollection         string
	PublicSearchHost            string
	PublicSearchIndexName       string
	PublicSearchAPIKey          string
	TypesensePublicSearchHost   string
	TypesensePublicSearchIndex  string
	TypesensePublicSearchAPIKey string
	PublicSearchInstantsearchOn bool

	StateDBFile         string
	CheckpointTemplate  string
	ProgressFile        string
	SyncStateFile       string
	IncrementalQuery    string
	SyncWindowOverlapMS int64

	DefaultIncludeDepartments bool
	DefaultRootFolderIDs      []int64
	DefaultDepartmentIDs      []int64

	SyncMaxConcurrent            int
	SyncMinTimeMS                int
	SyncRootWorkers              int
	SyncProgressEvery            int
	InspectRootsMaxConcurrency   int
	InspectRootsPerFolderTimeout time.Duration

	Retry models.RetryPolicyOptions
}

func readString(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func readInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		slog.Warn("环境变量格式错误，使用默认值", "key", key, "value", raw, "fallback", fallback)
		return fallback
	}
	return parsed
}

func readInt64(key string, fallback int64) int64 {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		slog.Warn("环境变量格式错误，使用默认值", "key", key, "value", raw, "fallback", fallback)
		return fallback
	}
	return parsed
}

func readBool(key string, fallback bool) bool {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if raw == "" {
		return fallback
	}
	return raw == "1" || raw == "true" || raw == "yes"
}

func readDuration(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil {
		slog.Warn("环境变量格式错误，使用默认值", "key", key, "value", raw, "fallback", fallback)
		return fallback
	}
	return parsed
}

func readInt64List(key string) []int64 {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	result := make([]int64, 0, len(parts))
	for _, item := range parts {
		parsed, err := strconv.ParseInt(strings.TrimSpace(item), 10, 64)
		if err != nil {
			continue
		}
		result = append(result, parsed)
	}
	return result
}

func Load() Config {
	loadDotEnv()

	subType := npan.TokenSubjectType(readString("NPA_SUB_TYPE", "user"))
	if subType == "" {
		subType = npan.TokenSubjectUser
	}

	rootIDs := readInt64List("NPA_ROOT_FOLDER_IDS")
	if len(rootIDs) == 0 {
		rootIDs = []int64{0}
	}

	return Config{
		ServerAddr:              readString("SERVER_ADDR", ":1323"),
		ServerReadHeaderTimeout: readDuration("SERVER_READ_HEADER_TIMEOUT", 5*time.Second),
		ServerReadTimeout:       readDuration("SERVER_READ_TIMEOUT", 10*time.Second),
		ServerWriteTimeout:      readDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
		ServerIdleTimeout:       readDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
		MetricsAddr:             readString("METRICS_ADDR", ":9091"),
		AdminAPIKey:             readString("NPA_ADMIN_API_KEY", ""),
		AllowConfigAuthFallback: readBool("NPA_ALLOW_CONFIG_AUTH_FALLBACK", false),

		BaseURL:   readString("NPA_BASE_URL", "https://npan.novastar.tech:6001/openapi"),
		OAuthHost: readString("NPA_OAUTH_HOST", npan.DefaultOAuthHost),

		Token:        readString("NPA_TOKEN", ""),
		ClientID:     readString("NPA_CLIENT_ID", ""),
		ClientSecret: readString("NPA_CLIENT_SECRET", ""),
		SubID:        readInt64("NPA_SUB_ID", 0),
		SubType:      subType,

		SearchBackend:               readString("NPA_SEARCH_BACKEND", string(search.BackendMeilisearch)),
		MeiliHost:                   readString("MEILI_HOST", "http://127.0.0.1:7700"),
		MeiliAPIKey:                 readString("MEILI_API_KEY", ""),
		MeiliIndex:                  readString("MEILI_INDEX", "npan_items"),
		TypesenseHost:               readString("TYPESENSE_HOST", "http://127.0.0.1:8108"),
		TypesenseAPIKey:             readString("TYPESENSE_API_KEY", ""),
		TypesenseCollection:         readString("TYPESENSE_COLLECTION", "npan_items"),
		PublicSearchHost:            readString("MEILI_PUBLIC_SEARCH_HOST", ""),
		PublicSearchIndexName:       readString("MEILI_PUBLIC_SEARCH_INDEX", ""),
		PublicSearchAPIKey:          readString("MEILI_PUBLIC_SEARCH_API_KEY", ""),
		TypesensePublicSearchHost:   readString("TYPESENSE_PUBLIC_SEARCH_HOST", ""),
		TypesensePublicSearchIndex:  readString("TYPESENSE_PUBLIC_SEARCH_INDEX", ""),
		TypesensePublicSearchAPIKey: readString("TYPESENSE_PUBLIC_SEARCH_API_KEY", ""),
		PublicSearchInstantsearchOn: readBool("NPA_PUBLIC_INSTANTSEARCH_ENABLED", readBool("MEILI_PUBLIC_INSTANTSEARCH_ENABLED", false)),

		StateDBFile:         readString("NPA_STATE_DB_FILE", "./data/state/sync-state.sqlite"),
		CheckpointTemplate:  readString("NPA_CHECKPOINT_FILE", "./data/checkpoints/full-crawl.json"),
		ProgressFile:        readString("NPA_PROGRESS_FILE", "./data/progress/full-sync-progress.json"),
		SyncStateFile:       readString("NPA_SYNC_STATE_FILE", "./data/progress/incremental-sync-state.json"),
		IncrementalQuery:    readString("NPA_INCREMENTAL_QUERY_WORDS", "* OR *"),
		SyncWindowOverlapMS: readInt64("NPA_SYNC_WINDOW_OVERLAP_MS", 2000),

		DefaultIncludeDepartments: readBool("NPA_INCLUDE_DEPARTMENTS", true),
		DefaultRootFolderIDs:      rootIDs,
		DefaultDepartmentIDs:      readInt64List("NPA_DEPARTMENT_IDS"),

		SyncMaxConcurrent:            readInt("NPA_SYNC_MAX_CONCURRENT", 2),
		SyncMinTimeMS:                readInt("NPA_SYNC_MIN_TIME_MS", 200),
		SyncRootWorkers:              readInt("NPA_SYNC_ROOT_WORKERS", 2),
		SyncProgressEvery:            readInt("NPA_SYNC_PROGRESS_EVERY", 1),
		InspectRootsMaxConcurrency:   readInt("NPA_INSPECT_ROOTS_MAX_CONCURRENCY", 6),
		InspectRootsPerFolderTimeout: readDuration("NPA_INSPECT_ROOTS_PER_FOLDER_TIMEOUT", 10*time.Second),

		Retry: models.RetryPolicyOptions{
			MaxRetries:  readInt("NPA_MAX_RETRIES", 3),
			BaseDelayMS: readInt("NPA_BASE_DELAY_MS", 500),
			MaxDelayMS:  readInt("NPA_MAX_DELAY_MS", 5000),
			JitterMS:    readInt("NPA_JITTER_MS", 200),
		},
	}
}

func (c Config) PublicSearchBootstrap(backend search.Backend) (host string, index string, apiKey string) {
	switch backend {
	case search.BackendTypesense:
		return strings.TrimSpace(c.TypesensePublicSearchHost), strings.TrimSpace(c.TypesensePublicSearchIndex), strings.TrimSpace(c.TypesensePublicSearchAPIKey)
	case search.BackendMeilisearch:
		fallthrough
	default:
		return strings.TrimSpace(c.PublicSearchHost), strings.TrimSpace(c.PublicSearchIndexName), strings.TrimSpace(c.PublicSearchAPIKey)
	}
}
