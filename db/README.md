# Example
```go
package main

import (
	"context"
	"github.com/AdiSaripuloh/go-common/db"
	"github.com/AdiSaripuloh/go-common/logger"
	_ "github.com/lib/pq"
)

func main() {
	logger.Init(true)
	ctx := context.Background()

	var config = db.Config{
		Driver:                 "postgres",
		Host:                   "localhost",
		Port:                   5432,
		DBName:                 "go",
		User:                   "go",
		Password:               "go",
		MaxOpenConnections:     10,
		MaxLifeTimeConnections: 30,
		MaxIdleConnections:     20,
		MaxIdleTime:            30,
	}
	sqlDB, err := db.NewDB(config)
	if err != nil {
		logger.Error(ctx, "Open Connection", logger.Field{Key: "error", Value: err.Error()})
		return
	}
	defer sqlDB.Close()

	type X struct {
		Pid     int64  `db:"pid"`
		Datname string `db:"datname"`
	}
	var dest X
	var datname *string
	stmts := []*db.Statement{
		db.NewStatement(&dest, "SELECT pid, datname FROM pg_stat_activity WHERE datname IS NOT NULL"),
		db.NewStatement(&datname, "SELECT datname FROM pg_stat_activity WHERE pid = $1", &dest.Pid),
		db.NewStatement(&dest, "SELECT pid, datname FROM pg_stat_activity WHERE pid = $1", &dest.Pid),
	}
	err1 := db.WithTx(ctx, sqlDB, stmts...)
	if err1 != nil {
		logger.Error(ctx, "WithTx", logger.Field{Key: "error", Value: err1.Error()})
		return
	}
	logger.Info(ctx, "PID", logger.Field{Key: "ID", Value: dest}, logger.Field{Key: "datname", Value: datname})
}

```