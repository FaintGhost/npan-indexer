package config

import (
	"fmt"
	"log/slog"
	"strings"

	"npan/internal/search"
)

func (c Config) Validate() error {
	var errs []string
	backend, backendErr := search.ParseBackend(c.SearchBackend)
	if backendErr != nil {
		errs = append(errs, backendErr.Error())
	}

	if strings.TrimSpace(c.AdminAPIKey) == "" {
		errs = append(errs, "NPA_ADMIN_API_KEY 不能为空")
	} else if len(c.AdminAPIKey) < 16 {
		errs = append(errs, "NPA_ADMIN_API_KEY 长度不应少于 16 字符")
	}

	switch backend {
	case search.BackendTypesense:
		if strings.TrimSpace(c.TypesenseHost) == "" {
			errs = append(errs, "TYPESENSE_HOST 不能为空")
		}
		if strings.TrimSpace(c.TypesenseCollection) == "" {
			errs = append(errs, "TYPESENSE_COLLECTION 不能为空")
		}
	case search.BackendMeilisearch, "":
		if strings.TrimSpace(c.MeiliHost) == "" {
			errs = append(errs, "MEILI_HOST 不能为空")
		}
		if strings.TrimSpace(c.MeiliIndex) == "" {
			errs = append(errs, "MEILI_INDEX 不能为空")
		}
	}
	if c.BaseURL == "" {
		errs = append(errs, "NPA_BASE_URL 不能为空")
	}
	if strings.TrimSpace(c.StateDBFile) == "" {
		errs = append(errs, "NPA_STATE_DB_FILE 不能为空")
	}

	publicSearchHost := strings.TrimSpace(c.PublicSearchHost)
	publicSearchIndexName := strings.TrimSpace(c.PublicSearchIndexName)
	publicSearchAPIKey := strings.TrimSpace(c.PublicSearchAPIKey)
	typesensePublicSearchHost := strings.TrimSpace(c.TypesensePublicSearchHost)
	typesensePublicSearchIndex := strings.TrimSpace(c.TypesensePublicSearchIndex)
	typesensePublicSearchAPIKey := strings.TrimSpace(c.TypesensePublicSearchAPIKey)
	privateMeiliAPIKey := strings.TrimSpace(c.MeiliAPIKey)
	privateTypesenseAPIKey := strings.TrimSpace(c.TypesenseAPIKey)

	if backend == search.BackendMeilisearch && publicSearchAPIKey == "" && privateMeiliAPIKey != "" {
		errs = append(errs, "MEILI_PUBLIC_SEARCH_API_KEY 不能为空，不能回落复用私有 MEILI_API_KEY，浏览器公开 search 必须使用 dedicated search-only key")
	}
	if backend == search.BackendTypesense && typesensePublicSearchAPIKey == "" && privateTypesenseAPIKey != "" {
		errs = append(errs, "TYPESENSE_PUBLIC_SEARCH_API_KEY 不能为空，不能回落复用私有 TYPESENSE_API_KEY，浏览器公开 search 必须使用 dedicated search-only key")
	}
	if c.PublicSearchInstantsearchOn {
		if search.SupportsPublicInstantsearch(backend) {
			switch backend {
			case search.BackendTypesense:
				if typesensePublicSearchHost == "" {
					errs = append(errs, "TYPESENSE_PUBLIC_SEARCH_HOST 不能为空，开启公开搜索时必须显式提供 public host")
				}
				if typesensePublicSearchIndex == "" {
					errs = append(errs, "TYPESENSE_PUBLIC_SEARCH_INDEX 不能为空，开启公开搜索时必须显式提供 public index")
				}
				if typesensePublicSearchAPIKey == "" {
					errs = append(errs, "TYPESENSE_PUBLIC_SEARCH_API_KEY 不能为空，开启公开搜索时必须使用 dedicated search-only key")
				}
			case search.BackendMeilisearch:
				fallthrough
			default:
				if publicSearchHost == "" {
					errs = append(errs, "MEILI_PUBLIC_SEARCH_HOST 不能为空，开启公开搜索时必须显式提供 public host")
				}
				if publicSearchIndexName == "" {
					errs = append(errs, "MEILI_PUBLIC_SEARCH_INDEX 不能为空，开启公开搜索时必须显式提供 public index")
				}
				if publicSearchAPIKey == "" {
					errs = append(errs, "MEILI_PUBLIC_SEARCH_API_KEY 不能为空，开启公开搜索时必须使用 dedicated search-only key")
				}
			}
		} else {
			errs = append(errs, "当前搜索后端不支持浏览器直连 InstantSearch")
		}
	}
	if backend == search.BackendMeilisearch && publicSearchAPIKey != "" && publicSearchAPIKey == privateMeiliAPIKey {
		errs = append(errs, "MEILI_PUBLIC_SEARCH_API_KEY 不能复用私有 MEILI_API_KEY")
	}
	if backend == search.BackendTypesense && typesensePublicSearchAPIKey != "" && typesensePublicSearchAPIKey == privateTypesenseAPIKey {
		errs = append(errs, "TYPESENSE_PUBLIC_SEARCH_API_KEY 不能复用私有 TYPESENSE_API_KEY")
	}

	if c.SyncMaxConcurrent <= 0 || c.SyncMaxConcurrent > 20 {
		errs = append(errs, "NPA_SYNC_MAX_CONCURRENT 应在 1-20 之间")
	}
	if c.Retry.MaxRetries < 0 || c.Retry.MaxRetries > 10 {
		errs = append(errs, "NPA_MAX_RETRIES 应在 0-10 之间")
	}

	hasClientCreds := c.ClientID != "" && c.ClientSecret != "" && c.SubID > 0
	hasToken := c.Token != ""
	if c.AllowConfigAuthFallback && !hasClientCreds && !hasToken {
		errs = append(errs, "NPA_ALLOW_CONFIG_AUTH_FALLBACK=true 但未提供有效凭据")
	}

	if len(errs) > 0 {
		return fmt.Errorf("配置验证失败:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return nil
}

func (c Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("ServerAddr", c.ServerAddr),
		slog.String("MetricsAddr", c.MetricsAddr),
		slog.String("BaseURL", c.BaseURL),
		slog.String("SearchBackend", c.SearchBackend),
		slog.String("MeiliHost", c.MeiliHost),
		slog.String("MeiliIndex", c.MeiliIndex),
		slog.String("TypesenseHost", c.TypesenseHost),
		slog.String("TypesenseCollection", c.TypesenseCollection),
		slog.String("PublicSearchHost", c.PublicSearchHost),
		slog.String("PublicSearchIndexName", c.PublicSearchIndexName),
		slog.String("TypesensePublicSearchHost", c.TypesensePublicSearchHost),
		slog.String("TypesensePublicSearchIndex", c.TypesensePublicSearchIndex),
		slog.Bool("PublicSearchInstantsearchOn", c.PublicSearchInstantsearchOn),
		slog.String("AdminAPIKey", "[REDACTED]"),
		slog.String("ClientSecret", "[REDACTED]"),
		slog.String("Token", "[REDACTED]"),
		slog.String("MeiliAPIKey", "[REDACTED]"),
		slog.String("TypesenseAPIKey", "[REDACTED]"),
		slog.String("PublicSearchAPIKey", "[REDACTED]"),
		slog.String("TypesensePublicSearchAPIKey", "[REDACTED]"),
	)
}
