package mx

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	mockMxServer := NewMockServer()
	mx, err := NewService(&ServiceArgs{
		BaseUrl:  mockMxServer.URL,
		Username: "test",
		Password: "test",
	})
	if err != nil {
		t.Fatal(err)
	}

	userGuid, err := mx.CreateUser(ctx)
	if err != nil {
		t.Fatal(err)
	}

	_ = uuid.MustParse(userGuid)
}

func TestGetWidgetUrl(t *testing.T) {
	ctx := context.Background()
	mockMxServer := NewMockServer()
	mx, err := NewService(&ServiceArgs{
		BaseUrl:  mockMxServer.URL,
		Username: "test",
		Password: "test",
	})
	if err != nil {
		t.Fatal(err)
	}

	url, err := mx.GetWidgetUrl(ctx, uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}

	assert.NotEqual(t, "", url)
}
