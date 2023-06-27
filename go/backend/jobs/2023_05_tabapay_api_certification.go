package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/wallets"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/providers/basistheory"
	gmt_workflow "gitlab.com/fynbos/backend/providers/gmt/ops"
	httplog "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/backend/providers/tabapay"
	tabapay_external "gitlab.com/fynbos/backend/providers/tabapay/external"
	tabapay_workflow "gitlab.com/fynbos/backend/providers/tabapay/workflows"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type TabapayTestCase struct {
	Case            string
	CardNumber      string
	CVV             string
	AVS             bool
	Address         *kyc.Address
	WalletID        string
	Amount          currency.Amount
	LinkedAccountID string
	Pull            bool
}

func (a *Activity) SeedTabapayWallets(ctx context.Context, tcs []TabapayTestCase) error {
	for _, tc := range tcs {
		w, err := a.b.Wallets().Get(ctx, tc.WalletID)
		if err != nil && !errors.Is(err, wallets.ErrNoWalletFound) {
			return err
		}
		if w != nil {
			continue
		}

		w, err = a.b.Wallets().Create(ctx, wallets.CreateArgs{
			ID:     tc.WalletID,
			UserID: uuid.NewString(),
		})
		if err != nil {
			return err
		}

		_, err = a.b.KYC().UpdateIndividualDetails(ctx, kyc.IndividualDetails{
			WalletID:  w.ID,
			FirstName: "Tabapay",
			LastName:  "Test",
			IPAddress: "127.0.0.1",
			Address:   tc.Address,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Activity) SeedTabapayLinkedAccounts(ctx context.Context, tcs []TabapayTestCase) error {
	return nil
}

func (a *Activity) CreateBasisTheoryCardToken(ctx context.Context, args basistheory.CreateCardTokenArgs) (string, error) {
	tokenID, err := a.b.BasisTheory().CreateCardToken(ctx, args)
	if err != nil {
		return "", err
	}

	return tokenID, nil
}

func TabapayCertificationWorkflow(ctx workflow.Context) error {
	var a *Activity
	var tabapayActivity *tabapay_workflow.Activity
	var gmtActivity *gmt_workflow.Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	workflowID := workflow.GetInfo(ctx).WorkflowExecution.ID

	cardQueryTestCases := []TabapayTestCase{
		{
			Case:       fmt.Sprintf("%s_CQ01", workflowID),
			CardNumber: "9400111999999990",
			WalletID:   "f56deeaf-011d-48cb-9431-6fda0058b364",
		},
		{
			Case:       fmt.Sprintf("%s_CQ02", workflowID),
			CardNumber: "9500100999999992",
			WalletID:   "f56deeaf-011d-48cb-9431-6fda0058b364",
		},
		{
			Case:       fmt.Sprintf("%s_CQ03", workflowID),
			CardNumber: "9401111999999999",
			WalletID:   "989b5375-d22d-4c4a-91f0-9bf4f82f9ce0",
			AVS:        true,
			Address: &kyc.Address{
				Line1:   "999",
				ZipCode: "99992",
			},
		},
		{
			Case:       fmt.Sprintf("%s_CQ04", workflowID),
			CardNumber: "9401111999999999",
			WalletID:   "fcf1f8c8-d59b-45b1-a177-7ed9c08d3b4a",
			AVS:        true,
			Address: &kyc.Address{
				ZipCode: "99990",
			},
		},
		{
			Case:       fmt.Sprintf("%s_CQ05", workflowID),
			CardNumber: "9401111999999999",
			WalletID:   "a68b1971-3889-4283-85fd-28d9a8380279",
			AVS:        true,
			Address: &kyc.Address{
				ZipCode: "99991",
			},
		},
		{
			Case:       fmt.Sprintf("%s_CQ06", workflowID),
			CardNumber: "9401111999999999",
			WalletID:   "3628433b-4afe-41ff-9356-44306d708d66",
			AVS:        true,
			Address: &kyc.Address{
				Line1:   "257 Market Street",
				ZipCode: "99991",
			},
		},
		{
			Case:       fmt.Sprintf("%s_CQ07", workflowID),
			CardNumber: "9401111999999999",
			WalletID:   "3597b0c0-4193-4948-8f26-8a8284f01055",
			AVS:        true,
			Address: &kyc.Address{
				Line1:   "257 Market Street",
				ZipCode: "99992",
			},
		},
		{
			Case:       fmt.Sprintf("%s_CQ08", workflowID),
			CardNumber: "9401111999999999",
			WalletID:   "140110bd-c94c-4dcf-b735-9a56069b40c0",
			AVS:        true,
			Address: &kyc.Address{
				Line1:   "257 Market Street",
				ZipCode: "02135",
			},
		},
		{
			Case:       fmt.Sprintf("%s_CQ09", workflowID),
			CardNumber: "9401111999999999",
			WalletID:   "44924b23-4d27-4a4c-93ef-89690f934912",
			AVS:        true,
			Address: &kyc.Address{
				ZipCode: "02135",
			},
		},
		{
			Case:       fmt.Sprintf("%s_CQ10", workflowID),
			CardNumber: "9401111999999999",
			WalletID:   "733a2d23-f3e7-44a4-b99d-5c1ec550af84",
			CVV:        "123",
			AVS:        true,
			Address: &kyc.Address{
				ZipCode: "02135",
			},
		},
		{
			Case:       fmt.Sprintf("%s_CQ11", workflowID),
			CardNumber: "9401111999999999",
			WalletID:   "733a2d23-f3e7-44a4-b99d-5c1ec550af84",
			CVV:        "999",
			AVS:        true,
			Address: &kyc.Address{
				ZipCode: "02135",
			},
		},
	}
	err := workflow.ExecuteActivity(ctx, a.SeedTabapayWallets, cardQueryTestCases).Get(ctx, nil)
	if err != nil {
		return err
	}
	for _, tc := range cardQueryTestCases {
		var basisTheoryTokenID string
		err = workflow.ExecuteActivity(ctx, a.CreateBasisTheoryCardToken, basistheory.CreateCardTokenArgs{
			WalletID:        tc.WalletID,
			Number:          tc.CardNumber,
			ExpirationMonth: 4,
			ExpirationYear:  2025,
			CVC:             "123",
		}).Get(ctx, &basisTheoryTokenID)
		if err != nil {
			return err
		}

		var response tabapay_external.CreateAccountResponse
		newCtx := workflow.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Context: tc.Case,
		})
		err = workflow.ExecuteActivity(
			newCtx,
			tabapayActivity.QueryCard,
			tabapay_workflow.QueryCard{
				WalletID:       tc.WalletID,
				CardNumber:     fmt.Sprintf("{{ %s | json: '$.number' }}", basisTheoryTokenID),
				ExpirationDate: fmt.Sprintf("{{ %s | json: '$.expiration_year' | to_string }}{{ %s | json: '$.expiration_month' | pad_left: 2,'0' }}", basisTheoryTokenID, basisTheoryTokenID),
				AVS:            tc.AVS,
				CVV:            tc.CVV,
			},
		).Get(newCtx, &response)
		if err != nil {
			logger.Error("Error with case.", tc.Case, err)
		}
	}

	createAccountTestCases := []TabapayTestCase{
		{
			Case:       fmt.Sprintf("%s_AC01", workflowID),
			CardNumber: "9400111999999990",
			WalletID:   "48e4a782-bc13-438b-9252-8f14a2496322",
			Address: &kyc.Address{
				City:        "Suffolk County",
				State:       "US-MA",
				CountryCode: "US",
				ZipCode:     "02135",
				Line1:       "257 Market Street",
				Line2:       "Boston",
			},
		},
		{
			Case:       fmt.Sprintf("%s_AC03", workflowID),
			CardNumber: "9400111999999990",
			WalletID:   "b989667c-362f-4460-b75a-b8ec0180b1af",
			Address: &kyc.Address{
				Line1:       "1001 Fairchild St, , Suite 300",
				State:       "US-MA",
				CountryCode: "US",
				ZipCode:     "02135",
			},
		},
		{
			Case:       fmt.Sprintf("%s_AC04", workflowID),
			CardNumber: "9400111999999990",
			WalletID:   "a853672a-68d5-4824-af82-515b22502627",
			Address: &kyc.Address{
				Line1:       "123 Select St",
				City:        "Select",
				State:       "US-MA",
				CountryCode: "US",
				ZipCode:     "02135",
			},
		},
	}
	err = workflow.ExecuteActivity(ctx, a.SeedTabapayWallets, createAccountTestCases).Get(ctx, nil)
	if err != nil {
		return err
	}
	for _, tc := range createAccountTestCases {
		var basisTheoryTokenID string
		err = workflow.ExecuteActivity(ctx, a.CreateBasisTheoryCardToken, basistheory.CreateCardTokenArgs{
			WalletID:        tc.WalletID,
			Number:          tc.CardNumber,
			ExpirationMonth: 4,
			ExpirationYear:  2025,
			CVC:             "123",
		}).Get(ctx, &basisTheoryTokenID)
		if err != nil {
			return err
		}

		var response tabapay_external.CreateAccountResponse
		newCtx := workflow.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Context: tc.Case,
		})
		err = workflow.ExecuteActivity(
			newCtx,
			tabapayActivity.CreateExternalCard,
			tabapay_workflow.CreateExternalCardArgs{
				RejectDuplicateCard: true,
				WalletID:            tc.WalletID,
				CardNumber:          fmt.Sprintf("{{ %s | json: '$.number' }}", basisTheoryTokenID),
				ExpirationDate:      fmt.Sprintf("{{ %s | json: '$.expiration_year' | to_string }}{{ %s | json: '$.expiration_month' | pad_left: 2,'0' }}", basisTheoryTokenID, basisTheoryTokenID),
			},
		).Get(newCtx, &response)
		if err != nil {
			logger.Error("Error with case.", tc.Case, err)
		}
	}

	createTransactionTestCases := []TabapayTestCase{
		{
			Case: fmt.Sprintf("%s_TC01", workflowID),
			Amount: currency.Amount{
				Value:    1,
				Currency: currency.ParseCurrency("USD"),
			},
			Pull:            true,
			LinkedAccountID: "50a05498-c3c3-4646-a967-8435c3e5e20a",
		},
		{
			Case: fmt.Sprintf("%s_TC02", workflowID),
			Amount: currency.Amount{
				Value:    4,
				Currency: currency.ParseCurrency("USD"),
			},
			Pull:            true,
			LinkedAccountID: "50a05498-c3c3-4646-a967-8435c3e5e20a",
		},
		{
			Case: fmt.Sprintf("%s_TC03", workflowID),
			Amount: currency.Amount{
				Value:    4,
				Currency: currency.ParseCurrency("USD"),
			},
			Pull:            true,
			LinkedAccountID: "50a05498-c3c3-4646-a967-8435c3e5e20a",
		},
		{
			Case: fmt.Sprintf("%s_TC04", workflowID),
			Amount: currency.Amount{
				Value:    1,
				Currency: currency.ParseCurrency("USD"),
			},
			Pull:            false,
			LinkedAccountID: "50a05498-c3c3-4646-a967-8435c3e5e20a",
		},
		{
			Case: fmt.Sprintf("%s_TC05", workflowID),
			Amount: currency.Amount{
				Value:    4,
				Currency: currency.ParseCurrency("USD"),
			},
			Pull:            false,
			LinkedAccountID: "50a05498-c3c3-4646-a967-8435c3e5e20a",
		},
		{
			Case: fmt.Sprintf("%s_TC06", workflowID),
			Amount: currency.Amount{
				Value:    4,
				Currency: currency.ParseCurrency("USD"),
			},
			Pull:            false,
			LinkedAccountID: "50a05498-c3c3-4646-a967-8435c3e5e20a",
		},
		{
			Case: fmt.Sprintf("%s_TC07", workflowID),
			Amount: currency.Amount{
				Value:    25,
				Currency: currency.ParseCurrency("USD"),
			},
			Pull:            true,
			LinkedAccountID: "50a05498-c3c3-4646-a967-8435c3e5e20a",
		},
		{
			Case: fmt.Sprintf("%s_TC08", workflowID),
			Amount: currency.Amount{
				Value:    25,
				Currency: currency.ParseCurrency("USD"),
			},
			Pull:            true,
			LinkedAccountID: "50a05498-c3c3-4646-a967-8435c3e5e20a",
		},
		{
			Case: fmt.Sprintf("%s_TC09", workflowID),
			Amount: currency.Amount{
				Value:    8,
				Currency: currency.ParseCurrency("USD"),
			},
			Pull:            true,
			LinkedAccountID: "50a05498-c3c3-4646-a967-8435c3e5e20a",
		},
		{
			Case: fmt.Sprintf("%s_TC10", workflowID),
			Amount: currency.Amount{
				Value:    7,
				Currency: currency.ParseCurrency("USD"),
			},
			Pull:            true,
			LinkedAccountID: "50a05498-c3c3-4646-a967-8435c3e5e20a",
		},
		{
			Case: fmt.Sprintf("%s_TC11", workflowID),
			Amount: currency.Amount{
				Value:    210,
				Currency: currency.ParseCurrency("USD"),
			},
			Pull:            true,
			LinkedAccountID: "50a05498-c3c3-4646-a967-8435c3e5e20a",
		},
		{
			Case: fmt.Sprintf("%s_TC12", workflowID),
			Amount: currency.Amount{
				Value:    201,
				Currency: currency.ParseCurrency("USD"),
			},
			Pull:            true,
			LinkedAccountID: "50a05498-c3c3-4646-a967-8435c3e5e20a",
		},
	}
	for _, tc := range createTransactionTestCases {
		newCtx := workflow.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Context: tc.Case,
		})

		var tabapayTransaction tabapay.Transaction
		if tc.Pull {
			err = workflow.ExecuteActivity(newCtx, gmtActivity.PullFromCard, gmt_workflow.PullFromCardArgs{
				ReferenceID:         tabapay.NewReferenceID(),
				CardLinkedAccountID: tc.LinkedAccountID,
				Amount:              tc.Amount,
			}).Get(newCtx, &tabapayTransaction)
		} else {
			err = workflow.ExecuteActivity(newCtx, gmtActivity.PushToCard, gmt_workflow.PullFromCardArgs{
				ReferenceID:         tabapay.NewReferenceID(),
				CardLinkedAccountID: tc.LinkedAccountID,
				Amount:              tc.Amount,
			}).Get(newCtx, &tabapayTransaction)
		}
		if err != nil {
			logger.Error("Error with case.", tc.Case, err)
		}
	}

	return nil
}
