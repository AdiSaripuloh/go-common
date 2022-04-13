# Package `config`

?

# Usage

```go
package main

import (
	"github.com/AdiSaripuloh/go-common/cache"
	"github.com/AdiSaripuloh/go-common/config"
	"github.com/AdiSaripuloh/go-common/db"
	"github.com/AdiSaripuloh/go-common/logger"
	// import this when using config.BindFromConsul
	_ "github.com/spf13/viper/remote"
)

func main() {
	type Config struct {
		Logger   logger.Config `yaml:"logger"`
		Database db.Config     `yaml:"database"`
		Cache    cache.Config  `yaml:"cache"`
	}

	var cfgFile Config
	err := config.BindFromFile(&cfgFile, "config.yaml", ".")
	if err != nil {
		logger.Info(ctx, "BindFromFile",
			logger.Field{Key: "error", Value: err.Error()},
		)
		panic(err)
	}

	var cfgConsul Config
	err = config.BindFromConsul(&cfgConsul, "localhost:8500", "PATH")
	if err != nil {
		logger.Info(ctx, "BindFromConsul",
			logger.Field{Key: "error", Value: err.Error()},
		)
		panic(err)
	}
}
```