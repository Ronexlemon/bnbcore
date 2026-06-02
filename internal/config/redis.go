package config

import (
	"context"
	"log"
	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Client *redis.Client
}

func NewRedisClient(addr string) *RedisConfig {
	if addr == "" {
		addr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   0,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}

	return &RedisConfig{
		Client: client,
	}
}