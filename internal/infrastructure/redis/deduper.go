package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Deduplicator struct {
	client *redis.Client
}

func NewDeduplicator(addr string) *Deduplicator {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &Deduplicator{client: rdb}
}

func (d *Deduplicator) IsUnique(ctx context.Context, clickID string) (bool, error) {
	added, err := d.client.SetNX(ctx, "click:"+clickID, time.Now().Unix(), 24*time.Hour).Result()
	if err != nil {
		return false, err
	}

	return added, nil
}
