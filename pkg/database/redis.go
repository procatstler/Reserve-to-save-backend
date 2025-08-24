package database

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	PoolSize int
}

type RedisClient struct {
	*redis.Client
	ctx context.Context
}

func NewRedisClient(cfg RedisConfig) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx := context.Background()
	
	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{
		Client: client,
		ctx:    ctx,
	}, nil
}

func (r *RedisClient) SetWithExpiry(key string, value interface{}, expiration time.Duration) error {
	return r.Set(r.ctx, key, value, expiration).Err()
}

func (r *RedisClient) GetString(key string) (string, error) {
	return r.Get(r.ctx, key).Result()
}

func (r *RedisClient) GetAndDelete(key string) (string, error) {
	val, err := r.Get(r.ctx, key).Result()
	if err != nil {
		return "", err
	}
	
	if err := r.Del(r.ctx, key).Err(); err != nil {
		return val, err
	}
	
	return val, nil
}

func (r *RedisClient) Exists(key string) (bool, error) {
	val, err := r.Exists(r.ctx, key).Result()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}

func (r *RedisClient) SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.Client.SetNX(r.ctx, key, value, expiration).Result()
}

func (r *RedisClient) Close() error {
	return r.Client.Close()
}