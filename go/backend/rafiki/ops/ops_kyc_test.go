package ops

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/backend/kyc"
	kyc_client_mock "gitlab.com/fynbos/backend/kyc/client/mock"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/chimoney"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/providers/xago"
	"gitlab.com/fynbos/backend/rafiki/external"
	"gitlab.com/fynbos/backend/wallets"
	temporal "go.temporal.io/sdk/client"
)

type kycOnlyBackends struct {
	kycClient kyc.Client
}

func (b kycOnlyBackends) DB() *sqlx.DB                          { return nil }
func (b kycOnlyBackends) External() external.Client             { return nil }
func (b kycOnlyBackends) Payments() payments.Client             { return nil }
func (b kycOnlyBackends) Temporal() temporal.Client             { return nil }
func (b kycOnlyBackends) LinkedAccounts() linkedaccounts.Client { return nil }
func (b kycOnlyBackends) Wallets() wallets.Client               { return nil }
func (b kycOnlyBackends) Keys() keys.Client                     { return nil }
func (b kycOnlyBackends) PTI() pti.Client                       { return nil }
func (b kycOnlyBackends) Gatehub() gatehub.Client               { return nil }
func (b kycOnlyBackends) Xago() xago.Client                     { return nil }
func (b kycOnlyBackends) Chimoney() chimoney.Client             { return nil }
func (b kycOnlyBackends) KYC() kyc.Client                       { return b.kycClient }

func TestGetWalletAddress_DocumentsRequiredBlocked(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	kycClient := kyc_client_mock.NewMockClient(ctrl)
	b := kycOnlyBackends{kycClient: kycClient}

	walletID := "wallet-id"
	kycClient.EXPECT().GetKYCStatus(gomock.Any(), walletID).Return(kyc.StatusDocumentsRequired, nil)

	_, err := GetWalletAddress(context.Background(), b, walletID)
	require.ErrorIs(t, err, kyc.ErrKYCResubmissionRequired)
}
