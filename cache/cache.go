package cache

import (
	"context"
	"fmt"
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
		Prefix   string `yaml:"prefix"`
	}

	cacher interface {
		set(ctx context.Context, key string, value any, ttl time.Duration) error
		get(ctx context.Context, key string, dest any) error
		del(ctx context.Context, key ...string) error
		close(ctx context.Context) error
	}

	Cache struct {
		cacher cacher
		prefix string
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

	return &Cache{cacher: c, prefix: config.Prefix}, nil
}

func (c *Cache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return c.cacher.set(ctx, setPrefix(c.prefix, key), value, ttl)
}

func (c *Cache) Get(ctx context.Context, key string, dest any) error {
	return c.cacher.get(ctx, setPrefix(c.prefix, key), dest)
}

func (c *Cache) Del(ctx context.Context, keys ...string) error {
	return c.cacher.del(ctx, setPrefixes(c.prefix, keys...)...)
}

func (c *Cache) Close(ctx context.Context) error {
	return c.cacher.close(ctx)
}

func setPrefix(prefix, key string) string {
	if prefix == "" {
		return key
	}

	return fmt.Sprintf("%s:%s", prefix, key)
}

func setPrefixes(prefix string, keys ...string) []string {
	var keysWithPrefix []string
	for _, key := range keys {
		keysWithPrefix = append(keysWithPrefix, setPrefix(prefix, key))
	}

	return keysWithPrefix
}
