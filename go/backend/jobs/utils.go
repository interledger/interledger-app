package jobs

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	"github.com/uptrace/opentelemetry-go-extra/otelsqlx"
	"gitlab.com/fynbos/log"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/zap"
)

func ExecuteTransaction(db *sqlx.DB, queries []struct {
	query string
	args  []interface{}
}) (err error) {
	tx, err := db.BeginTxx(context.Background(), nil)
	if err != nil {
		log.Error("Error starting transaction", zap.Error(err))
		return err
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Error("tx.Rollback failed", zap.Error(rbErr))
			}
			log.Error("Error executing query", zap.Error(err))
		}
	}()

	for _, q := range queries {
		if _, err = tx.Exec(q.query, q.args...); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		log.Error("Error committing transaction", zap.Error(err))
		return err
	}

	return nil
}

func DbConnection(connString string) (*sqlx.DB, error) {
	if connString == "" {
		log.Error("DB connection string is empty")
		return nil, errors.New("DB connection string is empty")
	}
	log.Info("Establishing db connection")
	db, err := otelsqlx.Connect("postgres", connString, otelsql.WithAttributes(semconv.DBSystemCockroachdb), otelsql.WithDBName("cockroachdb"))
	if err != nil {
		log.Error("Failed to connect to the database", zap.Error(err))
		return nil, err
	}
	return db, nil
}
