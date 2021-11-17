package organisation

import (
	"context"
	"errors"

	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbsqlx"
	"github.com/jmoiron/sqlx"
	"github.com/osohq/go-oso"
	"gitlab.com/fynbos/backend/user"
)

type Service interface {
	Create(name string, user user.User) (*Organisation, error)
	Get(id string, user user.User) (*Organisation, error)
	GetAllOwnedBy(user user.User) ([]*Organisation, error)
}

type service struct {
	db    *sqlx.DB
	authz *oso.Oso
}

func NewService(db *sqlx.DB, authz *oso.Oso) (Service, error) {
	if db == nil {
		return nil, errors.New("db cannot be nil.")
	}

	if authz == nil {
		return nil, errors.New("authz cannot be nil.")
	}

	return &service{
		db:    db,
		authz: authz,
	}, nil
}

func (self *service) Create(name string, user user.User) (*Organisation, error) {
	// we have to wrap it in a Cockroach trx helper so that the trx retries are performed correctly.

	var ret Organisation
	err := crdbsqlx.ExecuteTx(context.Background(), self.db, nil, func(tx *sqlx.Tx) error {
		stmt, err := tx.PrepareNamed("INSERT INTO organisations (name, owner_id) VALUES (:name, :ownerid) RETURNING *")
		if err != nil {
			return err
		}

		err = stmt.Stmt.Get(&ret, name, user.ID)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, &FailedToCreateError{
			Err: err,
		}
	}

	return &ret, nil
}

func (self *service) Get(id string, user user.User) (*Organisation, error) {
	org := Organisation{}
	err := self.db.Get(&org, "SELECT * FROM organisations WHERE id=$1 LIMIT 1", id)
	if err != nil {
		return nil, &FailedToGetError{
			Err: err,
		}
	}

	err = self.authz.Authorize(user, "read", org)
	if err != nil {
		return nil, err
	}

	return &org, nil
}

func (self service) GetAllOwnedBy(user user.User) ([]*Organisation, error) {
	var orgs []*Organisation
	err := self.db.Select(&orgs, "SELECT * FROM organisations WHERE owner_id=$1 ORDER BY created_at DESC", user.ID)
	if err != nil {
		return nil, err
	}

	return orgs, nil
}

// Model
type Organisation struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Verified  bool   `json:"verified"`
	OwnerID   string `db:"owner_id"` // for simplicity we store the owner here for now
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

// Error set
type FailedToGetError struct {
	Err error
}

func (r *FailedToGetError) Error() string {
	return r.Err.Error()
}

type FailedToCreateError struct {
	Err error
}

func (r *FailedToCreateError) Error() string {
	return r.Err.Error()
}
