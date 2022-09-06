package activities

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/twilio"

	"gitlab.com/fynbos/backend/providers/mx"
	mx_mock "gitlab.com/fynbos/backend/providers/mx/client/mock"
	identity_mock "gitlab.com/fynbos/backend/identity/client/mock"
	funding_mock "gitlab.com/fynbos/backend/fundingsources/client/mock"
	"github.com/bxcodec/faker/v3"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/accounts"
	accounts_mock "gitlab.com/fynbos/backend/accounts/client/mock"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/identity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"gitlab.com/fynbos/backend/providers/mx/external"
	"gitlab.com/fynbos/backend/providers/unit"
	unit_mock "gitlab.com/fynbos/backend/providers/unit/client/mock"
)

func TestCreateFundingsource(t *testing.T) {
	ctx := context.Background()
	activity, b := NewTestActivity(t)
	mxAccountGuid := "acct_" + uuid.NewString()
	userID := uuid.NewString()
	accountID := uuid.NewString()
	fundingsourceID := uuid.NewString()

	b.mx.EXPECT().GetAccount(ctx, mxAccountGuid).Return(
		&mx.Account{
			Guid:            mxAccountGuid,
			AccountID:       accountID,
			FundingsourceID: fundingsourceID,
		},
		nil,
	).Times(1)
	b.acc.EXPECT().Get(ctx, accountID).Return(
		&accounts.Account{
			ID:         accountID,
			IdentityID: userID,
		},
		nil,
	).Times(1)
	b.ident.EXPECT().Get(ctx, userID).Return(
		&identity.Identity{
			ID: userID,
		},
		nil,
	).Times(1)
	b.mx.EXPECT().ReadAccount(ctx, mxAccountGuid).Return(
		&mx.AccountDetails{
			Guid:          mxAccountGuid,
			AccountNumber: "81818181234", // will be used to set the mask on the funding source
		},
		nil,
	).Times(1)

	fundingsourceName := "test-mx"
	b.fs.EXPECT().Create(ctx, &fundingsources.CreateArgs{
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
	activity, b := NewTestActivity(t)
	mxAccountGuid := "acct_" + uuid.NewString()
	userID := uuid.NewString()
	firstName := faker.FirstName()
	lastName := faker.LastName()
	accountID := uuid.NewString()
	fundingsourceID := uuid.NewString()

	b.mx.EXPECT().GetAccount(ctx, mxAccountGuid).Return(
		&mx.Account{
			Guid:            mxAccountGuid,
			AccountID:       accountID,
			FundingsourceID: fundingsourceID,
		},
		nil,
	).Times(1)
	b.acc.EXPECT().Get(ctx, accountID).Return(
		&accounts.Account{
			ID:         accountID,
			IdentityID: userID,
		},
		nil,
	).Times(1)
	b.ident.EXPECT().Get(ctx, userID).Return(
		&identity.Identity{
			ID:        userID,
			FirstName: firstName,
			LastName:  lastName,
		},
		nil,
	).Times(1)

	unitCustomerID := "8"
	b.unit.EXPECT().GetCustomerByIdentityID(ctx, userID).Return(
		&unit.Customer{
			ID:         unitCustomerID,
			IdentityID: userID,
			Type:       "person",
		},
		nil,
	)

	b.mx.EXPECT().ReadAccount(ctx, mxAccountGuid).Return(
		&mx.AccountDetails{
			Guid:              mxAccountGuid,
			AccountNumber:     "81818181234", // will be used to set the mask on the funding source
			RoutingNumber:     "71717171717",
			InstitutionNumber: "616161616",
			Type:              "SAVINGS",
		},
		nil,
	).Times(1)
	idempotencyKey := sha256.Sum256([]byte(fundingsourceID))
	b.unit.EXPECT().CreateCounterParty(ctx, &unit.CreateCounterPartyArgs{
		FundingsourceID: fundingsourceID,
		Name:            fmt.Sprintf("%s %s", firstName, lastName),
		UnitCustomerID:  unitCustomerID,
		RoutingNumber:   "71717171717",
		AccountNumber:   "81818181234",
		AccountType:     "Savings",
		Type:            "Person",
		IdempotencyKey:  string(idempotencyKey[0:]),
	})

	err := activity.CreateUnitCounterParty(ctx, mxAccountGuid)

	assert.NoError(t, err)
}

func TestWaitForAggregation(t *testing.T) {
	ctx := context.Background()
	activity, b := NewTestActivity(t)
	mxUserGuid := "usr_" + uuid.NewString()
	mxMemberGuid := "mbr_" + uuid.NewString()

	b.mx.EXPECT().GetMemberStatus(ctx, mxUserGuid, mxMemberGuid).Return(nil, mx.ErrInternal).Times(2)
	b.mx.EXPECT().GetMemberStatus(ctx, mxUserGuid, mxMemberGuid).Return(
		&mx.Member{
			Guid:              "mbr_" + uuid.NewString(),
			UserGuid:          "usr_" + uuid.NewString(),
			IsBeingAggregated: false,
		},
		nil,
	).Times(1)

	err := activity.WaitForAggregation(ctx, &WaitForAggregationArgs{
		MxMemberGuid: mxMemberGuid,
		MxUserGuid:   mxUserGuid,
		MaxRetries:   2,
		PollInterval: 10 * time.Millisecond,
	})

	assert.NoError(t, err)
}

func TestStartBalanceAggregation(t *testing.T) {
	ctx := context.Background()
	activity, b := NewTestActivity(t)
	mxAccountGuid := "acct_" + uuid.NewString()

	t.Run("returns non retryable error if mx account not found", func(st *testing.T) {
		b.mx.EXPECT().StartBalanceAggregation(ctx, mxAccountGuid).Return(nil, mx.ErrNotFound).Times(1)

		err := activity.StartBalanceAggregation(ctx, mxAccountGuid)

		assert.True(st, temporal.IsApplicationError(err))
		applicationError := err.(*temporal.ApplicationError)
		assert.True(st, applicationError.NonRetryable())
	})

	t.Run("returns non retryable error if member status indicates it cannot be aggregated", func(st *testing.T) {
		b.mx.EXPECT().StartBalanceAggregation(ctx, mxAccountGuid).Return(
			&mx.Member{
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

func NewTestActivity(t *testing.T) (*Activity, *testBackends) {
	ctrl := gomock.NewController(t)
	mocks := &MockArgs{
		Mx:                   NewMockService(ctrl),
		Unit:                 unit_mock.NewMockClient(ctrl),
		AccountService:       accounts_mock.NewMockClient(ctrl),
		IdentityService:      identity_mock.NewMockClient(ctrl),
		FundingsourceService: funding_mock.NewMockClient(ctrl),
	}

	activity := NewActivity(b)

	return activity, b
}

type testBackends struct {
	val   *validator.Validate
	mx    *mx_mock.MockClient
	unit  *unit_mock.MockClient
	acc   *accounts_mock.MockClient
	ident *identity_mock.MockClient
	fs    *funding_mock.MockClient
}

func (t testBackends) Validator() *validator.Validate {
	return t.val
}

func (t testBackends) DB() *sqlx.DB {
	return nil
}

func (t testBackends) Accounts() accounts.Client {
	return t.acc
}

func (t testBackends) Identity() identity.Client {
	return t.ident
}

func (t testBackends) Temporal() client.Client {
	return nil
}

func (t testBackends) Twilio() twilio.Service {
	return nil
}

func (t testBackends) MX() mx.Client {
	return t.mx
}

func (t testBackends) Unit() unit.Service {
	return t.unit
}

func (t testBackends) FundingSources() fundingsources.Client {
	return t.fs
}
