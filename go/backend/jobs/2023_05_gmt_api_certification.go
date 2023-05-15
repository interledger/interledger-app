package jobs

import (
	"context"
	"strings"
	"time"

	"go.temporal.io/sdk/activity"

	"gitlab.com/fynbos/env"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers"
	gmt_ops "gitlab.com/fynbos/backend/providers/gmt/ops"
	"go.temporal.io/sdk/workflow"
)

func (a *Activity) UpdateWalletAddress(ctx context.Context, walletID, state, zip string) error {
	if state == "" || zip == "" {
		return nil
	}

	logger := activity.GetLogger(ctx)

	id, err := a.b.KYC().GetIndividualDetails(ctx, walletID)
	if err != nil {
		return err
	}

	id.Address.ZipCode = zip
	id.Address.State = state

	logger.Info("Updating User state", "state", id.Address.State, "zip", id.Address.ZipCode)

	_, err = a.b.KYC().UpdateIndividualDetails(ctx, *id)
	if err != nil {
		return err
	}

	return nil
}

// RunGMTCertification does a series of create transactions calls on GMT API in the dev-eu1 environment.
// Each wallet and linked account was set up manually and will use the exact same flow as any normal user operation
// The exception being the users will not have sender and receiver IDs defined in the database as we do not go
// through the entire flow just the create transaction part
func RunGMTCertification(ctx workflow.Context) error {
	var a *Activity
	var gmtActivity *gmt_ops.Activity

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)

	if !env.IsDev() {
		logger.Error("not going to run GMT certification in environment", "env", env.GetEnv())
		return nil
	}

	cases := []struct {
		name     string
		zip      string
		state    string
		args     providers.TransfersArgs
		expected string
	}{
		{
			name:  "case 1.1",
			zip:   "86033",
			state: "US-AZ",
			args: providers.TransfersArgs{
				FromLinkedAccountID: "dc02ec43-eeb0-455a-95df-32335c415375",
				ToLinkedAccountID:   "d3df3382-3bd1-412c-96ad-7c243b4b6f72",
				FromWalletID:        "84a92788-7d07-4a69-9f25-687de4bce8ef", //barnard+gmt+autotest+sender+1@fynbos.dev
				ToWalletID:          "0b523224-8345-4314-b2bb-955f8b82d225", //barnard+gmt+autotest+receiver+1@fynbos.dev
				Amount:              currency.FromFloat64(9000, currency.USD),
				FromTransactionID:   uuid.NewString(),
			},
			expected: "rejected",
		},
		{
			name:  "case 1.2",
			zip:   "93001",
			state: "US-CA",
			args: providers.TransfersArgs{
				FromLinkedAccountID: "dc02ec43-eeb0-455a-95df-32335c415375",
				ToLinkedAccountID:   "d3df3382-3bd1-412c-96ad-7c243b4b6f72",
				FromWalletID:        "84a92788-7d07-4a69-9f25-687de4bce8ef", //barnard+gmt+autotest+sender+1@fynbos.dev
				ToWalletID:          "0b523224-8345-4314-b2bb-955f8b82d225", //barnard+gmt+autotest+receiver+1@fynbos.dev
				Amount:              currency.FromFloat64(9000, currency.USD),
				FromTransactionID:   uuid.NewString(),
			},
			expected: "created",
		},
		{
			name:  "case 1.3",
			zip:   "93001",
			state: "CA",
			args: providers.TransfersArgs{
				FromLinkedAccountID: "dc02ec43-eeb0-455a-95df-32335c415375",
				ToLinkedAccountID:   "93856ba9-eaf5-4df2-a30f-1a4b7bbf74fa",
				FromWalletID:        "84a92788-7d07-4a69-9f25-687de4bce8ef", //barnard+gmt+autotest+sender+1@fynbos.dev
				ToWalletID:          "b532a5e4-e572-4d7f-a9e8-565065f4f2bc", //barnard+gmt+autotest+receiver+2@fynbos.dev
				Amount:              currency.FromFloat64(2100, currency.USD),
				FromTransactionID:   uuid.NewString(),
			},
			expected: "rejected",
		},
		{
			name:  "case 1.4",
			zip:   "93001",
			state: "US-CA",
			args: providers.TransfersArgs{
				FromLinkedAccountID: "dc02ec43-eeb0-455a-95df-32335c415375",
				ToLinkedAccountID:   "4199deeb-b4d0-4007-ba11-bd4c9b8906bf",
				FromWalletID:        "84a92788-7d07-4a69-9f25-687de4bce8ef", //barnard+gmt+autotest+sender+1@fynbos.dev
				ToWalletID:          "378b1743-89be-4422-b679-0345ebcd3e83", //barnard+gmt+autotest+receiver+3@fynbos.dev
				Amount:              currency.FromFloat64(2100, currency.USD),
				FromTransactionID:   uuid.NewString(),
				ForceEDD:            true,
			},
			expected: "created",
		},
		{
			name: "case 1.5",
			args: providers.TransfersArgs{
				FromLinkedAccountID: "dc02ec43-eeb0-455a-95df-32335c415375",
				ToLinkedAccountID:   "93856ba9-eaf5-4df2-a30f-1a4b7bbf74fa",
				FromWalletID:        "84a92788-7d07-4a69-9f25-687de4bce8ef", //barnard+gmt+autotest+sender+1@fynbos.dev
				ToWalletID:          "b532a5e4-e572-4d7f-a9e8-565065f4f2bc", //barnard+gmt+autotest+receiver+2@fynbos.dev
				Amount:              currency.FromFloat64(8000, currency.USD),
				FromTransactionID:   uuid.NewString(),
				ForceEDD:            true,
			},
			expected: "created",
		},
		{
			name:  "case 1.6",
			zip:   "92101",
			state: "US-CA",
			args: providers.TransfersArgs{
				FromLinkedAccountID: "6b5ca5f0-7148-4e6d-a6f3-83d7c139fada",
				ToLinkedAccountID:   "531ffd56-2b60-4085-bbae-f557bb742a7e",
				FromWalletID:        "2396e098-25cf-4eb1-8a8f-f79843520cbb", //barnard+gmt+autotest+newremnewman+sender@fynbos.dev
				ToWalletID:          "d94e01f7-3152-4379-bb8a-fc3b26c5854c", //barnard+gmt+autotest+juliosolano+recv@fynbos.dev
				Amount:              currency.FromFloat64(10, currency.USD),
				FromTransactionID:   uuid.NewString(),
			},
			expected: "hold",
		},
	}

	for _, tc := range cases {
		err := workflow.ExecuteActivity(ctx, a.UpdateWalletAddress, tc.args.FromWalletID, tc.state, tc.zip).Get(ctx, nil)
		if err != nil {
			logger.Error("failed to update wallet address", "err", err)
			continue
		}

		var tr gmt_ops.TransactionResp
		err = workflow.ExecuteActivity(ctx, gmtActivity.InsertACH, gmt_ops.InsertACHArgs{
			TransfersArgs:         tc.args,
			OriginalPaymentMethod: "ACH",
		}).Get(ctx, &tr)

		if err != nil {
			logger.Error("failed to add transaction", "err", err)
			continue
		}

		logger.Info("Test Case Result",
			"test_name", tc.name,
			"result", tr.Status,
			"expected", tc.expected,
			"matches", strings.EqualFold(tc.expected, tr.Status),
			"references", tr.ID)
	}

	return nil
}
