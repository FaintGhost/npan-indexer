package main

import (
	"context"
	"fmt"
	"os"

	"npan/internal/cli"
	"npan/internal/config"
)

func main() {
	cfg := config.Load()
	rootCmd := cli.NewRootCommand(cfg)

	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "错误:", err)
		os.Exit(1)
	}
}
