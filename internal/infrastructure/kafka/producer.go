package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/tokyosplif/floodgate/internal/domain"
)

type Producer struct {
	writer *kafka.Writer
	logger *slog.Logger
}

func NewProducer(brokers string, l *slog.Logger) *Producer {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(brokers),
		Topic:                  "clicks",
		Balancer:               &kafka.LeastBytes{},
		WriteTimeout:           10 * time.Second,
		ReadTimeout:            10 * time.Second,
		AllowAutoTopicCreation: true,
		BatchSize:              100,
		BatchTimeout:           50 * time.Millisecond,
		Async:                  true,
	}

	return &Producer{
		writer: w,
		logger: l,
	}
}

func (p *Producer) Publish(ctx context.Context, click domain.Click) error {
	payload, err := json.Marshal(click)
	if err != nil {
		p.logger.Error("failed to marshal click", slog.Any("error", err))
		return err
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(click.ID),
		Value: payload,
	})

	if err != nil {
		p.logger.Error("failed to publish message to kafka",
			slog.String("click_id", click.ID),
			slog.Any("error", err),
		)
		return err
	}

	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
