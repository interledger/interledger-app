package services

import (
	"context"
	"errors"

	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbsqlx"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/models"
)

type Organisations struct {
	Db *sqlx.DB
}

func NewOrganisationsService(db *sqlx.DB) (*Organisations, error) {
	if db == nil {
		return nil, errors.New("db cannot be nil.")
	}
	return &Organisations{
		Db: db,
	}, nil
}

func (self *Organisations) Create(name string) (*models.Organisation, error) {
	// we have to wrap it in a Cockroach trx helper so that the trx retries are performed correctly.

	var ret models.Organisation
	err := crdbsqlx.ExecuteTx(context.Background(), self.Db, nil, func(tx *sqlx.Tx) error {
		stmt, err := tx.PrepareNamed("INSERT INTO organisations (name) VALUES (:name) RETURNING id, name")
		if err != nil {
			return err
		}
		err = stmt.Stmt.Get(&ret, name)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (self *Organisations) Get(id string) (*models.Organisation, error) {
	org := models.Organisation{}
	err := self.Db.Get(&org, "SELECT * FROM organisations WHERE id=$1 LIMIT 1", id)
	if err != nil {
		return nil, err
	}

	return &org, nil
}
