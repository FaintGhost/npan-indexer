package search

import "npan/internal/models"

// Searcher 定义搜索服务的统一接口，支持缓存装饰器等扩展。
type Searcher interface {
  Query(params models.LocalSearchParams) (QueryResult, error)
  Ping() error
}

type QueryService struct {
	index IndexOperator
}

type QueryResult struct {
	Items []models.IndexDocument `json:"items"`
	Total int64                  `json:"total"`
}

func NewQueryService(index IndexOperator) *QueryService {
	return &QueryService{index: index}
}

func (s *QueryService) Query(params models.LocalSearchParams) (QueryResult, error) {
	normalized := params
	if normalized.Page <= 0 {
		normalized.Page = 1
	}
	if normalized.PageSize <= 0 {
		normalized.PageSize = 20
	}
	if normalized.PageSize > 100 {
		normalized.PageSize = 100
	}

	items, total, err := s.index.Search(normalized)
	if err != nil {
		return QueryResult{}, err
	}

	return QueryResult{Items: items, Total: total}, nil
}

// Ping 委托给底层 MeiliIndex 检查连通性。
func (s *QueryService) Ping() error {
	return s.index.Ping()
}
