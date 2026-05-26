package config

import (
	"os"
	"strings"
)

type KafkaConfig struct {
	Brokers []string
}

func LoadKafkaConfig() *KafkaConfig {
	brokersEnv := os.Getenv("KAFKA_BROKERS")
	if brokersEnv == "" {
		brokersEnv = "kafka:29092"
	}

	return &KafkaConfig{
		Brokers: strings.Split(brokersEnv, ","),
	}
}