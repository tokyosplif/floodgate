package filters

import (
	"context"

	"github.com/tokyosplif/floodgate/internal/domain"
)

type InMemoryBotDetector struct {
	blacklistedIPs map[string]struct{}
}

func NewBotDetector() *InMemoryBotDetector {
	return &InMemoryBotDetector{
		blacklistedIPs: map[string]struct{}{
			"1.1.1.1": {},
			"8.8.8.8": {},
		},
	}
}

func (d *InMemoryBotDetector) IsBot(ctx context.Context, c domain.Click) (bool, string) {
	if _, ok := d.blacklistedIPs[c.IP]; ok {
		return true, "ip_blacklist"
	}

	if c.UA == "" {
		return true, "empty_ua"
	}

	return false, ""
}
