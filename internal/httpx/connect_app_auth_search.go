package httpx

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"connectrpc.com/connect"

	npanv1 "npan/gen/go/npan/v1"
	"npan/internal/models"
	"npan/internal/npan"
	"npan/internal/search"
	"npan/internal/service"
)

type appConnectServer struct {
	handlers *Handlers
}

func newAppConnectServer(handlers *Handlers) *appConnectServer {
	return &appConnectServer{handlers: handlers}
}

type authConnectServer struct {
	handlers *Handlers
}

func newAuthConnectServer(handlers *Handlers) *authConnectServer {
	return &authConnectServer{handlers: handlers}
}

type searchConnectServer struct {
	handlers *Handlers
}

func newSearchConnectServer(handlers *Handlers) *searchConnectServer {
	return &searchConnectServer{handlers: handlers}
}

func (h *Handlers) resolveAuthOptionsForConnect(header http.Header, payload authPayload, allowFallback bool) npan.AuthResolverOptions {
	tokenFromHeader := ""
	if header != nil {
		tokenFromHeader = parseBearerHeaderValue(header.Get("Authorization"))
	}

	tokenCandidates := []string{payload.Token, tokenFromHeader}
	clientIDCandidates := []string{payload.ClientID}
	clientSecretCandidates := []string{payload.ClientSecret}
	subIDCandidates := []int64{payload.SubID}
	oauthHostCandidates := []string{payload.OAuthHost}

	if allowFallback {
		tokenCandidates = append(tokenCandidates, h.cfg.Token)
		clientIDCandidates = append(clientIDCandidates, h.cfg.ClientID)
		clientSecretCandidates = append(clientSecretCandidates, h.cfg.ClientSecret)
		subIDCandidates = append(subIDCandidates, h.cfg.SubID)
		oauthHostCandidates = append(oauthHostCandidates, h.cfg.OAuthHost)
	}

	subType := npan.TokenSubjectType(payload.SubType)
	if subType == "" && allowFallback {
		subType = h.cfg.SubType
	}
	if subType == "" {
		subType = npan.TokenSubjectUser
	}

	oauthHost := firstNotEmpty(oauthHostCandidates...)
	if oauthHost == "" {
		oauthHost = npan.DefaultOAuthHost
	}

	return npan.AuthResolverOptions{
		Token:        firstNotEmpty(tokenCandidates...),
		ClientID:     firstNotEmpty(clientIDCandidates...),
		ClientSecret: firstNotEmpty(clientSecretCandidates...),
		SubID:        firstPositive(subIDCandidates...),
		SubType:      subType,
		OAuthHost:    oauthHost,
	}
}

func (h *Handlers) resolveTokenForConnect(ctx context.Context, header http.Header, payload authPayload, allowFallback bool) (string, npan.AuthResolverOptions, error) {
	authOptions := h.resolveAuthOptionsForConnect(header, payload, allowFallback)
	token, err := npan.ResolveBearerToken(ctx, nil, authOptions)
	if err != nil {
		return "", authOptions, err
	}
	return token, authOptions, nil
}

func (s *appConnectServer) AppSearch(_ context.Context, req *connect.Request[npanv1.AppSearchRequest]) (*connect.Response[npanv1.AppSearchResponse], error) {
	query := strings.TrimSpace(req.Msg.GetQuery())
	if query == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("缺少 query 参数"))
	}

	page := int64(1)
	if req.Msg.Page != nil {
		page = req.Msg.GetPage()
		if page <= 0 {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("page 必须是正整数"))
		}
	}

	pageSize := int64(30)
	if req.Msg.PageSize != nil {
		pageSize = req.Msg.GetPageSize()
		if pageSize <= 0 {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("page_size 必须是正整数"))
		}
	}
	if err := validatePageSize(pageSize); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	result, err := s.handlers.queryService.Query(models.LocalSearchParams{
		Query:          query,
		Type:           string(models.ItemTypeFile),
		Page:           page,
		PageSize:       pageSize,
		IncludeDeleted: false,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("搜索服务暂不可用"))
	}

	return connect.NewResponse(&npanv1.AppSearchResponse{
		Result: toProtoQueryResult(result),
	}), nil
}

func (s *appConnectServer) AppDownloadURL(ctx context.Context, req *connect.Request[npanv1.AppDownloadURLRequest]) (*connect.Response[npanv1.AppDownloadURLResponse], error) {
	fileID := req.Msg.GetFileId()
	if fileID <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("file_id 必须是正整数"))
	}

	var validPeriod *int64
	if req.Msg.ValidPeriod != nil {
		v := req.Msg.GetValidPeriod()
		validPeriod = &v
	}

	token, authOptions, err := s.handlers.resolveTokenForConnect(ctx, req.Header(), authPayload{}, true)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("下载服务暂不可用，请联系管理员检查服务端凭据"))
	}

	api := s.handlers.newAPIClient(token, authOptions)
	downloadService := service.NewDownloadURLService(api)
	downloadURL, err := downloadService.GetDownloadURL(ctx, fileID, validPeriod)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("生成下载链接失败，请稍后重试"))
	}

	return connect.NewResponse(&npanv1.AppDownloadURLResponse{
		Result: &npanv1.DownloadURLResult{
			FileId:      fileID,
			DownloadUrl: downloadURL,
		},
	}), nil
}

func (s *authConnectServer) CreateToken(ctx context.Context, req *connect.Request[npanv1.CreateTokenRequest]) (*connect.Response[npanv1.CreateTokenResponse], error) {
	payload := authPayload{
		Token:        req.Msg.GetToken(),
		ClientID:     req.Msg.GetClientId(),
		ClientSecret: req.Msg.GetClientSecret(),
		SubID:        req.Msg.GetSubId(),
		SubType:      req.Msg.GetSubType(),
		OAuthHost:    req.Msg.GetOauthHost(),
	}
	authOptions := s.handlers.resolveAuthOptionsForConnect(req.Header(), payload, s.handlers.cfg.AllowConfigAuthFallback)
	if authOptions.ClientID == "" || authOptions.ClientSecret == "" || authOptions.SubID <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("缺少认证参数: client_id/client_secret/sub_id"))
	}

	token, err := npan.RequestAccessToken(ctx, nil, npan.TokenRequestOptions{
		OAuthHost:    authOptions.OAuthHost,
		ClientID:     authOptions.ClientID,
		ClientSecret: authOptions.ClientSecret,
		SubID:        authOptions.SubID,
		SubType:      authOptions.SubType,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("认证失败，请检查凭据"))
	}

	return connect.NewResponse(&npanv1.CreateTokenResponse{
		Token: token.AccessToken,
	}), nil
}

func (s *searchConnectServer) RemoteSearch(ctx context.Context, req *connect.Request[npanv1.RemoteSearchRequest]) (*connect.Response[npanv1.RemoteSearchResponse], error) {
	queryWords := strings.TrimSpace(req.Msg.GetQuery())
	if queryWords == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("缺少 query 参数"))
	}

	pageID := int64(0)
	if req.Msg.PageId != nil {
		pageID = req.Msg.GetPageId()
		if pageID < 0 {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("page_id 必须是 >= 0 的整数"))
		}
	}

	token, authOptions, err := s.handlers.resolveTokenForConnect(ctx, req.Header(), authPayload{}, false)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("搜索请求失败，请稍后重试"))
	}

	api := s.handlers.newAPIClient(token, authOptions)
	result, err := api.SearchItems(ctx, models.RemoteSearchParams{
		QueryWords:       queryWords,
		Type:             firstNotEmpty(req.Msg.GetType(), "all"),
		PageID:           pageID,
		QueryFilter:      firstNotEmpty(req.Msg.GetQueryFilter(), "all"),
		SearchInFolder:   req.Msg.SearchInFolder,
		UpdatedTimeRange: strings.TrimSpace(req.Msg.GetUpdatedTimeRange()),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("搜索请求失败，请稍后重试"))
	}

	return connect.NewResponse(toProtoRemoteSearchResponse(result)), nil
}

func (s *searchConnectServer) LocalSearch(_ context.Context, req *connect.Request[npanv1.LocalSearchRequest]) (*connect.Response[npanv1.LocalSearchResponse], error) {
	query := strings.TrimSpace(req.Msg.GetQuery())
	if query == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("缺少 query 参数"))
	}

	page := int64(1)
	if req.Msg.Page != nil {
		page = req.Msg.GetPage()
		if page <= 0 {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("page 必须是正整数"))
		}
	}

	pageSize := int64(20)
	if req.Msg.PageSize != nil {
		pageSize = req.Msg.GetPageSize()
		if pageSize <= 0 {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("page_size 必须是正整数"))
		}
	}
	if err := validatePageSize(pageSize); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	typeParam := firstNotEmpty(req.Msg.GetType(), "all")
	if err := validateType(typeParam); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	result, err := s.handlers.queryService.Query(models.LocalSearchParams{
		Query:          query,
		Type:           typeParam,
		Page:           page,
		PageSize:       pageSize,
		ParentID:       req.Msg.ParentId,
		UpdatedAfter:   req.Msg.UpdatedAfter,
		UpdatedBefore:  req.Msg.UpdatedBefore,
		IncludeDeleted: req.Msg.GetIncludeDeleted(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("搜索服务暂不可用"))
	}

	return connect.NewResponse(&npanv1.LocalSearchResponse{
		Result: toProtoQueryResult(result),
	}), nil
}

func (s *searchConnectServer) DownloadURL(ctx context.Context, req *connect.Request[npanv1.DownloadURLRequest]) (*connect.Response[npanv1.DownloadURLResponse], error) {
	fileID := req.Msg.GetFileId()
	if fileID <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("file_id 必须是正整数"))
	}

	var validPeriod *int64
	if req.Msg.ValidPeriod != nil {
		v := req.Msg.GetValidPeriod()
		validPeriod = &v
	}

	token, authOptions, err := s.handlers.resolveTokenForConnect(ctx, req.Header(), authPayload{}, false)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("获取下载链接失败"))
	}

	api := s.handlers.newAPIClient(token, authOptions)
	downloadService := service.NewDownloadURLService(api)
	downloadURL, err := downloadService.GetDownloadURL(ctx, fileID, validPeriod)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("获取下载链接失败"))
	}

	return connect.NewResponse(&npanv1.DownloadURLResponse{
		Result: &npanv1.DownloadURLResult{
			FileId:      fileID,
			DownloadUrl: downloadURL,
		},
	}), nil
}

func toProtoQueryResult(result search.QueryResult) *npanv1.QueryResult {
	items := make([]*npanv1.IndexDocument, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, &npanv1.IndexDocument{
			DocId:           item.DocID,
			SourceId:        item.SourceID,
			Type:            toProtoItemType(item.Type),
			Name:            item.Name,
			PathText:        item.PathText,
			ParentId:        item.ParentID,
			ModifiedAt:      item.ModifiedAt,
			CreatedAt:       item.CreatedAt,
			Size:            item.Size,
			Sha1:            item.SHA1,
			InTrash:         item.InTrash,
			IsDeleted:       item.IsDeleted,
			HighlightedName: toOptionalString(item.HighlightedName),
		})
	}

	return &npanv1.QueryResult{
		Items: items,
		Total: result.Total,
	}
}

func toProtoItemType(itemType models.ItemType) npanv1.ItemType {
	switch itemType {
	case models.ItemTypeFile:
		return npanv1.ItemType_ITEM_TYPE_FILE
	case models.ItemTypeFolder:
		return npanv1.ItemType_ITEM_TYPE_FOLDER
	default:
		return npanv1.ItemType_ITEM_TYPE_UNSPECIFIED
	}
}

func toOptionalString(raw string) *string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func toProtoRemoteSearchResponse(result models.RemoteSearchResponse) *npanv1.RemoteSearchResponse {
	files := make([]*npanv1.RemoteSearchItem, 0, len(result.Files))
	for _, item := range result.Files {
		files = append(files, &npanv1.RemoteSearchItem{
			Id:   item.ID,
			Name: item.Name,
			Type: item.Type,
		})
	}

	folders := make([]*npanv1.RemoteSearchItem, 0, len(result.Folders))
	for _, item := range result.Folders {
		folders = append(folders, &npanv1.RemoteSearchItem{
			Id:   item.ID,
			Name: item.Name,
			Type: item.Type,
		})
	}

	return &npanv1.RemoteSearchResponse{
		Files:        files,
		Folders:      folders,
		TotalCount:   result.TotalCount,
		PageId:       result.PageID,
		PageCapacity: result.PageCapacity,
		PageCount:    result.PageCount,
	}
}
