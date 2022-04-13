package main

import (
	"context"
	"time"

	"github.com/AdiSaripuloh/go-common/cache"
	"github.com/AdiSaripuloh/go-common/db"
	"github.com/AdiSaripuloh/go-common/logger"
	_ "github.com/lib/pq"
)

func main() {
	// logger
	logger.Init(false)
	defer logger.Sync()

	// db
	var (
		ctx    = context.Background()
		config = db.Config{
			Driver:                "postgres",
			Host:                  "localhost",
			Port:                  5432,
			DBName:                "go",
			User:                  "go",
			Password:              "go",
			SSLMode:               "disable",
			MaxOpenConnections:    10,
			MaxLifeTimeConnection: 30,
			MaxIdleConnections:    20,
			MaxIdleTimeConnection: 30,
		}
	)
	conn, err := db.NewDB(&config)
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
		statement = db.NewStatement(&dest, "SELECT pid, datname FROM pg_stat_activity WHERE datname IS NOT NULL")
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

	// redis
	cfg := cache.Config{
		Scheme:   "tcp",
		Host:     "localhost",
		Port:     6379,
		Database: 1,
	}
	redis, errR := cache.NewRedis(&cfg)
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
