package main

import (
	"log/slog"
	"os"

	"github.com/tokyosplif/floodgate/internal/app"
	"github.com/tokyosplif/floodgate/internal/config"
	"github.com/tokyosplif/floodgate/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	l := logger.InitLogger(cfg.App.Env)
	slog.SetDefault(l)

	processor := app.NewProcessor(cfg, l)

	if err := processor.Run(); err != nil {
		l.Error("processor stopped with error", "error", err)
		os.Exit(1)
	}
}
