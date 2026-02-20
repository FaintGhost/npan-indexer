package npan

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

const DefaultOAuthHost = "https://npan.novastar.tech:6001/openoauth"

type TokenSubjectType string

const (
	TokenSubjectUser       TokenSubjectType = "user"
	TokenSubjectEnterprise TokenSubjectType = "enterprise"
)

type TokenRequestOptions struct {
	OAuthHost    string
	ClientID     string
	ClientSecret string
	SubID        int64
	SubType      TokenSubjectType
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type,omitempty"`
	ExpiresIn   int64  `json:"expires_in,omitempty"`
}

type AuthResolverOptions struct {
	Token        string
	ClientID     string
	ClientSecret string
	SubID        int64
	SubType      TokenSubjectType
	OAuthHost    string
}

func normalizeSubType(subType TokenSubjectType) TokenSubjectType {
	if subType == "" {
		return TokenSubjectUser
	}
	return subType
}

func normalizeOAuthHost(host string) string {
	trimmed := strings.TrimSpace(host)
	if trimmed == "" {
		return DefaultOAuthHost
	}
	return trimmed
}

func normalizeHost(host string) string {
	return strings.TrimRight(host, "/")
}

func BuildTokenURL(oauthHost string, assertion string) (string, error) {
	base := normalizeHost(oauthHost)
	if base == "" {
		base = DefaultOAuthHost
	}

	parsed, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	parsed.Path = strings.TrimRight(parsed.Path, "/") + "/oauth/token"
	query := parsed.Query()
	query.Set("grant_type", "jwt_simple")
	query.Set("assertion", assertion)
	parsed.RawQuery = query.Encode()

	return parsed.String(), nil
}

func RequestAccessToken(ctx context.Context, client *http.Client, options TokenRequestOptions) (*AccessTokenResponse, error) {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	subType := normalizeSubType(options.SubType)

	now := time.Now().Unix()
	payload := map[string]string{
		"yifangyun_sub_type": string(subType),
		"sub":                fmt.Sprintf("%d", options.SubID),
		"iat":                fmt.Sprintf("%d", now),
		"exp":                fmt.Sprintf("%d", now+60),
		"jti":                fmt.Sprintf("%d-%s", time.Now().UnixMilli(), uuid.NewString()),
	}

	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	assertion := base64.StdEncoding.EncodeToString(rawPayload)
	tokenURL, err := BuildTokenURL(normalizeOAuthHost(options.OAuthHost), assertion)
	if err != nil {
		return nil, err
	}

	basicAuth := base64.StdEncoding.EncodeToString([]byte(options.ClientID + ":" + options.ClientSecret))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, bytes.NewReader(nil))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Basic "+basicAuth)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &StatusError{Status: resp.StatusCode, Message: "获取 token 失败"}
	}

	var body AccessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	if body.AccessToken == "" {
		return nil, fmt.Errorf("获取 token 失败: 响应中缺少 access_token")
	}

	return &body, nil
}

func ResolveBearerToken(ctx context.Context, client *http.Client, options AuthResolverOptions) (string, error) {
	if strings.TrimSpace(options.Token) != "" {
		return strings.TrimSpace(options.Token), nil
	}

	if options.ClientID == "" || options.ClientSecret == "" || options.SubID <= 0 {
		return "", fmt.Errorf("缺少认证参数: token 或 client_id/client_secret/sub_id")
	}

	tokenResp, err := RequestAccessToken(ctx, client, TokenRequestOptions{
		OAuthHost:    normalizeOAuthHost(options.OAuthHost),
		ClientID:     options.ClientID,
		ClientSecret: options.ClientSecret,
		SubID:        options.SubID,
		SubType:      normalizeSubType(options.SubType),
	})
	if err != nil {
		return "", err
	}
	return tokenResp.AccessToken, nil
}

func CanAutoRefresh(options AuthResolverOptions) bool {
	return strings.TrimSpace(options.ClientID) != "" &&
		strings.TrimSpace(options.ClientSecret) != "" &&
		options.SubID > 0
}

func NewTokenRefresher(client *http.Client, options AuthResolverOptions) func(ctx context.Context) (string, error) {
	if !CanAutoRefresh(options) {
		return nil
	}

	refreshOptions := TokenRequestOptions{
		OAuthHost:    normalizeOAuthHost(options.OAuthHost),
		ClientID:     strings.TrimSpace(options.ClientID),
		ClientSecret: strings.TrimSpace(options.ClientSecret),
		SubID:        options.SubID,
		SubType:      normalizeSubType(options.SubType),
	}

	return func(ctx context.Context) (string, error) {
		tokenResp, err := RequestAccessToken(ctx, client, refreshOptions)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(tokenResp.AccessToken) == "" {
			return "", fmt.Errorf("刷新 token 失败: access_token 为空")
		}
		return strings.TrimSpace(tokenResp.AccessToken), nil
	}
}
