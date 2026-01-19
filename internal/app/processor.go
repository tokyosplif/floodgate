package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
		Brokers: []string{p.cfg.Kafka.Addr},
		Topic:   p.cfg.Kafka.Topic,
		GroupID: "click-processor-group",
		MaxWait: 1 * time.Second,
	})

	defer func() {
		if err := reader.Close(); err != nil {
			p.logger.Error("error closing reader", "err", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	batch := make([]domain.Click, 0, 1000)
	kafkaMses := make([]kafka.Message, 0, 1000)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("Shutdown signal received")
			flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := p.flushAndCommit(flushCtx, reader, &batch, &kafkaMses)
			cancel()
			if err != nil {
				return fmt.Errorf("final flush failed: %w", err)
			}
			return nil

		case <-ticker.C:
			if len(batch) > 0 {
				fCtx, fCancel := context.WithTimeout(ctx, 3*time.Second)
				if err := p.flushAndCommit(fCtx, reader, &batch, &kafkaMses); err != nil {
					p.logger.Error("Ticker flush failed", "error", err)
				}
				fCancel()
			}

		default:
			m, err := reader.FetchMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				p.logger.Error("Fetch error", "err", err)

				continue
			}

			var click domain.Click
			if err := json.Unmarshal(m.Value, &click); err != nil {
				p.logger.Error("Unmarshal error", "error", err)
				_ = reader.CommitMessages(ctx, m)

				continue
			}

			batch = append(batch, click)
			kafkaMses = append(kafkaMses, m) // Запоминаем сообщение

			if len(batch) >= 1000 {
				fCtx, fCancel := context.WithTimeout(ctx, 3*time.Second)
				if err := p.flushAndCommit(fCtx, reader, &batch, &kafkaMses); err != nil {
					p.logger.Error("Size flush failed", "error", err)
				}
				fCancel()
			}
		}
	}
}

func (p *Processor) flushAndCommit(ctx context.Context, r *kafka.Reader, clicks *[]domain.Click, megs *[]kafka.Message) error {
	if len(*clicks) == 0 {
		return nil
	}

	if err := p.chRepo.BatchInsert(ctx, *clicks); err != nil {
		return err
	}

	if err := r.CommitMessages(ctx, *megs...); err != nil {
		return fmt.Errorf("kafka commit error: %w", err)
	}

	p.logger.Info("batch processed successfully", "count", len(*clicks))

	*clicks = (*clicks)[:0]
	*megs = (*megs)[:0]

	return nil
}

func (p *Processor) flush(ctx context.Context, clicks []domain.Click) error {
	if len(clicks) == 0 {
		return nil
	}
	p.logger.Info("flushing clicks to ClickHouse", "count", len(clicks))
	return p.chRepo.BatchInsert(ctx, clicks)
}
