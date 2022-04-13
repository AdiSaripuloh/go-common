package db

import (
	"context"
	"reflect"
	"sync"

	"github.com/AdiSaripuloh/go-common/logger"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// A Statement is a simple wrapper for creating a statement consisting of
// a query and a set of arguments to be passed to that query.
type Statement struct {
	dest        any // if query doesn't have any result, leave it nil.
	query       string
	args        []any
	enableDebug bool
	mu          *sync.Mutex
}

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
