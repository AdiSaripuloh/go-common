package db

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type (
	Config struct {
		Driver   string `yaml:"driver"`
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		DBName   string `yaml:"dbName"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`

		MaxOpenConnections    int `yaml:"maxOpenConnections"`
		MaxLifeTimeConnection int `yaml:"maxLifeTimeConnection"` // Seconds
		MaxIdleConnections    int `yaml:"maxIdleConnections"`
		MaxIdleTimeConnection int `yaml:"maxIdleTimeConnection"` // Seconds
	}
	DB struct {
		Conn *sqlx.DB
	}
)

func NewDB(config Config) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.DBName,
	)

	conn, err := sqlx.Open(config.Driver, dsn)
	if err != nil {
		return nil, err
	}

	conn.SetConnMaxLifetime(time.Duration(config.MaxLifeTimeConnection) * time.Second)
	conn.SetMaxOpenConns(config.MaxOpenConnections)
	conn.SetMaxIdleConns(config.MaxIdleConnections)
	conn.SetConnMaxIdleTime(time.Duration(config.MaxIdleTimeConnection) * time.Second)

	return &DB{Conn: conn}, nil
}

// A Statement is a simple wrapper for creating a statement consisting of
// a query and a set of arguments to be passed to that query.
type Statement struct {
	Dest  any // if query doesn't have any result, leave it nil.
	Query string
	Args  []any
}

// NewStatement creating new pipeline statement.
func NewStatement(dest any, query string, args ...any) *Statement {
	return &Statement{dest, query, args}
}

// Exec Execute the statement within supplied transaction and store to destination if not nil.
// TODO: Add logger for query.
func (ps *Statement) Exec(ctx context.Context, tx *sqlx.Tx) error {
	stmt, err := tx.PreparexContext(ctx, ps.Query)
	if err != nil {
		return err
	}

	// destination nil it's mean query doesn't need result
	if ps.Dest == nil {
		_, err = stmt.ExecContext(ctx, ps.Args...)
		return err
	}

	rt := reflect.TypeOf(ps.Dest)
	switch rt.Elem().Kind() {
	case reflect.Slice, reflect.Array:
		err = stmt.SelectContext(ctx, ps.Dest, ps.Args...)
		if err != nil {
			return err
		}
		break
	default:
		err = stmt.GetContext(ctx, ps.Dest, ps.Args...)
		if err != nil {
			return err
		}
		break
	}

	return nil
}

// flushDest free memory of the destination
func (ps *Statement) flushDest() {
	ps.Dest = nil
}

// RunPipeline run multiple statements in single pipeline.
func RunPipeline(ctx context.Context, tx *sqlx.Tx, pipelineStmts ...*Statement) error {
	// TODO: isolation level configurable
	for i, ps := range pipelineStmts {
		err := ps.Exec(ctx, tx)
		if err != nil {
			for j := i; j < 0; i-- {
				pipelineStmts[j].flushDest()
			}
			return errors.Errorf("stmt[%d]: %s", i, err.Error())
		}
	}

	return nil
}

// WithTx run multiple query in single transaction.
func (db *DB) WithTx(ctx context.Context, statements ...*Statement) error {
	tx, err := db.Conn.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	if err = RunPipeline(ctx, tx, statements...); err != nil {
		if er := tx.Rollback(); er != nil {
			return er
		}
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

// Close connection
func (db *DB) Close(ctx context.Context) error {
	return db.Conn.Close()
}
