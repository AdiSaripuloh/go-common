# Package `cache`
?

# Usage
```go
package main

import (
	"context"
	"time"

	"github.com/AdiSaripuloh/go-common/cache"
	"github.com/AdiSaripuloh/go-common/logger"
)

func main() {
	// logger
	logger.Init(false)
	defer logger.Sync()
	var ctx = context.Background()

	// redis
	cfg := cache.Config{
		Scheme:   "tcp",
		Host:     "localhost",
		Port:     6379,
		Database: 1,
	}
	redis, errR := cache.NewRedis(cfg)
	if errR != nil {
		logger.Error(ctx, "redis", logger.Field{Key: "error", Value: errR.Error()})
		return
	}
	type Cache struct {
		Some  string
		Thing int
	}
	var c = Cache{Some: "test", Thing: 1}
	errS := redis.Set(ctx, "redis:key", c, time.Duration(1)*time.Minute)
	if errS != nil {
		logger.Error(ctx, "Set", logger.Field{Key: "error", Value: errS.Error()})
		return
	}
	logger.Info(ctx, "c", logger.Field{Key: "cache", Value: c})
	var cd Cache
	errCd := redis.Get(ctx, "redis:key", &cd)
	if errCd != nil {
		logger.Error(ctx, "Get", logger.Field{Key: "error", Value: errCd.Error()})
		return
	}
	logger.Info(ctx, "cd", logger.Field{Key: "cache", Value: cd})
}
```

# Output
`Set`
```json
{
  "level": "INFO",
  "ts": "2022-04-13T07:23:50.057+0700",
  "caller": "go-common/main.go:40",
  "msg": "c",
  "data": {
    "cache": {
      "Some": "test",
      "Thing": 1
    }
  }
}
```
`Get`
```json
{
  "level": "INFO",
  "ts": "2022-04-13T07:23:50.057+0700",
  "caller": "go-common/main.go:47",
  "msg": "cd",
  "data": {
    "cache": {
      "Some": "test",
      "Thing": 1
    }
  }
}
```