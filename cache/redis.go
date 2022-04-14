package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	gr "github.com/go-redis/redis"
)

type redis struct {
	client *gr.Client
	mu     *sync.Mutex
}

func newRedis(config *Config) (*redis, error) {
	client := gr.NewClient(&gr.Options{
		Addr:     fmt.Sprintf("%s://%s:%d", config.Scheme, config.Host, config.Port),
		Password: config.Password,
		DB:       config.Database,
	})

	if _, err := client.Ping().Result(); err != nil {
		return nil, err
	}

	return &redis{client: client, mu: &sync.Mutex{}}, nil
}

// Set store with TTL
func (r *redis) set(ctx context.Context, key string, value any, ttl time.Duration) error {
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
func (r *redis) get(ctx context.Context, key string, dest any) error {
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

// del by key
func (r *redis) del(ctx context.Context, key ...string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.client.Del(key...)

	return nil
}

// Close client
func (r *redis) close(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.client.Close()
}
