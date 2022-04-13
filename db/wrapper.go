package db

import (
	"context"

	"github.com/jmoiron/sqlx"
	"go.mongodb.org/mongo-driver/mongo"
)

// Exec wrapping multiple queries or single query without transaction.
func Exec(ctx context.Context, db *sqlx.DB, statements ...*Statement) error {
	err := run(ctx, db, statements...)
	if err != nil {
		return err
	}

	return nil
}

// ExecTx wrapping multiple queries or single query in a transaction.
// TODO: Isolation level configurable per transaction.
func ExecTx(ctx context.Context, db *sqlx.DB, statements ...*Statement) error {
	tx, err := db.BeginTxx(ctx, nil)
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

// SessionPipeline should executes multiple mongo queries.
// example:
// func() error {
//    ...
//    err := mongoDB.Collection("example").FindOne(ctx, nil).Decode(&v)
//    ...
//    return err
// }
type SessionPipeline func() error

// ExecSession wrapping multiple queries or single query of mongo in a session.
func ExecSession(ctx context.Context, db *mongo.Database, sp SessionPipeline) error {
	session, err := db.Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	err = session.StartTransaction()
	if err != nil {
		return err
	}

	return mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		er := sp()
		if er != nil {
			return er
		}

		return session.CommitTransaction(sc)
	})
}
