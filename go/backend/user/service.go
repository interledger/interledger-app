package user

import (
	"context"
	"net/http"
	"time"

	kratos "github.com/ory/kratos-client-go"
)

// A private key for context that only this package can access. This is important
// to prevent collisions between different context uses
var userCtxKey = &contextKey{"user"}

type contextKey struct {
	name string
}

type Service interface {
	GetUser(cookie *http.Cookie) (*User, error)
	ForContext(ctx context.Context) (*User, error)
}

type service struct {
	kratos        *kratos.APIClient
	kratosTimeout time.Duration
}

func NewService(kratosClient *kratos.APIClient) (Service, error) {
	return &service{
		kratos:        kratosClient,
		kratosTimeout: 500 * time.Millisecond,
	}, nil
}

// ForContext finds the user from the context. REQUIRES Middleware to have run.
func (self *service) ForContext(ctx context.Context) (*User, error) {
	raw, ok := ctx.Value(userCtxKey).(*User)
	if !ok {
		return nil, NoUserFoundError{}
	}
	return raw, nil
}

// This will call out to Kratos to check if the session is valid.
func (self *service) GetUser(cookie *http.Cookie) (*User, error) {
	if cookie != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(self.kratosTimeout))
		defer cancel()

		session, _, err := self.kratos.V0alpha2Api.ToSession(ctx).Cookie(cookie.Value).Execute()
		if err != nil {
			return nil, err
		}
		user := User{
			ID: session.Identity.Id,
		}
		return &user, nil
	}

	return nil, nil
}

// Model
type User struct {
	ID string
}

// Error set
type NoUserFoundError struct{}

func (r NoUserFoundError) Error() string {
	return "No user found."
}
