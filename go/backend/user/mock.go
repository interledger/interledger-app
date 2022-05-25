package user

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/machinebox/graphql"
	"google.golang.org/grpc/metadata"
)

type mockService struct{}

var _testCookieName string = "test-cookie"

func (self *mockService) GetUser(r http.Request) (*User, error) {
	c, err := r.Cookie(_testCookieName)
	if err != nil || c == nil {
		return nil, ErrNoUserFound
	}

	user := User{}
	unescapedCookie, err := url.QueryUnescape(c.Value)
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
		return nil, ErrNoUserFound
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
			Name:  _testCookieName,
			Value: url.QueryEscape(string(b)),
		}
		req.Header.Add("Cookie", cookie.String())

		return nil
	}

	return nil
}

func ActingAsContext(t *testing.T, ctx context.Context, user *User) context.Context {
	if user != nil {
		b, err := json.Marshal(user)
		if err != nil {
			t.Fatal(err)
		}

		cookie := http.Cookie{
			Name:  _testCookieName,
			Value: url.QueryEscape(string(b)),
		}
		return metadata.AppendToOutgoingContext(ctx, cookieMetadataKey, cookie.String())
	}

	return ctx
}

func NewMockService() Service {
	return &mockService{}
}
