package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/cuducos/docs-cpi-pandemia/cli"
)

func main() {
	if os.Getenv("DEBUG") != "" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	if err := cli.CLI().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
