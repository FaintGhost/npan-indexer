package service

import (
	"context"
	"errors"

	"npan/internal/npan"
)

type DownloadURLService struct {
	api npan.API
}

func NewDownloadURLService(api npan.API) *DownloadURLService {
	return &DownloadURLService{api: api}
}

func (s *DownloadURLService) GetDownloadURL(ctx context.Context, fileID int64, validPeriod *int64) (string, error) {
	result, err := s.api.GetDownloadURL(ctx, fileID, validPeriod)
	if err != nil {
		var statusErr *npan.StatusError
		if errors.As(err, &statusErr) {
			switch statusErr.Status {
			case 404:
				return "", errors.New("文件不存在")
			case 403:
				return "", errors.New("无权限访问该文件")
			case 429:
				return "", errors.New("下载接口限流，请稍后重试")
			}
		}
		return "", err
	}

	if result.DownloadURL == "" {
		return "", errors.New("下载链接为空")
	}

	return result.DownloadURL, nil
}
