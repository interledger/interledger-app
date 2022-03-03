package user

import (
	"context"
	"errors"
	"net/http"
	"time"

	kratos "github.com/ory/kratos-client-go"
)

// A private key for context that only this package can access. This is important
// to prevent collisions between different context uses
var userCtxKey = &contextKey{"user"}

var ErrNoUserFound = errors.New("no user found")

type contextKey struct {
	name string
}

type Service interface {
	GetUser(request http.Request) (*User, error)
	ForContext(ctx context.Context) (*User, error)
}

type service struct {
	kratos           *kratos.APIClient
	kratosTimeout    time.Duration
	kratosCookieName string
}

func NewService(kratosClient *kratos.APIClient) (Service, error) {
	return &service{
		kratos:           kratosClient,
		kratosTimeout:    500 * time.Millisecond,
		kratosCookieName: "ory_kratos_session",
	}, nil
}

// ForContext finds the user from the context. REQUIRES Middleware to have run.
func (self *service) ForContext(ctx context.Context) (*User, error) {
	raw, ok := ctx.Value(userCtxKey).(*User)
	if !ok || raw == nil {
		return nil, ErrNoUserFound
	}
	return raw, nil
}

// GetUser will either return the user
func (self *service) GetUser(r http.Request) (*User, error) {
	c, err := r.Cookie(self.kratosCookieName)
	if err != nil || c == nil {
		return nil, ErrNoUserFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), self.kratosTimeout)
	defer cancel()

	session, resp, err := self.kratos.V0alpha2Api.ToSession(ctx).Cookie(self.kratosCookieName + "=" + c.Value).Execute()
	if err != nil {
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, nil
		}
		return nil, err
	}
	traits := session.Identity.Traits.(map[string]interface{})
	user := User{
		ID:    session.Identity.Id,
		Email: traits["email"].(string),
	}
	return &user, nil
}

// Model
type User struct {
	ID    string
	Email string
}

type traits struct {
	Email string
}
