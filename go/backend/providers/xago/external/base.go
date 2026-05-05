package external

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/providers/xago/external/domain/dto"
	"gitlab.com/fynbos/backend/user"
)

type baseService struct {
	client *client
}

// checked
func (bs *baseService) BankAccounts(ctx context.Context) ([]dto.Currency, error) {
	resp, err := bs.client.get(ctx, bs.client.baseURL, "currencies")
	if err != nil {
		return nil, err
	}
	return consumeResponse[[]dto.Currency](resp, http.StatusOK)
}

// this is a beneficiary
// todo, needs to be tested with persona
func (bs *baseService) CreateSubAccount(ctx context.Context, u user.User, details kyc.IndividualDetails, idNumber, personaInquiryURL string) (dto.SubAccount, error) {
	resp, err := bs.client.post(ctx, bs.client.baseURL, "company/accounts", dto.SubAccountRequest{
		FirstName:       details.FirstName,
		LastName:        details.LastName,
		Email:           u.Email,
		MobileNumber:    u.PhoneNumber,
		IdentityType:    IdentityTypeIndividual,
		PersonaURL:      personaInquiryURL,
		IDNumber:        idNumber,
		PhysicalAddress: details.Address.String(),
	})
	if err != nil {
		return dto.SubAccount{}, err
	}
	return consumeResponse[dto.SubAccount](resp, http.StatusOK)
}

// checked
func (bs *baseService) ListDeposits(ctx context.Context, page int) ([]dto.Deposit, error) {
	resp, err := bs.client.get(ctx, bs.client.baseURL, "company/transactions",
		withQueryParam("limit", "10"),
		withQueryParam("page", fmt.Sprintf("%d", page)),
	)
	if err != nil {
		return nil, err
	}
	result, err := consumeResponse[dto.ListDepositsResponse](resp, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return result.Deposits, nil
}

func (bs *baseService) GetDeposit(ctx context.Context, id string) (dto.Deposit, error) {
	resp, err := bs.client.get(ctx, bs.client.baseURL, fmt.Sprintf("company/transactions/%s", id))
	if err != nil {
		return dto.Deposit{}, err
	}
	return consumeResponse[dto.Deposit](resp, http.StatusOK)
}

func (bs *baseService) GetWithdrawal(ctx context.Context, id string) (dto.Withdrawal, error) {
	resp, err := bs.client.get(ctx, bs.client.baseURL, "transactions",
		withQueryParam("transactionId", id),
	)
	if err != nil {
		return dto.Withdrawal{}, err
	}
	return consumeResponse[dto.Withdrawal](resp, http.StatusOK)
}

// Only used in local/sandbox environments to simulate a deposit.
func (bs *baseService) TestDeposit(ctx context.Context, reqStruct dto.TestDepositRequest) error {
	resp, err := bs.client.post(ctx, bs.client.baseURL, "company/accounts/testdeposit", reqStruct)
	if err != nil {
		return err
	}
	_, err = consumeResponse[struct{}](resp, http.StatusOK)
	return err
}

func (bs *baseService) CreateTransaction(ctx context.Context, amt currency.Amount, idempotencyKey, beneficiaryID, reference string) (string, error) {
	if reference == "" {
		reference = "Interledger Wallet"
	}

	resp, err := bs.client.post(ctx, bs.client.baseURL, "transfers", dto.CreateTransferRequest{
		Amount:        amt.Float64(),
		CurrencyCode:  amt.Currency.String(),
		BeneficiaryID: beneficiaryID,
		Reference:     reference,
	})
	if err != nil {
		return "", err
	}

	// todo remove
	if resp.StatusCode == http.StatusUnprocessableEntity {
		resp.Body.Close()
		return "", fmt.Errorf("failed to add xargo transaction, transaction already exists")
	}
	return consumeResponse[string](resp, http.StatusOK)
}

func (bs *baseService) EstimateCurrencyConvert(ctx context.Context, pair dto.ConvertCurrencyPairEnum, amount float64) (dto.ConvertCurrencyResponse, error) {
	resp, err := bs.client.post(ctx, bs.client.baseURL, "currencyconvert", dto.ConvertCurrencyRequest{
		ConvertCurrencyPair: pair,
		Amount:              amount,
		EstimateCalculation: true,
	})
	if err != nil {
		return dto.ConvertCurrencyResponse{}, err
	}
	return consumeResponse[dto.ConvertCurrencyResponse](resp, http.StatusOK)
}

// just like EstimateCurrencyConvert, but doesn't estimate
// but has different response
// thanks Xago, great design
func (bs *baseService) CurrencyConvert(ctx context.Context, pair dto.ConvertCurrencyPairEnum, amount float64) (string, error) {
	resp, err := bs.client.post(ctx, bs.client.baseURL, "currencyconvert", dto.ConvertCurrencyRequest{
		ConvertCurrencyPair: pair,
		Amount:              amount,
		EstimateCalculation: false,
	})
	if err != nil {
		return "", err
	}
	return consumeResponse[string](resp, http.StatusOK)
}
