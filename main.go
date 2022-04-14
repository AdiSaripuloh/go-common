package main

import (
	"context"
	"time"

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

	// db
	sqlxDB, err := db.NewDB(&cfg.Database)
	if err != nil {
		logger.Error(ctx, "open connection", logger.Field{Key: "error", Value: err.Error()})
		return
	}
	defer sqlxDB.Close()
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
	errTx := db.ExecTx(ctx, sqlxDB, statements...)
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
		statement = db.NewStatement(&dest, "SELECT pid, datname FROM pg_stat_activity WHERE datname IS NOT NULL")
	)
	// execute multiple queries or single query without transaction.
	err = db.Exec(ctx, sqlxDB, statement)
	if err != nil {
		logger.Error(ctx, "Exec", logger.Field{Key: "error", Value: err.Error()})
		return
	}
	logger.Info(ctx, "Exec",
		logger.Field{Key: "dest", Value: dest},
	)

	// cache
	cac, errR := cache.NewCache(&cfg.Cache)
	if errR != nil {
		logger.Error(ctx, "redis", logger.Field{Key: "error", Value: errR.Error()})
		return
	}
	type Cache struct {
		Some  string
		Thing int
	}
	var c = Cache{Some: "test", Thing: 1}
	errS := cac.Set(ctx, "redis:key", c, time.Duration(1)*time.Minute)
	if errS != nil {
		logger.Error(ctx, "Set", logger.Field{Key: "error", Value: errS.Error()})
		return
	}
	logger.Info(ctx, "c", logger.Field{Key: "cache", Value: c})
	var cd Cache
	errCd := cac.Get(ctx, "redis:key", &cd)
	if errCd != nil {
		logger.Error(ctx, "Get", logger.Field{Key: "error", Value: errCd.Error()})
		return
	}
	logger.Info(ctx, "cd", logger.Field{Key: "cache", Value: cd})
}
