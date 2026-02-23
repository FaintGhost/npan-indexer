package config

import (
	"errors"
	"log/slog"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

var dotenvOnce sync.Once

func loadDotEnv() {
	dotenvOnce.Do(func() {
		err := godotenv.Load(".env")
		if err == nil {
			return
		}

		var pathErr *os.PathError
		if errors.As(err, &pathErr) && errors.Is(pathErr.Err, os.ErrNotExist) {
			return
		}

		slog.Warn("加载 .env 文件异常", "error", err)
	})
}
