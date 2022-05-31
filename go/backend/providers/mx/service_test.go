package mx

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	test_utils "gitlab.com/fynbos/backend/utils"
)

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	mockMxServer := NewMockServer()
	mx, err := NewService(&ServiceArgs{
		BaseUrl:  mockMxServer.URL,
		Username: "test",
		Password: "test",
		Db:       &sqlx.DB{},
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
		Db:       &sqlx.DB{},
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

func TestCreateAndGetAccount(t *testing.T) {
	ctx := context.Background()
	db, dbCleanup := test_utils.MigrateCockroachDB(t, ctx)
	t.Cleanup(func() {
		dbCleanup()
	})
	mockMxServer := NewMockServer()
	mx, err := NewService(&ServiceArgs{
		BaseUrl:  mockMxServer.URL,
		Username: "test",
		Password: "test",
		Db:       db,
	})
	if err != nil {
		t.Fatal(err)
	}
	args := &CreateMxFundingSourceArgs{
		ID:            uuid.NewString(),
		AccountID:     uuid.NewString(),
		MxUserGuid:    uuid.NewString(),
		MxMemberGuid:  uuid.NewString(),
		MxAccountGuid: uuid.NewString(),
	}

	mxFs, err := mx.CreateMxFundingSource(ctx, args)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, args.ID, mxFs.ID)
	assert.Equal(t, args.AccountID, mxFs.AccountID)
	assert.Equal(t, args.MxUserGuid, mxFs.MxUserGuid)
	assert.Equal(t, args.MxMemberGuid, mxFs.MxMemberGuid)

	freshMxFs, err := mx.GetMxFundingSource(ctx, args.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, args.ID, freshMxFs.ID)
	assert.Equal(t, args.AccountID, freshMxFs.AccountID)
	assert.Equal(t, args.MxUserGuid, freshMxFs.MxUserGuid)
	assert.Equal(t, args.MxMemberGuid, freshMxFs.MxMemberGuid)

	noAcc, err := mx.GetMxFundingSource(ctx, uuid.NewString())
	assert.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, noAcc)

	// idempotency
	idempotent, err := mx.CreateMxFundingSource(ctx, args)
	assert.ErrorIs(t, err, ErrDuplicate)
	assert.Nil(t, idempotent)
}

func TestGetMemberStatus(t *testing.T) {
	ctx := context.Background()
	mockMxServer := NewMockServer()
	db, dbCleanup := test_utils.MigrateCockroachDB(t, ctx)
	t.Cleanup(func() {
		dbCleanup()
	})
	mx, err := NewService(&ServiceArgs{
		BaseUrl:  mockMxServer.URL,
		Username: "test",
		Password: "test",
		Db:       db,
	})
	if err != nil {
		t.Fatal(err)
	}

	mxFs, err := mx.CreateMxFundingSource(ctx, &CreateMxFundingSourceArgs{
		ID:            uuid.NewString(),
		AccountID:     uuid.NewString(),
		MxUserGuid:    uuid.NewString(),
		MxMemberGuid:  uuid.NewString(),
		MxAccountGuid: uuid.NewString(),
	})
	if err != nil {
		t.Fatal(err)
	}

	member, err := mx.GetMemberStatus(ctx, mxFs.ID)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, mxFs.MxUserGuid, member.UserGuid)
	assert.Equal(t, mxFs.MxMemberGuid, member.Guid)
	assert.Equal(t, false, member.IsBeingAggregated)
}

func TestStartIdentityAggregation(t *testing.T) {
	ctx := context.Background()
	mockMxServer := NewMockServer()
	db, dbCleanup := test_utils.MigrateCockroachDB(t, ctx)
	t.Cleanup(func() {
		dbCleanup()
	})
	mx, err := NewService(&ServiceArgs{
		BaseUrl:  mockMxServer.URL,
		Username: "test",
		Password: "test",
		Db:       db,
	})
	if err != nil {
		t.Fatal(err)
	}

	mxFs, err := mx.CreateMxFundingSource(ctx, &CreateMxFundingSourceArgs{
		ID:            uuid.NewString(),
		AccountID:     uuid.NewString(),
		MxUserGuid:    uuid.NewString(),
		MxMemberGuid:  uuid.NewString(),
		MxAccountGuid: uuid.NewString(),
	})
	if err != nil {
		t.Fatal(err)
	}

	member, err := mx.StartIdentityAggregation(ctx, mxFs.ID)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, mxFs.MxUserGuid, member.UserGuid)
	assert.Equal(t, mxFs.MxMemberGuid, member.Guid)
	assert.Equal(t, true, member.IsBeingAggregated)
}
