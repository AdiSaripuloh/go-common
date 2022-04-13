package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

type Redis struct {
	client *redis.Client
	mu     *sync.Mutex
}

var ErrKeyNotFound = errors.New("key not found")

func NewRedis(config *Config) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s://%s:%d", config.Scheme, config.Host, config.Port),
		Password: config.Password,
		DB:       config.Database,
	})

	if _, err := client.Ping().Result(); err != nil {
		return nil, err
	}

	return &Redis{client: client, mu: &sync.Mutex{}}, nil
}

// Set store with TTL
func (r *Redis) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var str, err = json.Marshal(value)
	if err != nil {
		return err
	}

	r.client.Set(key, str, ttl)

	return nil
}

// Get by key
func (r *Redis) Get(ctx context.Context, key string, dest any) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var value = strings.TrimSpace(r.client.Get(key).Val())
	if len(value) == 0 {
		return ErrKeyNotFound
	}

	err := json.Unmarshal([]byte(value), dest)
	if err != nil {
		return err
	}

	return nil
}

// Del by key
func (r *Redis) Del(ctx context.Context, key []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.client.Del(key...)

	return nil
}

// Close client
func (r *Redis) Close(ctx context.Context, key []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.client.Close()
}
