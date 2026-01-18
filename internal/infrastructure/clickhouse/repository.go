package clickhouse

import (
	"context"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/tokyosplif/floodgate/internal/domain"
)

type Repository struct {
	conn clickhouse.Conn
}

func NewRepository(addr string) (*Repository, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: "default",
			Password: "",
		},
		DialTimeout: time.Second * 30,
	})
	if err != nil {
		return nil, err
	}

	return &Repository{conn: conn}, nil
}

func (r *Repository) BatchInsert(ctx context.Context, clicks []domain.Click) error {
	if len(clicks) == 0 {
		return nil
	}

	batch, err := r.conn.PrepareBatch(ctx, "INSERT INTO clicks (id, campaign_id, source, ip, user_agent, processed_at)")
	if err != nil {
		return err
	}

	for _, c := range clicks {
		if err := batch.Append(c.ID, c.CampaignID, c.Source, c.IP, c.UA, c.ProcessedAt); err != nil {
			return err
		}
	}

	return batch.Send()
}
