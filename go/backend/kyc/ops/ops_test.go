package ops_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/kyc/ops"
	"gitlab.com/fynbos/backend/notify"
	notifymock "gitlab.com/fynbos/backend/notify/client/mock"
)

func TestUpdateUserDetails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, db, nil, nil, nil, nil, nil, nil)

	walletID := uuid.NewString()

	ud, err := ops.UpdateIndividualDetails(ctx, b, kyc.IndividualDetails{
		WalletID:    walletID,
		FirstName:   "InitFirst",
		CountryCode: "ZA",
		Gender:      kyc.GenderMale,
		IPAddress:   "198.16.0.1",
	})
	require.NoError(t, err)

	assert.Equal(t, kyc.GenderMale, ud.Gender)
	assert.Equal(t, "ZA", ud.CountryCode)
	assert.Equal(t, "InitFirst", ud.FirstName)
	assert.Equal(t, "198.16.0.1", ud.IPAddress)
	assert.Empty(t, ud.LastName)
	assert.True(t, ud.DateOfBirth.IsZero())
	assert.Nil(t, ud.Address)

	var revisionCnt int
	err = db.GetContext(ctx, &revisionCnt, "select count(*) from individual_kyc_details where wallet_id=$1", walletID)
	require.NoError(t, err)
	assert.Equal(t, 1, revisionCnt)

	// Now look it up
	ud, err = ops.GetIndividualDetails(ctx, b, walletID)
	require.NoError(t, err)

	assert.Equal(t, kyc.GenderMale, ud.Gender)
	assert.Equal(t, "ZA", ud.CountryCode)
	assert.Equal(t, "InitFirst", ud.FirstName)
	assert.Equal(t, "198.16.0.1", ud.IPAddress)
	assert.Empty(t, ud.LastName)
	assert.True(t, ud.DateOfBirth.IsZero())
	assert.Nil(t, ud.Address)

	// Update
	ud, err = ops.UpdateIndividualDetails(ctx, b, kyc.IndividualDetails{
		WalletID:  walletID,
		FirstName: "Updated",
		LastName:  "New",
		IPAddress: "198.16.0.2",
	})
	require.NoError(t, err)

	assert.Equal(t, kyc.GenderMale, ud.Gender)
	assert.Equal(t, "ZA", ud.CountryCode)
	assert.Equal(t, "Updated", ud.FirstName)
	assert.Equal(t, "New", ud.LastName)
	assert.Equal(t, "198.16.0.2", ud.IPAddress)
	assert.True(t, ud.DateOfBirth.IsZero())
	assert.Nil(t, ud.Address)

	err = db.GetContext(ctx, &revisionCnt, "select count(*) from individual_kyc_details where wallet_id=$1", walletID)
	require.NoError(t, err)
	assert.Equal(t, 2, revisionCnt)

	// Now look it up
	ud, err = ops.GetIndividualDetails(ctx, b, walletID)
	require.NoError(t, err)

	assert.Equal(t, kyc.GenderMale, ud.Gender)
	assert.Equal(t, "ZA", ud.CountryCode)
	assert.Equal(t, "Updated", ud.FirstName)
	assert.Equal(t, "New", ud.LastName)
	assert.Equal(t, "198.16.0.2", ud.IPAddress)
	assert.True(t, ud.DateOfBirth.IsZero())
	assert.Nil(t, ud.Address)

	// Add an address
	ud, err = ops.UpdateIndividualDetails(ctx, b, kyc.IndividualDetails{
		WalletID:  walletID,
		IPAddress: "198.16.0.3",
		Address: &kyc.Address{
			Line1:       "Line1",
			Line2:       "Line2",
			Building:    "Building",
			Apartment:   "Apartment",
			City:        "Cape Town",
			State:       "ZA-WC",
			ZipCode:     "8001",
			CountryCode: "ZA",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, kyc.GenderMale, ud.Gender)
	assert.Equal(t, "ZA", ud.CountryCode)
	assert.Equal(t, "Updated", ud.FirstName)
	assert.Equal(t, "New", ud.LastName)
	assert.Equal(t, "198.16.0.3", ud.IPAddress)
	assert.True(t, ud.DateOfBirth.IsZero())
	require.NotNil(t, ud.Address)
	assert.Equal(t, "Line1", ud.Address.Line1)
	assert.Equal(t, "Line2", ud.Address.Line2)
	assert.Equal(t, "Building", ud.Address.Building)
	assert.Equal(t, "Apartment", ud.Address.Apartment)
	assert.Equal(t, "Cape Town", ud.Address.City)
	assert.Equal(t, "ZA-WC", ud.Address.State)
	assert.Equal(t, "8001", ud.Address.ZipCode)
	assert.Equal(t, "ZA", ud.Address.CountryCode)

	err = db.GetContext(ctx, &revisionCnt, "select count(*) from individual_kyc_details where wallet_id=$1", walletID)
	require.NoError(t, err)
	assert.Equal(t, 3, revisionCnt)

	// Now look it up
	ud, err = ops.GetIndividualDetails(ctx, b, walletID)
	require.NoError(t, err)

	assert.Equal(t, kyc.GenderMale, ud.Gender)
	assert.Equal(t, "ZA", ud.CountryCode)
	assert.Equal(t, "Updated", ud.FirstName)
	assert.Equal(t, "New", ud.LastName)
	assert.Equal(t, "198.16.0.3", ud.IPAddress)
	assert.True(t, ud.DateOfBirth.IsZero())
	require.NotNil(t, ud.Address)
	assert.Equal(t, "Line1", ud.Address.Line1)
	assert.Equal(t, "Line2", ud.Address.Line2)
	assert.Equal(t, "Building", ud.Address.Building)
	assert.Equal(t, "Apartment", ud.Address.Apartment)
	assert.Equal(t, "Cape Town", ud.Address.City)
	assert.Equal(t, "ZA-WC", ud.Address.State)
	assert.Equal(t, "8001", ud.Address.ZipCode)
	assert.Equal(t, "ZA", ud.Address.CountryCode)

	// Noop Update
	_, err = ops.UpdateIndividualDetails(ctx, b, kyc.IndividualDetails{
		WalletID:    walletID,
		FirstName:   "Updated",
		LastName:    "New",
		CountryCode: "ZA",
		IPAddress:   "198.16.0.3",
		Gender:      kyc.GenderMale,
	})
	require.NoError(t, err)

	err = db.GetContext(ctx, &revisionCnt, "select count(*) from individual_kyc_details where wallet_id=$1", walletID)
	require.NoError(t, err)
	assert.Equal(t, 3, revisionCnt)
}

func TestGetKYCStatusMetadata_ReturnsUnknownWhenRowMissing(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	b := ops.NewTestBackends(t, db, nil, nil, nil, nil, nil, nil)

	meta, err := ops.GetKYCStatusMetadata(ctx, b, uuid.NewString())
	require.NoError(t, err)
	require.Equal(t, kyc.StatusUnknown, meta.Status)
	require.Zero(t, meta.ResubmissionCount)
	require.Nil(t, meta.ExpirationDate)
}

func TestGetKYCStatusMetadata_ReturnsStoredMetadata(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	b := ops.NewTestBackends(t, db, nil, nil, nil, nil, nil, nil)

	walletID := uuid.NewString()
	expirationDate := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	_, err := db.ExecContext(ctx, `
		INSERT INTO wallet_kyc_status (wallet_id, status, status_reason, last_webhook_event_type, resubmission_count, document_expiration_date)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, walletID, kyc.StatusDocumentsRequired, "Photo is blurry", "id.verification.action_required", 2, expirationDate)
	require.NoError(t, err)

	meta, err := ops.GetKYCStatusMetadata(ctx, b, walletID)
	require.NoError(t, err)
	require.Equal(t, kyc.StatusDocumentsRequired, meta.Status)
	require.Equal(t, "Photo is blurry", meta.Reason)
	require.Equal(t, "id.verification.action_required", meta.LastWebhookEvent)
	require.EqualValues(t, 2, meta.ResubmissionCount)
	require.NotNil(t, meta.ExpirationDate)
	require.Equal(t, expirationDate.Format("2006-01-02"), meta.ExpirationDate.Format("2006-01-02"))
}

func TestCanResubmitKYC(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	b := ops.NewTestBackends(t, db, nil, nil, nil, nil, nil, nil)

	walletID := uuid.NewString()

	canResubmit, err := ops.CanResubmitKYC(ctx, b, walletID)
	require.NoError(t, err)
	require.True(t, canResubmit)

	_, err = db.ExecContext(ctx, "INSERT INTO wallet_kyc_status (wallet_id, status) VALUES ($1, $2)", walletID, kyc.StatusApproved)
	require.NoError(t, err)

	canResubmit, err = ops.CanResubmitKYC(ctx, b, walletID)
	require.NoError(t, err)
	require.False(t, canResubmit)
}

func TestActivityUpdateKYCStatus_StoresMetadataWhenEventTypeProvided(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	database := db.MigrateTestDB(t, ctx)
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	notifyClient := notifymock.NewMockClient(ctrl)
	b := ops.NewTestBackends(t, database, nil, nil, nil, notifyClient, nil, nil)
	activity := ops.NewActivity(b)
	walletID := uuid.NewString()

	notifyClient.EXPECT().NotifyWallet(gomock.Any(), walletID, notify.NotificationType("kyc")).Return(nil).Times(1)

	err := activity.UpdateKYCStatus(ctx, walletID, kyc.StatusDocumentsRequired, "Photo is blurry", "id.verification.action_required")
	require.NoError(t, err)

	meta, err := ops.GetKYCStatusMetadata(ctx, b, walletID)
	require.NoError(t, err)
	require.Equal(t, kyc.StatusDocumentsRequired, meta.Status)
	require.Equal(t, "Photo is blurry", meta.Reason)
	require.Equal(t, "id.verification.action_required", meta.LastWebhookEvent)
	require.EqualValues(t, 1, meta.ResubmissionCount)

	notifyClient.EXPECT().NotifyWallet(gomock.Any(), walletID, notify.NotificationType("kyc")).Return(nil).Times(1)

	err = activity.UpdateKYCStatus(ctx, walletID, kyc.StatusDocumentsRequired, "Need proof of address", "id.document_notice.expired")
	require.NoError(t, err)

	meta, err = ops.GetKYCStatusMetadata(ctx, b, walletID)
	require.NoError(t, err)
	require.Equal(t, "Need proof of address", meta.Reason)
	require.Equal(t, "id.document_notice.expired", meta.LastWebhookEvent)
	require.EqualValues(t, 2, meta.ResubmissionCount)
}

func TestActivityUpdateKYCStatus_PreservesMetadataWithoutEventType(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	database := db.MigrateTestDB(t, ctx)
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	notifyClient := notifymock.NewMockClient(ctrl)
	b := ops.NewTestBackends(t, database, nil, nil, nil, notifyClient, nil, nil)
	activity := ops.NewActivity(b)
	walletID := uuid.NewString()

	_, err := database.ExecContext(ctx, `
		INSERT INTO wallet_kyc_status (wallet_id, status, status_reason, last_webhook_event_type, resubmission_count)
		VALUES ($1, $2, $3, $4, $5)
	`, walletID, kyc.StatusDocumentsRequired, "Initial reason", "id.verification.action_required", 3)
	require.NoError(t, err)

	notifyClient.EXPECT().NotifyWallet(gomock.Any(), walletID, notify.NotificationType("kyc")).Return(nil).Times(1)

	err = activity.UpdateKYCStatus(ctx, walletID, kyc.StatusPending, "", "")
	require.NoError(t, err)

	meta, err := ops.GetKYCStatusMetadata(ctx, b, walletID)
	require.NoError(t, err)
	require.Equal(t, kyc.StatusPending, meta.Status)
	require.Equal(t, "Initial reason", meta.Reason)
	require.Equal(t, "id.verification.action_required", meta.LastWebhookEvent)
	require.EqualValues(t, 3, meta.ResubmissionCount)
}
