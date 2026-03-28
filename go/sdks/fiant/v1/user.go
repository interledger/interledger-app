package v1

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"gitlab.com/fynbos/sdks/fiant/v1/domain/dto"
)

type usersService struct {
	client *Client
}

// https://developers.platform.fiant.io/reference/addauser
func (us *usersService) Create(ctx context.Context, user dto.User) (dto.User, error) {
	path := "users" // POST /users
	resp, err := us.client.post(ctx, path, user)
	if err != nil {
		return dto.User{}, err
	}
	return consumeResponse[dto.User](resp, http.StatusCreated)
}

func (us *usersService) Get(ctx context.Context, userID string) (dto.User, error) {
	path := fmt.Sprintf("users/%v", userID) // GET /users/{userId}
	resp, err := us.client.get(ctx, path)
	if err != nil {
		return dto.User{}, err
	}
	return consumeResponse[dto.User](resp, http.StatusOK)
}

// https://developers.platform.fiant.io/reference/getlistofusers
func (us *usersService) ListAll(ctx context.Context) (dto.UserPage, error) {
	path := "users" // GET /users
	resp, err := us.client.get(ctx, path)
	if err != nil {
		return dto.UserPage{}, err
	}
	return consumeResponse[dto.UserPage](resp, http.StatusOK)
}

// https://developers.platform.fiant.io/reference/getlastkyc
func (us *usersService) GetAssessment(ctx context.Context, user dto.User) (dto.UserAssessment, error) {
	path := fmt.Sprintf("users/%v/assessments", user.ID) // GET /users/{userId}/assessments
	resp, err := us.client.get(ctx, path)
	if err != nil {
		return dto.UserAssessment{}, err
	}
	return consumeResponse[dto.UserAssessment](resp, http.StatusOK)
}

// https://developers.platform.fiant.io/reference/startuserassessment
func (us *usersService) StartAssessment(ctx context.Context, user dto.User, scenarioID string) (dto.ObjectReference, error) {
	path := fmt.Sprintf("users/%v/assessments", user.ID) // POST /users/{userId}/assessments
	resp, err := us.client.post(ctx, path, struct {
		ScenarioID string `json:"scenarioId"`
		dto.User
	}{
		ScenarioID: scenarioID,
		User:       user,
	},
		withHeader(ptiRequestIDHeader, uuid.NewString()),
		withHeader(ptiScenarioIDHeader, scenarioID),
	)
	if err != nil {
		return dto.ObjectReference{}, err
	}
	return consumeResponse[dto.ObjectReference](resp, http.StatusCreated)
}

// https://developers.platform.fiant.io/reference/createwallet
func (us *usersService) CreateWallet(ctx context.Context, user dto.User, wallet dto.Wallet) (dto.Wallet, error) {
	path := fmt.Sprintf("users/%v/wallets", user.ID) // POST /users/{userId}/wallets
	resp, err := us.client.post(ctx, path, wallet)
	if err != nil {
		return dto.Wallet{}, err
	}
	return consumeResponse[dto.Wallet](resp, http.StatusCreated)
}

// https://developers.platform.fiant.io/reference/getwallet
func (us *usersService) GetWallet(ctx context.Context, user dto.User, walletID string) (dto.Wallet, error) {
	path := fmt.Sprintf("users/%v/wallets/%v", user.ID, walletID) // GET /users/{userId}/wallets/{walletId}
	resp, err := us.client.get(ctx, path)
	if err != nil {
		return dto.Wallet{}, err
	}
	return consumeResponse[dto.Wallet](resp, http.StatusOK)
}

// https://developers.platform.fiant.io/reference/getwallets
func (us *usersService) ListWallets(ctx context.Context, user dto.User) ([]dto.Wallet, error) {
	path := fmt.Sprintf("users/%v/wallets", user.ID) // GET /users/{userId}/wallets
	resp, err := us.client.get(ctx, path)
	if err != nil {
		return nil, err
	}
	return consumeResponse[[]dto.Wallet](resp, http.StatusOK)
}

// https://developers.platform.fiant.io/reference/getuserpaymentinformations
func (us *usersService) GetPaymentInformations(ctx context.Context, user dto.User) ([]dto.PaymentInformation, error) {
	path := fmt.Sprintf("users/%v/payment-information", user.ID) // GET /users/{userId}/payment-information
	resp, err := us.client.get(ctx, path)
	if err != nil {
		return nil, err
	}
	return consumeResponse[[]dto.PaymentInformation](resp, http.StatusOK)
}

// https://developers.platform.fiant.io/reference/adduserpaymentinformation
func (us *usersService) AddPaymentInformation(ctx context.Context, user dto.User, paymentInformation dto.PaymentInformation) (dto.PaymentInformation, error) {
	path := fmt.Sprintf("users/%v/payment-information", user.ID) // POST /users/{userId}/payment-information
	resp, err := us.client.post(ctx, path, paymentInformation)
	if err != nil {
		return dto.PaymentInformation{}, err
	}
	return consumeResponse[dto.PaymentInformation](resp, http.StatusCreated)
}

// https://developers.platform.fiant.io/reference/getuserpaymentinformation
func (us *usersService) GetPaymentInformation(ctx context.Context, user dto.User, paymentInformationID string) (dto.PaymentInformation, error) {
	path := fmt.Sprintf("users/%v/payment-information/%v", user.ID, paymentInformationID) // GET /users/{userId}/payment-information/{paymentInformationId}
	resp, err := us.client.get(ctx, path)
	if err != nil {
		return dto.PaymentInformation{}, err
	}
	return consumeResponse[dto.PaymentInformation](resp, http.StatusOK)
}
