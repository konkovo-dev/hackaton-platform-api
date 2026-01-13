package client

import (
	"time"

	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	gocache "github.com/patrickmn/go-cache"
)

type cache interface {
	Get(key string) (claims *auth.Claims, err error)
	Set(key string, claims *auth.Claims, ttl time.Duration) error
}

func newCache(cfg *Config) cache {
	return newInMemoryCache(cfg)
}

type inMemoryCache struct {
	*gocache.Cache
}

func newInMemoryCache(cfg *Config) *inMemoryCache {
	return &inMemoryCache{
		Cache: gocache.New(gocache.NoExpiration, cfg.CacheCleanupInterval),
	}
}

func (c *inMemoryCache) Get(key string) (claims *auth.Claims, err error) {
	var (
		v  interface{}
		ok bool
	)

	if v, ok = c.Cache.Get(key); !ok {
		return nil, ErrTokenNotFound
	}

	claims, ok = v.(*auth.Claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (c *inMemoryCache) Set(key string, claims *auth.Claims, ttl time.Duration) error {
	c.Cache.Set(key, claims, ttl)
	return nil
}
