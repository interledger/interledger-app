package mx

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/identity"
	"go.temporal.io/sdk/temporal"

	"gitlab.com/fynbos/backend/providers/mx/external"
	"gitlab.com/fynbos/backend/providers/unit"
)

func TestCreateFundingsource(t *testing.T) {
	ctx := context.Background()
	activity, mocks := NewTestActivity(t)
	mxAccountGuid := "acct_" + uuid.NewString()
	userID := uuid.NewString()
	accountID := uuid.NewString()
	fundingsourceID := uuid.NewString()

	mocks.Mx.EXPECT().GetAccount(ctx, mxAccountGuid).Return(
		&Account{
			Guid:            mxAccountGuid,
			AccountID:       accountID,
			FundingsourceID: fundingsourceID,
		},
		nil,
	).Times(1)
	mocks.AccountService.EXPECT().Get(ctx, accountID).Return(
		&accounts.Account{
			ID:         accountID,
			IdentityID: userID,
		},
		nil,
	).Times(1)
	mocks.IdentityService.EXPECT().Get(ctx, userID).Return(
		&identity.Identity{
			ID: userID,
		},
		nil,
	).Times(1)
	mocks.Mx.EXPECT().ReadAccount(ctx, mxAccountGuid).Return(
		&AccountDetails{
			Guid:          mxAccountGuid,
			AccountNumber: "81818181234", // will be used to set the mask on the funding source
		},
		nil,
	).Times(1)

	fundingsourceName := "test-mx"
	mocks.FundingsourceService.EXPECT().Create(ctx, &fundingsources.CreateArgs{
		ID:                fundingsourceID,
		IdentityID:        userID,
		AccountID:         accountID,
		Name:              fundingsourceName,
		Mask:              "1234", // last 4
		VerificationState: string(fundingsources.VERIFIED),
		Type:              "mx",
		SubType:           "bank",
	}).Return(
		&fundingsources.FundingSource{
			ID: fundingsourceID,
		},
		nil,
	).Times(1)

	err := activity.CreateFundingSource(ctx, mxAccountGuid, fundingsourceName)

	assert.NoError(t, err)
}

func TestCreateUnitCounterparty(t *testing.T) {
	ctx := context.Background()
	activity, mocks := NewTestActivity(t)
	mxAccountGuid := "acct_" + uuid.NewString()
	userID := uuid.NewString()
	firstName := faker.FirstName()
	lastName := faker.LastName()
	accountID := uuid.NewString()
	fundingsourceID := uuid.NewString()

	mocks.Mx.EXPECT().GetAccount(ctx, mxAccountGuid).Return(
		&Account{
			Guid:            mxAccountGuid,
			AccountID:       accountID,
			FundingsourceID: fundingsourceID,
		},
		nil,
	).Times(1)
	mocks.AccountService.EXPECT().Get(ctx, accountID).Return(
		&accounts.Account{
			ID:         accountID,
			IdentityID: userID,
		},
		nil,
	).Times(1)
	mocks.IdentityService.EXPECT().Get(ctx, userID).Return(
		&identity.Identity{
			ID:        userID,
			FirstName: firstName,
			LastName:  lastName,
		},
		nil,
	).Times(1)

	unitCustomerID := "8"
	mocks.Unit.EXPECT().GetCustomerByIdentityID(ctx, userID).Return(
		&unit.Customer{
			ID:         unitCustomerID,
			IdentityID: userID,
			Type:       "person",
		},
		nil,
	)

	mocks.Mx.EXPECT().ReadAccount(ctx, mxAccountGuid).Return(
		&AccountDetails{
			Guid:              mxAccountGuid,
			AccountNumber:     "81818181234", // will be used to set the mask on the funding source
			RoutingNumber:     "71717171717",
			InstitutionNumber: "616161616",
			Type:              "SAVINGS",
		},
		nil,
	).Times(1)
	idempotencyKey := sha256.Sum256([]byte(fundingsourceID))
	mocks.Unit.EXPECT().CreateCounterParty(ctx, &unit.CreateCounterPartyArgs{
		FundingsourceID: fundingsourceID,
		Name:            fmt.Sprintf("%s %s", firstName, lastName),
		UnitCustomerID:  unitCustomerID,
		RoutingNumber:   "71717171717",
		AccountNumber:   "81818181234",
		AccountType:     "SAVINGS",
		Type:            "person",
		IdempotencyKey:  string(idempotencyKey[0:]),
	})

	err := activity.CreateUnitCounterParty(ctx, mxAccountGuid)

	assert.NoError(t, err)
}

func TestWaitForAggregation(t *testing.T) {
	ctx := context.Background()
	activity, mocks := NewTestActivity(t)
	mxAccountGuid := "acct_" + uuid.NewString()

	mocks.Mx.EXPECT().GetMemberStatus(ctx, mxAccountGuid).Return(nil, ErrInternal).Times(2)
	mocks.Mx.EXPECT().GetMemberStatus(ctx, mxAccountGuid).Return(
		&Member{
			Guid:              "mbr_" + uuid.NewString(),
			UserGuid:          "usr_" + uuid.NewString(),
			IsBeingAggregated: false,
		},
		nil,
	).Times(1)

	err := activity.WaitForAggregation(ctx, mxAccountGuid, 2, 10*time.Millisecond)

	assert.NoError(t, err)
}

func TestStartBalanceAggregation(t *testing.T) {
	ctx := context.Background()
	activity, mocks := NewTestActivity(t)
	mxAccountGuid := "acct_" + uuid.NewString()

	t.Run("returns non retryable error if mx account not found", func(st *testing.T) {
		mocks.Mx.EXPECT().StartBalanceAggregation(ctx, mxAccountGuid).Return(nil, ErrNotFound).Times(1)

		err := activity.StartBalanceAggregation(ctx, mxAccountGuid)

		assert.True(st, temporal.IsApplicationError(err))
		applicationError := err.(*temporal.ApplicationError)
		assert.True(st, applicationError.NonRetryable())
	})

	t.Run("returns non retryable error if member status indicates it cannot be aggregated", func(st *testing.T) {
		mocks.Mx.EXPECT().StartBalanceAggregation(ctx, mxAccountGuid).Return(
			&Member{
				IsBeingAggregated: false,
				ConnectionStatus:  external.CONNECTION_STATUS_CHALLENGED,
			},
			nil,
		).Times(1)

		err := activity.StartBalanceAggregation(ctx, mxAccountGuid)

		assert.True(st, temporal.IsApplicationError(err))
		applicationError := err.(*temporal.ApplicationError)
		assert.True(st, applicationError.NonRetryable())
	})
}

type MockArgs struct {
	Mx                   *MockService
	Unit                 *unit.MockService
	AccountService       *accounts.MockService
	IdentityService      *identity.MockService
	FundingsourceService *fundingsources.MockService
}

func NewTestActivity(t *testing.T) (*Activity, *MockArgs) {
	ctrl := gomock.NewController(t)
	mocks := &MockArgs{
		Mx:                   NewMockService(ctrl),
		Unit:                 unit.NewMockService(ctrl),
		AccountService:       accounts.NewMockService(ctrl),
		IdentityService:      identity.NewMockService(ctrl),
		FundingsourceService: fundingsources.NewMockService(ctrl),
	}

	activity, err := NewActivity(&ActivityArgs{
		Mx:                   mocks.Mx,
		Unit:                 mocks.Unit,
		AccountService:       mocks.AccountService,
		IdentityService:      mocks.IdentityService,
		FundingSourceService: mocks.FundingsourceService,
	})
	if err != nil {
		t.Fatal(err)
	}

	return activity, mocks
}
