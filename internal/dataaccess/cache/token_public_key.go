package cache

import (
	"context"
	"fmt"
	"log"
)

type TokenPublicKey interface {
	Get(ctx context.Context, id uint64) ([]byte, error)
	Set(ctx context.Context, id uint64, bytes []byte) error
}
type tokenPublicKey struct {
	client Client
}

func NewTokenPublicKey(client Client) TokenPublicKey {
	return &tokenPublicKey{
		client: client,
	}
}
func (c tokenPublicKey) getTokenPublicKeyCacheKey(id uint64) string {
	return fmt.Sprintf("token_public_key:%d", id)
}
func (c tokenPublicKey) Get(ctx context.Context, id uint64) ([]byte, error) {
	cacheKey := c.getTokenPublicKeyCacheKey(id)
	cacheEntry, err := c.client.Get(ctx, cacheKey)
	if err != nil {
		return nil, err
	}
	if cacheEntry == nil {
		return nil, ErrCacheMiss
	}
	publicKey, ok := cacheEntry.([]byte)
	if !ok {
		log.Printf("cache entry is not of type bytes")
		return nil, nil
	}
	return publicKey, nil
}
func (c tokenPublicKey) Set(ctx context.Context, id uint64, bytes []byte) error {
	cacheKey := c.getTokenPublicKeyCacheKey(id)
	if err := c.client.Set(ctx, cacheKey, bytes, 0); err != nil {
		log.Printf("failed to insert token public key into cache")
		return err
	}
	return nil
}
