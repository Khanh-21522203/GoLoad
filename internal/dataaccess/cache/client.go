package cache

import (
	"GoLoad/internal/configs"
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func NewClient(
	cacheConfig configs.Cache,
) (Client, error) {
	switch cacheConfig.Type {
	case configs.CacheTypeInMemory:
		return NewInMemoryClient(), nil
	case configs.CacheTypeRedis:
		return NewRedisClient(cacheConfig), nil
	default:
		return nil, fmt.Errorf("unsupported cache type: %s", cacheConfig.Type)
	}
}

type redisClient struct {
	redisClient *redis.Client
}

func NewRedisClient(cacheConfig configs.Cache) Client {
	return &redisClient{
		redisClient: redis.NewClient(&redis.Options{
			Addr:     cacheConfig.Address,
			Username: cacheConfig.Username,
			Password: cacheConfig.Password,
		}),
	}
}
func (r redisClient) Set(ctx context.Context, key string, data any, ttl time.Duration) error {
	if err := r.redisClient.Set(ctx, key, data, ttl).Err(); err != nil {
		log.Printf("failed to set data into cache")
		return status.Error(codes.Internal, "failed to set data into cache")
	}
	return nil
}
func (r redisClient) Get(ctx context.Context, key string) (any, error) {
	data, err := r.redisClient.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrCacheMiss
		}
		log.Printf("failed to get data from cache")
		return nil, status.Error(codes.Internal, "failed to get data from cache")
	}
	return data, nil
}
func (r redisClient) AddToSet(ctx context.Context, key string, data ...any) error {

	if err := r.redisClient.SAdd(ctx, key, data...).Err(); err != nil {
		log.Printf("failed to set data into set inside cache")
		return status.Error(codes.Internal, "failed to set data into set inside cache")
	}
	return nil
}
func (r redisClient) IsDataInSet(ctx context.Context, key string, data any) (bool, error) {

	result, err := r.redisClient.SIsMember(ctx, key, data).Result()
	if err != nil {
		log.Printf("failed to check if data is member of set inside cache")
		return false, status.Error(codes.Internal, "failed to check if data is member of set inside cache")
	}
	return result, nil
}

type inMemoryClient struct {
	cache      map[string]any
	cacheMutex *sync.Mutex
}

func NewInMemoryClient() Client {
	return &inMemoryClient{
		cache:      make(map[string]any),
		cacheMutex: new(sync.Mutex),
	}
}
func (c inMemoryClient) Set(_ context.Context, key string, data any, _ time.Duration) error {
	c.cache[key] = data
	return nil
}
func (c inMemoryClient) Get(_ context.Context, key string) (any, error) {
	data, ok := c.cache[key]
	if !ok {
		return nil, ErrCacheMiss
	}
	return data, nil
}
func (c inMemoryClient) AddToSet(_ context.Context, key string, data ...any) error {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	set := c.getSet(key)
	set = append(set, data...)
	c.cache[key] = set
	return nil
}
func (c inMemoryClient) IsDataInSet(_ context.Context, key string, data any) (bool, error) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	set := c.getSet(key)
	for i := range set {
		if set[i] == data {
			return true, nil
		}
	}
	return false, nil
}
func (c inMemoryClient) getSet(key string) []any {
	setValue, ok := c.cache[key]
	if !ok {
		return make([]any, 0)
	}
	set, ok := setValue.([]any)
	if !ok {
		return make([]any, 0)
	}
	return set
}
