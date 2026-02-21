package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"npan/internal/models"
	"npan/internal/npan"
)

type Config struct {
	ServerAddr             string
	ServerReadHeaderTimeout time.Duration
	ServerReadTimeout       time.Duration
	ServerWriteTimeout      time.Duration
	ServerIdleTimeout       time.Duration

	AdminAPIKey             string
	AllowConfigAuthFallback bool

	BaseURL   string
	OAuthHost string

	Token        string
	ClientID     string
	ClientSecret string
	SubID        int64
	SubType      npan.TokenSubjectType

	MeiliHost   string
	MeiliAPIKey string
	MeiliIndex  string

	CheckpointTemplate string
	ProgressFile       string
	SyncStateFile        string
	IncrementalQuery     string
	SyncWindowOverlapMS  int64

	DefaultIncludeDepartments bool
	DefaultRootFolderIDs      []int64
	DefaultDepartmentIDs      []int64

	SyncMaxConcurrent int
	SyncMinTimeMS     int
	SyncRootWorkers   int
	SyncProgressEvery int

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
		AdminAPIKey:             readString("NPA_ADMIN_API_KEY", ""),
		AllowConfigAuthFallback: readBool("NPA_ALLOW_CONFIG_AUTH_FALLBACK", false),

		BaseURL:   readString("NPA_BASE_URL", "https://npan.novastar.tech:6001/openapi"),
		OAuthHost: readString("NPA_OAUTH_HOST", npan.DefaultOAuthHost),

		Token:        readString("NPA_TOKEN", ""),
		ClientID:     readString("NPA_CLIENT_ID", ""),
		ClientSecret: readString("NPA_CLIENT_SECRET", ""),
		SubID:        readInt64("NPA_SUB_ID", 0),
		SubType:      subType,

		MeiliHost:   readString("MEILI_HOST", "http://127.0.0.1:7700"),
		MeiliAPIKey: readString("MEILI_API_KEY", ""),
		MeiliIndex:  readString("MEILI_INDEX", "npan_items"),

		CheckpointTemplate: readString("NPA_CHECKPOINT_FILE", "./data/checkpoints/full-crawl.json"),
		ProgressFile:       readString("NPA_PROGRESS_FILE", "./data/progress/full-sync-progress.json"),
		SyncStateFile:       readString("NPA_SYNC_STATE_FILE", "./data/progress/incremental-sync-state.json"),
		IncrementalQuery:    readString("NPA_INCREMENTAL_QUERY_WORDS", "* OR *"),
		SyncWindowOverlapMS: readInt64("NPA_SYNC_WINDOW_OVERLAP_MS", 2000),

		DefaultIncludeDepartments: readBool("NPA_INCLUDE_DEPARTMENTS", true),
		DefaultRootFolderIDs:      rootIDs,
		DefaultDepartmentIDs:      readInt64List("NPA_DEPARTMENT_IDS"),

		SyncMaxConcurrent: readInt("NPA_SYNC_MAX_CONCURRENT", 2),
		SyncMinTimeMS:     readInt("NPA_SYNC_MIN_TIME_MS", 200),
		SyncRootWorkers:   readInt("NPA_SYNC_ROOT_WORKERS", 2),
		SyncProgressEvery: readInt("NPA_SYNC_PROGRESS_EVERY", 1),

		Retry: models.RetryPolicyOptions{
			MaxRetries:  readInt("NPA_MAX_RETRIES", 3),
			BaseDelayMS: readInt("NPA_BASE_DELAY_MS", 500),
			MaxDelayMS:  readInt("NPA_MAX_DELAY_MS", 5000),
			JitterMS:    readInt("NPA_JITTER_MS", 200),
		},
	}
}
