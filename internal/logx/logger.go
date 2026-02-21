package logx

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

// NewLogger returns a terminal-friendly slog logger.
// Color mode: auto (default), always, never.
func NewLogger() *slog.Logger {
	noColor := !isTerminal(os.Stdout)

	switch strings.ToLower(strings.TrimSpace(os.Getenv("NPAN_LOG_COLOR"))) {
	case "always", "on", "true", "1":
		noColor = false
	case "never", "off", "false", "0":
		noColor = true
	}

	if strings.TrimSpace(os.Getenv("NO_COLOR")) != "" {
		noColor = true
	}

	handler := tint.NewHandler(os.Stdout, &tint.Options{
		Level:      slog.LevelInfo,
		TimeFormat: time.DateTime,
		NoColor:    noColor,
	})
	return slog.New(handler)
}

func isTerminal(file *os.File) bool {
	if file == nil {
		return false
	}
	fd := file.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}
