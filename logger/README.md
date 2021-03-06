# Package `logger`

Library code that's ok to use by external applications to log their application to console.

# Usage
```go
package main

import (
	"context"

	"github.com/AdiSaripuloh/go-common/cache"
	"github.com/AdiSaripuloh/go-common/config"
	"github.com/AdiSaripuloh/go-common/db"
	"github.com/AdiSaripuloh/go-common/logger"
	// import this when using config.BindFromConsul
	_ "github.com/spf13/viper/remote"
)

func main() {
	ctx := context.Background()

	// config
	type Config struct {
		Logger   logger.Config `yaml:"logger"`
		Database db.Config     `yaml:"database"`
		Cache    cache.Config  `yaml:"cache"`
	}
	var cfg Config
	err := config.BindFromFile(&cfg, "config.yaml", ".")
	if err != nil {
		logger.Info(ctx, "BindFromFile",
			logger.Field{Key: "error", Value: err.Error()},
		)
		panic(err)
	}

	// logger
	logger.Init(&cfg.Logger)
	defer logger.Sync()
	
	logger.Info(context.Background(), "logger without stacktrace")
	logger.Info(context.Background(), "logger with data", logger.Field{Key: "Key", Value: "Value"})
}
```