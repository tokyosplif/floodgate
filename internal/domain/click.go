package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidClickID = errors.New("click id is required")
)

type BotDetector interface {
	IsBot(ctx context.Context, click Click) (bool, string)
}

type Click struct {
	ID          string            `json:"id"`
	CampaignID  string            `json:"campaign_id"`
	Source      string            `json:"source"`
	IP          string            `json:"ip"`
	UA          string            `json:"user_agent"`
	Referer     string            `json:"referer"`
	Language    string            `json:"language"`
	Params      map[string]string `json:"params"`
	ProcessedAt time.Time         `json:"processed_at"`
}

func (c *Click) Validate() error {
	if c.ID == "" {
		return ErrInvalidClickID
	}
	return nil
}
