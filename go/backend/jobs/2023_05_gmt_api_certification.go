package jobs

import (
	"context"
	"time"

	"gitlab.com/fynbos/backend/providers"
	gmt_ops "gitlab.com/fynbos/backend/providers/gmt/ops"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/env"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"
)

func (a *Activity) UpdateWalletAddress(ctx context.Context, walletID, state, zip string) error {
	if state == "" || zip == "" {
		return nil
	}

	id, err := a.b.KYC().GetIndividualDetails(ctx, walletID)
	if err != nil {
		return err
	}

	id.Address.ZipCode = zip
	id.Address.State = state

	_, err = a.b.KYC().UpdateIndividualDetails(ctx, *id)
	if err != nil {
		return err
	}

	return nil
}

func RunGMTCertificationStep2(ctx workflow.Context) error {
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

	// Cancel test case 1.2 transaction
	/*var respStatus string
	err := workflow.ExecuteActivity(ctx, gmtActivity.RequestCancellation, "000581552673", "Automated test cancellation").Get(ctx, &respStatus)
	if err != nil {
		logger.Error("failed to request cancellation of 1.2 transaction")
		return err
	}

	logger.Info("test case 1.2 cancellation result", "result", respStatus)

	// Cancel test case 1.4 transaction
	err = workflow.ExecuteActivity(ctx, gmtActivity.RequestCancellation, "000197833254", "Automated test cancellation").Get(ctx, &respStatus)
	if err != nil {
		logger.Error("failed to request cancellation of 1.4 transaction")
		return err
	}

	logger.Info("test case 1.4  cancellation result", "result", respStatus)

	// Cancel test case 1.5 transaction
	err := workflow.ExecuteActivity(ctx, gmtActivity.RequestCancellation, "000759265219", "Automated test cancellation").Get(ctx, &respStatus)
	if err != nil {
		logger.Error("failed to request cancellation of 1.5 transaction")
		return err
	}

	logger.Info("test case 1.5  cancellation result", "result", respStatus)
	*/
	var notifications map[string]string
	err := workflow.ExecuteActivity(ctx, gmtActivity.GetNotifications).Get(ctx, &notifications)
	if err != nil {
		logger.Error("failed to request cancellation of 1.4 transaction")
		return err
	}

	for k, v := range notifications {
		logger.Info("Notification Status", "id", k, "status", v)
	}

	return nil
}

type GmtTestArgs struct {
	Name      string
	Zip       string
	State     string
	Args      providers.TransfersArgs
	Expected  string
	ExpectErr bool
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
	childWorkflowOptions := workflow.ChildWorkflowOptions{
		ParentClosePolicy: enums.PARENT_CLOSE_POLICY_REQUEST_CANCEL,
	}
	ctx = workflow.WithChildOptions(ctx, childWorkflowOptions)

	logger := workflow.GetLogger(ctx)

	if !env.IsDev() {
		logger.Error("not going to run GMT certification in environment", "env", env.GetEnv())
		return nil
	}

	cases := []GmtTestArgs{
		/*	{
				Name:  "case 1.1",
				Zip:   "86033",
				State: "US-AZ",
				Args: providers.TransfersArgs{
					FromPaymentPointer:  "https://eu1.fynbos.me/goldensender1",
					ToPaymentPointer:    "https://eu1.fynbos.me/goldenreceiver1",
					FromLinkedAccountID: "929a4912-8c4c-45d4-8e8f-e8b6bc2ae050",
					ToLinkedAccountID:   "93856ba9-eaf5-4df2-a30f-1a4b7bbf74fa",
					FromWalletID:        "8ee9a91e-5774-4410-b174-334cd1833e29", //barnard+gmttestsender1@fynbos.dev
					ToWalletID:          "b532a5e4-e572-4d7f-a9e8-565065f4f2bc", //barnard+gmt+autotest+juliosolano+recv@fynbos.dev
					Amount:              currency.FromFloat64(900, currency.USD),
					FromTransactionID:   uuid.NewString(),
					ForceNoEDD:          true,
				},
				Expected:  "rejected",
				ExpectErr: true,
				SetupAcc:  false,
			},
			{
				Name:  "case 1.2",
				Zip:   "93001",
				State: "US-CA",
				Args: providers.TransfersArgs{
					FromPaymentPointer:  "https://eu1.fynbos.me/goldensender2",
					ToPaymentPointer:    "https://eu1.fynbos.me/goldenreceiver2",
					FromWalletID:        "5b1f69b9-8e85-4a99-9932-2b522b07f6c8",
					FromLinkedAccountID: "73b06d41-0462-410e-84eb-21a0c157c512",
					ToWalletID:          "5096e217-756f-4758-8808-4788010831c7",
					ToLinkedAccountID:   "9224fd47-deb0-4f74-9234-6f5fdaf01528",
					Amount:              currency.FromFloat64(900, currency.USD),
					FromTransactionID:   uuid.NewString(),
					ForceNoEDD:          true,
					ForceEDD:            false,
				},
				Expected: "created",
			},*/
		/*	{
				Name:  "case 1.4",
				Zip:   "93001",
				State: "US-CA",
				Args: providers.TransfersArgs{
					FromPaymentPointer:  "https://eu1.fynbos.me/goldensender2",
					ToPaymentPointer:    "https://eu1.fynbos.me/goldenreceiver3",
					FromWalletID:        "5b1f69b9-8e85-4a99-9932-2b522b07f6c8",
					FromLinkedAccountID: "73b06d41-0462-410e-84eb-21a0c157c512",
					ToWalletID:          "84a92788-7d07-4a69-9f25-687de4bce8ef",
					ToLinkedAccountID:   "dc02ec43-eeb0-455a-95df-32335c415375",
					Amount:              currency.FromFloat64(2100, currency.USD),
					FromTransactionID:   uuid.NewString(),
					ForceNoEDD:          false,
					ForceEDD:            true,
				},
				Expected:  "rejected",
				ExpectErr: true,
			},
			{
				Name:  "case 1.5",
				Zip:   "93001",
				State: "US-CA",
				Args: providers.TransfersArgs{
					FromPaymentPointer:  "https://eu1.fynbos.me/goldensender3",
					ToPaymentPointer:    "https://eu1.fynbos.me/goldenreceiver5",
					FromWalletID:        "84a92788-7d07-4a69-9f25-687de4bce8ef",
					FromLinkedAccountID: "dc02ec43-eeb0-455a-95df-32335c415375",
					ToWalletID:          "5b1f69b9-8e85-4a99-9932-2b522b07f6c8",
					ToLinkedAccountID:   "73b06d41-0462-410e-84eb-21a0c157c512",
					Amount:              currency.FromFloat64(8000, currency.USD),
					FromTransactionID:   uuid.NewString(),
					ForceEDD:            true,
					ForceNoEDD:          false,
				},
				Expected: "created",
			},*/
		/*				{
								Name:  "case 1.4",
								Zip:   "93001",
								State: "US-CA",
								Args: providers.TransfersArgs{
									FromPaymentPointer: "https://eu1.fynbos.me/goldensender2",
									ToPaymentPointer:   "https://eu1.fynbos.me/goldenrethreever4",
									Amount:             currency.FromFloat64(2100, currency.USD),
									FromTransactionID:  uuid.NewString(),
									ForceEDD:           true,
								},
								Expected: "created",
								SetupAcc: true,
							},
							{
								Name:  "case 1.5",
								Zip:   "93001",
								State: "US-CA",
								Args: providers.TransfersArgs{
									FromPaymentPointer: "https://eu1.fynbos.me/goldensender3",
									ToPaymentPointer:   "https://eu1.fynbos.me/goldenreceiver5",
									Amount:             currency.FromFloat64(8000, currency.USD),
									FromTransactionID:  uuid.NewString(),
									ForceEDD:           true,
								},
								Expected: "created",
								SetupAcc: true,
							},
						{
							Name:  "case 1.6",
							Zip:   "92101",
							State: "US-CA",
							Args: providers.TransfersArgs{
								FromPaymentPointer:  "https://eu1.fynbos.me/newrem",
								ToPaymentPointer:    "https://eu1.fynbos.me/julio",
								FromLinkedAccountID: "6b5ca5f0-7148-4e6d-a6f3-83d7c139fada",
								ToLinkedAccountID:   "531ffd56-2b60-4085-bbae-f557bb742a7e",
								FromWalletID:        "2396e098-25cf-4eb1-8a8f-f79843520cbb", //barnard+gmt+autotest+newremnewman+sender@fynbos.dev
								ToWalletID:          "d94e01f7-3152-4379-bb8a-fc3b26c5854c", //barnard+gmt+autotest+juliosolano+recv@fynbos.dev
								Amount:              currency.FromFloat64(10, currency.USD),
								FromTransactionID:   uuid.NewString(),
							},
							Expected: "hold",
							SetupAcc: false,
						},*/
	}

	for _, tc := range cases {

		err := workflow.ExecuteActivity(ctx, a.UpdateWalletAddress, tc.Args.FromWalletID, tc.State, tc.Zip).Get(ctx, nil)
		if err != nil {
			logger.Error("failed to setup sender", "err", err, "testcase", tc.Name)
			continue
		}

		err = workflow.ExecuteActivity(ctx, gmtActivity.AddTransaction, transactions.CreateTransactionArgs{
			ID:          tc.Args.FromTransactionID,
			WalletID:    tc.Args.FromWalletID,
			ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
			Provider:    "gmt",
			State:       transactions.StatePending,
			Note:        "Used for compliance testing",
			Source:      tc.Args.FromPaymentPointer,
			Destination: tc.Args.ToPaymentPointer,
			Amount:      tc.Args.Amount,
		}).Get(ctx, nil)
		if err != nil {
			logger.Warn("failed to setup transaction preflight", "err", err, "testcase", tc.Name, "reference", tc.Args.FromTransactionID)
			continue
		}

		logger.Info("Starting test case workflow", "testcase", tc.Name, "")

		var resp providers.TransferResponse
		err = workflow.ExecuteChildWorkflow(ctx, gmt_ops.ACH2ACHTransferWorkflow, tc.Args).Get(ctx, &resp)

		if tc.ExpectErr && err != nil {
			logger.Info("Expected Test Case Error",
				"test_name", tc.Name,
				"err", err)
			continue
		}
		if err != nil {
			logger.Warn("failed to add transaction", "err", err, "testcase", tc.Name)
			continue
		}

		logger.Info("Test Case Result",
			"test_name", tc.Name,
			"result", resp.OutgoingTransferState,
			"expected", tc.Expected,
			"references", tc.Args.FromTransactionID)
	}

	return nil
}
