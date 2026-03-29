package v1

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/fynbos/sdks/fiant/v1/domain/dto"
)

type transactionsService struct {
	client *Client
}

type SandboxActionTypeEnum string

const (
	SETTLE_ACH SandboxActionTypeEnum = "SETTLE_ACH"
	RETURN_ACH SandboxActionTypeEnum = "RETURN_ACH"
)

// only available in sandbox environment, used to simulate settlement and returns for ACH transactions
// https://developers.platform.fiant.io/reference/performaction
func (ts *transactionsService) SandboxAction(ctx context.Context, requestID string, action SandboxActionTypeEnum) error {
	path := fmt.Sprintf("transactions/%v/actions", requestID)

	payload := []byte(`{"action":"` + string(action) + `"}`)
	resp, err := ts.client.post(ctx, path, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to perform sandbox action, code: %s", resp.Status)
	}

	return nil
}

// https://developers.platform.fiant.io/reference/gettransaction
// note: the requestID is the identifier
func (ts *transactionsService) Get(ctx context.Context, requestID string) (dto.Transaction, error) {
	path := fmt.Sprintf("transactions/%v", requestID)
	resp, err := ts.client.get(ctx, path)
	if err != nil {
		return dto.Transaction{}, err
	}
	return consumeResponse[dto.Transaction](resp, http.StatusOK)
}

// https://developers.platform.fiant.io/reference/deposit
// TODO: WIP
func (ts *transactionsService) DepositToWalletACH(ctx context.Context, user dto.User, requestId string, scenario string) (dto.ObjectReference, error) {
	path := "transactions/deposits"

	deposit := struct {
		Initiator dto.User `json:"initiator"`

		SourceMethod      dto.PaymentMethod `json:"sourceMethod"`
		DestinationMethod dto.PaymentMethod `json:"destinationMethod"`

		Amount   float64 `json:"amount"`
		USDValue float64 `json:"usdValue"`
		Type     string  `json:"type"`
		Date     string  `json:"date"`
	}{
		Initiator: user,

		SourceMethod: *dto.NewPaymentMethod(
			dto.ACH_METHOD,
			dto.WithACHPaymentMethod(*dto.NewACHPaymentMethod(
				dto.WithACHCurrency("USD"),
				dto.WithACHBillingEmail("boz@fiant.io"),
				dto.WithACHPaymentInformation(dto.PaymentInformation{
					Type: dto.BANK_ACCOUNT_PAYMENT,
					BankAccount: &dto.BankAccount{
						BankRoutingNumber: "026009593",
						BankAccountType:   "CHECKING",
						BankAccountNumner: "74600015199010",
						AccountBankName:   "Test Bank",
						AccountHolderName: "Boz",
					},
				}),
			)),
		),

		DestinationMethod: *dto.NewPaymentMethod(
			dto.WALLET_METHOD,
			dto.WithWalletPaymentMethod(*dto.NewWalletPaymentMethod(
				dto.WithWalletPaymentInformation(dto.Wallet{
					ID:   "c95698c2-57a8-44f0-adc5-907b53b93efe",
					Type: "WALLET",
				}),
			)),
		),

		Amount:   1,
		USDValue: 1,
		Type:     "DEPOSIT",
		Date:     time.Now().Format(time.RFC3339),
	}

	resp, err := ts.client.post(ctx, path, deposit, withHeader(ptiRequestIDHeader, requestId), withHeader(ptiScenarioIDHeader, scenario))
	if err != nil {
		return dto.ObjectReference{}, err
	}
	return consumeResponse[dto.ObjectReference](resp, http.StatusCreated)
}
