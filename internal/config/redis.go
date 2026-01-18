package config

import "errors"

type Redis struct {
	Addr string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
}

func (r *Redis) Validate() error {
	if r.Addr == "" {
		return errors.New("redis address is required")
	}
	return nil
}
