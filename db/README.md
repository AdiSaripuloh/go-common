# Package `db`

Database query wrapper.

# Usage
```go
package main

import (
	"context"

	"github.com/AdiSaripuloh/go-common/cache"
	"github.com/AdiSaripuloh/go-common/config"
	"github.com/AdiSaripuloh/go-common/db"
	"github.com/AdiSaripuloh/go-common/logger"
	// import this when using driver "postgres"
	_ "github.com/lib/pq"
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

	conn, err := db.NewDB(&cfg.Database)
	if err != nil {
		logger.Error(ctx, "open connection", logger.Field{Key: "error", Value: err.Error()})
		return
	}
	defer conn.Close(ctx)

	type Destination struct {
		Pid     int64  `db:"pid"`
		Datname string `db:"datname"`
	}
	var (
		destTx     Destination
		datnameTx  *string
		statements = []*db.Statement{
			db.NewStatement(&destTx, "SELECT pid, datname FROM pg_stat_activity WHERE datname IS NOT NULL"),
			db.NewStatement(&datnameTx, "SELECT datname FROM pg_stat_activity WHERE pid = $1", &destTx.Pid),
		}
	)
	// execute multiple queries or single query in single transaction.
	errTx := conn.ExecTx(ctx, statements...)
	if errTx != nil {
		logger.Error(ctx, "ExecTx", logger.Field{Key: "error", Value: errTx.Error()})
		return
	}
	logger.Info(ctx, "ExecTx",
		logger.Field{Key: "destTx", Value: destTx},
		logger.Field{Key: "datnameTx", Value: datnameTx},
	)

	var (
		dest      Destination
		// using statement.Debug() log query
		statement = db.NewStatement(&dest, "SELECT pid, datname FROM pg_stat_activity WHERE datname IS NOT NULL").Debug()
	)
	// execute multiple queries or single query without transaction.
	err = conn.Exec(ctx, statement)
	if err != nil {
		logger.Error(ctx, "Exec", logger.Field{Key: "error", Value: err.Error()})
		return
	}
	logger.Info(ctx, "Exec",
		logger.Field{Key: "dest", Value: dest},
	)
}
```
# Log
`ExecTx`
```json
{
  "level": "INFO",
  "ts": "2022-04-13T05:35:25.405+0700",
  "caller": "go-common/main.go:56",
  "msg": "ExecTx",
  "data": {
    "datnameTx": "go",
    "destTx": {
      "Pid": 5817,
      "Datname": "go"
    }
  }
}
```
`Exec`
```json
{
  "level": "INFO",
  "ts": "2022-04-13T05:35:25.406+0700",
  "caller": "go-common/main.go:71",
  "msg": "Exec",
  "data": {
    "dest": {
      "Pid": 5817,
      "Datname": "go"
    }
  }
}
```