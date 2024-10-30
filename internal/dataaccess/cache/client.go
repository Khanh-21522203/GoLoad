package cache

import (
	"GoLoad/internal/configs"
	"context"
	"errors"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	ErrCacheMiss = errors.New("cache miss")
)

type Client interface {
	Set(ctx context.Context, key string, data any, ttl time.Duration) error
	Get(ctx context.Context, key string) (any, error)
	AddToSet(ctx context.Context, key string, data ...any) error
	IsDataInSet(ctx context.Context, key string, data any) (bool, error)
}
type client struct {
	redisClient *redis.Client
}

func NewClient(cacheConfig configs.Cache) (Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cacheConfig.Address,
		Username: cacheConfig.Username,
		Password: cacheConfig.Password,
	})
	return &client{
		redisClient: redisClient,
	}, nil
}
func (r client) Set(ctx context.Context, key string, data any, ttl time.Duration) error {
	if err := r.redisClient.Set(ctx, key, data, ttl).Err(); err != nil {
		log.Printf("failed to set data into cache")
		return err
	}
	return nil
}
func (r client) Get(ctx context.Context, key string) (any, error) {
	data, err := r.redisClient.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrCacheMiss
		}
		log.Printf("failed to get data from cache")
		return nil, err
	}
	return data, nil
}
func (r client) AddToSet(ctx context.Context, key string, data ...any) error {

	if err := r.redisClient.SAdd(ctx, key, data...).Err(); err != nil {
		log.Printf("failed to set data into set inside cache")
		return err
	}
	return nil
}
func (r client) IsDataInSet(ctx context.Context, key string, data any) (bool, error) {

	result, err := r.redisClient.SIsMember(ctx, key, data).Result()
	if err != nil {
		log.Printf("failed to check if data is member of set inside cache")
		return false, err
	}
	return result, nil
}
