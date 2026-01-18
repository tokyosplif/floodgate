package config

import "errors"

type ClickHouse struct {
	Addr string `env:"CLICKHOUSE_ADDR" envDefault:"localhost:9000"`
}

func (c *ClickHouse) Validate() error {
	if c.Addr == "" {
		return errors.New("clickhouse address is required")
	}
	return nil
}
