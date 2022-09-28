package dev

import (
	"context"
	"reflect"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/providers/machnet/external"
)

var _ external.Client = Client{}

func New() *Client {
	return &Client{users: map[string]external.User{}}
}

type Client struct {
	users map[string]external.User
}

func (c Client) RegisterUser(ctx context.Context, user external.User) (*external.User, error) {
	ret := user
	ret.ID = uuid.NewString()
	ret.Status = external.StatusUnverified
	c.users[ret.ID] = ret

	return &ret, nil
}

func (c Client) UpdateUser(ctx context.Context, id string, newValues external.User) (*external.User, error) {
	user, found := c.users[id]
	if !found {
		return nil, external.ErrNotFound
	}

	v := reflect.ValueOf(newValues)
	merged := reflect.ValueOf(&user).Elem()
	for i, n := 0, v.NumField(); i < n; i++ {
		val := v.Field(i)
		// Update if field is not empty
		if !reflect.DeepEqual(val.Interface(), reflect.Zero(v.Field(i).Type()).Interface()) {
			merged.Field(i).Set(v.Field(i))
		}
	}

	c.users[id] = user

	return &user, nil
}

func (c Client) GetUserByID(ctx context.Context, id string) (*external.User, error) {
	user, found := c.users[id]
	if !found {
		return nil, external.ErrNotFound
	}

	return &user, nil
}
