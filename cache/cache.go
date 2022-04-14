package cache

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

type (
	Config struct {
		Driver   string `yaml:"driver"`
		Scheme   string `yaml:"scheme"`
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Database int    `yaml:"database"`
	}

	cacher interface {
		set(ctx context.Context, key string, value any, ttl time.Duration) error
		get(ctx context.Context, key string, dest any) error
		del(ctx context.Context, key ...string) error
		close(ctx context.Context) error
	}

	Cache struct {
		cacher cacher
	}
)

var ErrKeyNotFound = errors.New("key not found")

func NewCache(config *Config) (*Cache, error) {
	var (
		c   cacher
		err error
	)

	switch config.Driver {
	// TODO: add more driver
	default:
		c, err = newRedis(config)
		if err != nil {
			return nil, err
		}
		break
	}

	return &Cache{cacher: c}, nil
}

func (c *Cache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return c.cacher.set(ctx, key, value, ttl)
}

func (c *Cache) Get(ctx context.Context, key string, dest any) error {
	return c.cacher.get(ctx, key, dest)
}

func (c *Cache) Del(ctx context.Context, keys ...string) error {
	return c.cacher.del(ctx, keys...)
}

func (c *Cache) Close(ctx context.Context) error {
	return c.cacher.close(ctx)
}
