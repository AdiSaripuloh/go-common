package cache

import (
	"context"
	"time"
)

type (
	Config struct {
		Scheme   string `yaml:"scheme"`
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Database int    `yaml:"database"`
	}
	Cacher interface {
		Set(ctx context.Context, key string, value any, ttl time.Duration) error
		Get(ctx context.Context, key string, dest any) error
		Del(ctx context.Context, key []string) error
	}
)
