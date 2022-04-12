# Package `db`

Database query wrapper.

# Usage
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
	defer logger.Sync()

	var (
		ctx    = context.Background()
		config = db.Config{
			Driver:                "postgres",
			Host:                  "localhost",
			Port:                  5432,
			DBName:                "go",
			User:                  "go",
			Password:              "go",
			MaxOpenConnections:    10,
			MaxLifeTimeConnection: 30,
			MaxIdleConnections:    20,
			MaxIdleTimeConnection: 30,
		}
	)

	conn, err := db.NewDB(config)
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