package tracking

import (
	"context"
	"errors"
	"log/slog"

	"github.com/tokyosplif/floodgate/internal/domain"
)

var (
	ErrDuplicateClick = errors.New("duplicate click detected")
	ErrBotDetected    = errors.New("bot traffic detected")
)

type ClickRepo interface {
	Publish(ctx context.Context, click domain.Click) error
}

type Deduplicator interface {
	IsUnique(ctx context.Context, clickID string) (bool, error)
}

type Service struct {
	repo     ClickRepo
	dedup    Deduplicator
	detector domain.BotDetector
	logger   *slog.Logger
}

func NewService(repo ClickRepo, dedup Deduplicator, detector domain.BotDetector, l *slog.Logger) *Service {
	return &Service{
		repo:     repo,
		dedup:    dedup,
		detector: detector,
		logger:   l,
	}
}

func (s *Service) HandleClick(ctx context.Context, click domain.Click) error {
	if isBot, reason := s.detector.IsBot(ctx, click); isBot {
		s.logger.Debug("click rejected: bot detected",
			slog.String("reason", reason),
			slog.String("ip", click.IP))
		return ErrBotDetected
	}

	unique, err := s.dedup.IsUnique(ctx, click.ID)
	if err != nil {
		s.logger.Warn("deduplicator error", slog.Any("err", err))
	}
	if !unique {
		return ErrDuplicateClick
	}

	return s.repo.Publish(ctx, click)
}
