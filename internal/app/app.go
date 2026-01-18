package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tokyosplif/floodgate/internal/app/tracking"
	"github.com/tokyosplif/floodgate/internal/config"
	"github.com/tokyosplif/floodgate/internal/infrastructure/filters"
	"github.com/tokyosplif/floodgate/internal/infrastructure/kafka"
	"github.com/tokyosplif/floodgate/internal/infrastructure/metrics"
	"github.com/tokyosplif/floodgate/internal/infrastructure/redis"
	transport "github.com/tokyosplif/floodgate/internal/transport/http"
)

type App struct {
	server   *http.Server
	producer *kafka.Producer
	logger   *slog.Logger
}

func New(cfg *config.Config, l *slog.Logger) *App {
	m := metrics.NewGatewayMetrics()

	producer := kafka.NewProducer(cfg.Kafka.Addr, l)
	deduper := redis.NewDeduplicator(cfg.Redis.Addr)
	botDetector := filters.NewBotDetector()

	srv := tracking.NewService(producer, deduper, botDetector, l)

	h := transport.NewHandler(srv, l, m)

	return &App{
		logger:   l,
		producer: producer,
		server: &http.Server{
			Addr:         ":" + cfg.App.Port,
			Handler:      h.InitRouter(),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}
}

func (a *App) Run() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		a.logger.Info("server started", slog.String("addr", a.server.Addr))
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Error("server error", slog.Any("err", err))
		}
	}()

	<-stop

	return a.Shutdown()
}

func (a *App) Shutdown() error {
	a.logger.Info("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	if err := a.producer.Close(); err != nil {
		return fmt.Errorf("producer close: %w", err)
	}

	return nil
}
