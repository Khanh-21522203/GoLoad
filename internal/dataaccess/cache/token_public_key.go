package cache

import (
	"GoLoad/internal/utils"
	"context"
	"fmt"

	"go.uber.org/zap"
)

type TokenPublicKey interface {
	Get(ctx context.Context, id uint64) ([]byte, error)
	Set(ctx context.Context, id uint64, bytes []byte) error
}
type tokenPublicKey struct {
	client Client
	logger *zap.Logger
}

func NewTokenPublicKey(client Client, logger *zap.Logger) TokenPublicKey {
	return &tokenPublicKey{
		client: client,
		logger: logger,
	}
}
func (c tokenPublicKey) getTokenPublicKeyCacheKey(id uint64) string {
	return fmt.Sprintf("token_public_key:%d", id)
}
func (c tokenPublicKey) Get(ctx context.Context, id uint64) ([]byte, error) {
	logger := utils.LoggerWithContext(ctx, c.logger).With(zap.Uint64("id", id))

	cacheKey := c.getTokenPublicKeyCacheKey(id)
	cacheEntry, err := c.client.Get(ctx, cacheKey)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get token public key cache")
		return nil, err
	}
	if cacheEntry == nil {
		return nil, ErrCacheMiss
	}
	publicKey, ok := cacheEntry.([]byte)
	if !ok {
		logger.Error("cache entry is not of type string")
		return nil, nil
	}
	return publicKey, nil
}
func (c tokenPublicKey) Set(ctx context.Context, id uint64, bytes []byte) error {
	logger := utils.LoggerWithContext(ctx, c.logger).With(zap.Uint64("id", id))

	cacheKey := c.getTokenPublicKeyCacheKey(id)
	if err := c.client.Set(ctx, cacheKey, bytes, 0); err != nil {
		logger.With(zap.Error(err)).Error("failed to insert token public key into cache")
		return err
	}
	return nil
}
