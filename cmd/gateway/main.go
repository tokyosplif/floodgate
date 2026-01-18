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

	application := app.New(cfg, l)

	if err := application.Run(); err != nil {
		l.Error("gateway stopped with error", "error", err)
		os.Exit(1)
	}
}
