package db

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/AdiSaripuloh/go-common/logger"

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
		SSLMode  string `yaml:"sslMode"`

		MaxOpenConnections    int `yaml:"maxOpenConnections"`
		MaxLifeTimeConnection int `yaml:"maxLifeTimeConnection"` // Seconds
		MaxIdleConnections    int `yaml:"maxIdleConnections"`
		MaxIdleTimeConnection int `yaml:"maxIdleTimeConnection"` // Seconds
	}
	// A Statement is a simple wrapper for creating a statement consisting of
	// a query and a set of arguments to be passed to that query.
	Statement struct {
		dest        any // if query doesn't have any result, leave it nil.
		query       string
		args        []any
		enableDebug bool
		mu          *sync.Mutex
	}
	DB struct {
		conn *sqlx.DB
	}

	Executor interface {
		Exec(ctx context.Context, statements ...*Statement) error
		ExecTx(ctx context.Context, statements ...*Statement) error
	}
)

// NewStatement creating new pipeline statement.
func NewStatement(dest any, query string, args ...any) *Statement {
	return &Statement{dest, query, args, false, &sync.Mutex{}}
}

func (s *Statement) SetDestination(dest any) *Statement {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.dest = dest

	return s
}

func (s *Statement) GetDestination() any {
	return s.dest
}

func (s *Statement) SetQuery(query string) *Statement {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.query = query

	return s
}

func (s *Statement) GetQuery() string {
	return s.query
}

func (s *Statement) SetArgs(args []any) *Statement {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.args = args

	return s
}

func (s *Statement) GetArgs() []any {
	return s.args
}

func (s *Statement) log(ctx context.Context) {
	logger.Debug(ctx, "statement debug",
		logger.Field{Key: "query", Value: s.GetQuery()},
		logger.Field{Key: "args", Value: s.GetArgs()},
		logger.Field{Key: "dest", Value: s.GetDestination()},
	)
}

func (s *Statement) Debug() *Statement {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.enableDebug = true

	return s
}

// exec Execute the statement within supplied transaction and update the
// destination if not nil.
func (s *Statement) exec(ctx context.Context, stmt *sqlx.Stmt) error {
	var err error

	// destination nil it's mean statement doesn't need result
	if s.GetDestination() == nil {
		_, err = stmt.ExecContext(ctx, s.GetArgs()...)
		return err
	}

	rt := reflect.TypeOf(s.GetDestination())
	switch rt.Elem().Kind() {
	case reflect.Slice, reflect.Array:
		err = stmt.SelectContext(ctx, s.GetDestination(), s.GetArgs()...)
		if err != nil {
			return err
		}
		break
	default:
		err = stmt.GetContext(ctx, s.GetDestination(), s.GetArgs()...)
		if err != nil {
			return err
		}
		break
	}

	if s.enableDebug {
		s.log(ctx)
	}

	return err
}

// run Execute the statement without transaction within supplied
// query and update destination if not nil.
func (s *Statement) run(ctx context.Context, db *sqlx.DB) error {
	stmt, err := db.PreparexContext(ctx, s.GetQuery())
	if err != nil {
		return err
	}

	return s.exec(ctx, stmt)
}

// run Execute the statement with transaction within supplied
// query and update destination if not nil.
func (s *Statement) runTx(ctx context.Context, tx *sqlx.Tx) error {
	stmt, err := tx.PreparexContext(ctx, s.GetQuery())
	if err != nil {
		return err
	}

	return s.exec(ctx, stmt)
}

// sync free memory of the destination
func (s *Statement) sync() {
	s.SetDestination(nil)
}

// run multiple statements without transaction.
func run(ctx context.Context, db *sqlx.DB, statements ...*Statement) error {
	for i, statement := range statements {
		err := statement.run(ctx, db)
		if err != nil {
			for j := i; j < 0; i-- {
				statements[j].sync()
			}
			return errors.Errorf("stmt[%d]: %s", i, err.Error())
		}
	}

	return nil
}

// runTx run multiple statements in single transactions.
func runTx(ctx context.Context, tx *sqlx.Tx, statements ...*Statement) error {
	for i, statement := range statements {
		err := statement.runTx(ctx, tx)
		if err != nil {
			for j := i; j < 0; i-- {
				statements[j].sync()
			}
			return errors.Errorf("stmt[%d]: %s", i, err.Error())
		}
	}

	return nil
}

// NewDB create new DB connection.
func NewDB(config *Config) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	conn, err := sqlx.Open(config.Driver, dsn)
	if err != nil {
		return nil, err
	}

	conn.SetConnMaxLifetime(time.Duration(config.MaxLifeTimeConnection) * time.Second)
	conn.SetMaxOpenConns(config.MaxOpenConnections)
	conn.SetMaxIdleConns(config.MaxIdleConnections)
	conn.SetConnMaxIdleTime(time.Duration(config.MaxIdleTimeConnection) * time.Second)

	return &DB{conn: conn}, nil
}

// Exec execute multiple queries or single query without transaction.
func (db *DB) Exec(ctx context.Context, statements ...*Statement) error {
	err := run(ctx, db.conn, statements...)
	if err != nil {
		return err
	}

	return nil
}

// ExecTx execute multiple queries or single query in single transaction.
// TODO: Isolation level configurable per transaction.
func (db *DB) ExecTx(ctx context.Context, statements ...*Statement) error {
	tx, err := db.conn.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	if err = runTx(ctx, tx, statements...); err != nil {
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
	return db.conn.Close()
}
