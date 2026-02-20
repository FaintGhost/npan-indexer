package config

import (
	"errors"
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
	})
}
