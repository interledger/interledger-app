package user

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/machinebox/graphql"
)

type mockService struct{}

func (self *mockService) GetUser(cookie *http.Cookie) (*User, error) {
	user := User{}
	unescapedCookie, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(unescapedCookie), &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// ForContext finds the user from the context. REQUIRES Middleware to have run.
func (self *mockService) ForContext(ctx context.Context) (*User, error) {
	raw, ok := ctx.Value(userCtxKey).(*User)
	if !ok {
		return nil, &NoUserFoundError{}
	}
	return raw, nil
}

// Test helper function to set the cookie in the graphql request.
func ActingAs(req *graphql.Request, user *User) error {
	if user != nil {
		b, err := json.Marshal(user)
		if err != nil {
			return err
		}

		cookie := http.Cookie{
			Name:  "cookie",
			Value: url.QueryEscape(string(b)),
		}
		req.Header.Set("cookie", cookie.String())

		return nil
	}

	return nil
}

func NewMockService() Service {
	return &mockService{}
}
