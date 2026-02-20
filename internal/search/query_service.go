package search

import "npan/internal/models"

type QueryService struct {
	index *MeiliIndex
}

type QueryResult struct {
	Items []models.IndexDocument `json:"items"`
	Total int64                  `json:"total"`
}

func NewQueryService(index *MeiliIndex) *QueryService {
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

	items, total, err := s.index.Search(normalized)
	if err != nil {
		return QueryResult{}, err
	}

	return QueryResult{Items: items, Total: total}, nil
}
