package app

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/tokyosplif/floodgate/internal/config"
	"github.com/tokyosplif/floodgate/internal/domain"
	"github.com/tokyosplif/floodgate/internal/infrastructure/clickhouse"
)

type Processor struct {
	cfg    *config.Config
	logger *slog.Logger
	chRepo *clickhouse.Repository
}

func NewProcessor(cfg *config.Config, l *slog.Logger) *Processor {
	chRepo, err := clickhouse.NewRepository(cfg.ClickHouse.Addr)
	if err != nil {
		l.Error("failed to init clickhouse repo", "err", err)
		os.Exit(1)
	}

	return &Processor{
		cfg:    cfg,
		logger: l,
		chRepo: chRepo,
	}
}

func (p *Processor) Run() error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{p.cfg.Kafka.Addr},
		Topic:    p.cfg.Kafka.Topic,
		GroupID:  "click-processor-group",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	p.logger.Info("Processor started, consuming Kafka...")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	batch := make([]domain.Click, 0, 1000)
	ticker := time.NewTicker(5 * time.Second)

	p.logger.Info("Processor started and waiting for messages")

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("Shutdown signal received. Flushing remaining data...")

			flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			if err := p.flush(flushCtx, batch); err != nil {
				p.logger.Error("Final flush failed", "error", err)
			}

			cancel()

			return nil

		case <-ticker.C:
			if len(batch) > 0 {
				if err := p.flush(ctx, batch); err != nil {
					p.logger.Error("Ticker flush failed", "error", err)
				}
				batch = batch[:0]
			}

		default:
			m, err := reader.FetchMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					continue
				}
				continue
			}

			var click domain.Click
			if err := json.Unmarshal(m.Value, &click); err != nil {
				p.logger.Error("Unmarshal error", "error", err)
				continue
			}

			batch = append(batch, click)

			if len(batch) >= 1000 {
				if err := p.flush(ctx, batch); err != nil {
					p.logger.Error("Size flush failed", "error", err)
				}
				batch = batch[:0]
			}

			_ = reader.CommitMessages(ctx, m)
		}
	}
}

func (p *Processor) flush(ctx context.Context, clicks []domain.Click) error {
	if len(clicks) == 0 {
		return nil
	}
	p.logger.Info("flushing clicks to ClickHouse", "count", len(clicks))
	return p.chRepo.BatchInsert(ctx, clicks)
}
