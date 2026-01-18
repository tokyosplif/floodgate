package config

import "errors"

type Kafka struct {
	Addr  string `env:"KAFKA_ADDR" envDefault:"localhost:9092"`
	Topic string `env:"KAFKA_TOPIC" envDefault:"clicks"`
}

func (k *Kafka) Validate() error {
	if k.Addr == "" {
		return errors.New("kafka address is required")
	}
	if k.Topic == "" {
		return errors.New("kafka topic is required")
	}

	return nil
}
