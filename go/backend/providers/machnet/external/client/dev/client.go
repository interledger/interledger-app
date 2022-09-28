package dev

import (
	"context"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/providers/machnet/external"
)

var _ external.Client = client{}

type client struct {
	users map[string]external.User
}

func (c client) RegisterUser(ctx context.Context, user external.User) (*external.User, error) {
	ret := user
	ret.ID = uuid.NewString()
	ret.Status = external.StatusUnverified
	c.users[ret.ID] = ret

	return &ret, nil
}
