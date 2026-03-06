package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

type Config struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// Connect initializes the Redis client
func Connect(cfg *Config) error {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	Client = redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("Connected to Redis at %s", addr)
	return nil
}

// Close closes the Redis connection
func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}

// IsConnected checks if Redis is connected
func IsConnected() bool {
	if Client == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return Client.Ping(ctx).Err() == nil
}
