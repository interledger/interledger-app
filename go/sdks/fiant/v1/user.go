package v1

import (
	"context"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"gitlab.com/fynbos/sdks/fiant/v1/domain/dto"
)

type userHandler struct {
	path string
	ctrl *Controller
}

// https://developers.platform.fiant.io/reference/addauser
func (uh *userHandler) Create(ctx context.Context, user dto.User) (dto.User, error) {
	resp, err := uh.ctrl.post(ctx, uh.path, user) // POST /users
	if err != nil {
		return dto.User{}, err
	}
	return consumeResponse[dto.User](resp, http.StatusCreated)
}

func (uh *userHandler) Get(ctx context.Context, userID string) (dto.User, error) {
	path, err := url.JoinPath(uh.path, userID)
	if err != nil {
		return dto.User{}, err
	}
	resp, err := uh.ctrl.get(ctx, path) // GET /users/{userId}
	if err != nil {
		return dto.User{}, err
	}
	return consumeResponse[dto.User](resp, http.StatusOK)
}

// https://developers.platform.fiant.io/reference/getlistofusers
func (uh *userHandler) ListAll(ctx context.Context) (dto.UserPage, error) {
	resp, err := uh.ctrl.get(ctx, uh.path) // GET /users
	if err != nil {
		return dto.UserPage{}, err
	}
	return consumeResponse[dto.UserPage](resp, http.StatusOK)
}

// https://developers.platform.fiant.io/reference/getlastkyc
func (uh *userHandler) GetAssessment(ctx context.Context, user dto.User) (dto.UserAssessment, error) {
	path, err := url.JoinPath(uh.path, user.ID, "assessments")
	if err != nil {
		return dto.UserAssessment{}, err
	}
	resp, err := uh.ctrl.get(ctx, path) // GET /users/{userId}/assessments
	if err != nil {
		return dto.UserAssessment{}, err
	}
	return consumeResponse[dto.UserAssessment](resp, http.StatusOK)
}

// https://developers.platform.fiant.io/reference/startuserassessment
func (uh *userHandler) StartAssessment(ctx context.Context, user dto.User, scenarioID string) (dto.ObjectReference, error) {
	path, err := url.JoinPath(uh.path, "assessments")
	if err != nil {
		return dto.ObjectReference{}, err
	}
	resp, err := uh.ctrl.post(ctx, path, struct {
		ScenarioID string `json:"scenarioId"`
		dto.User
	}{
		ScenarioID: scenarioID,
		User:       user,
	},
		withHeader(ptiRequestIDHeader, uuid.NewString()),
		withHeader(ptiScenarioIDHeader, scenarioID),
	) // POST /users/assessments
	if err != nil {
		return dto.ObjectReference{}, err
	}
	return consumeResponse[dto.ObjectReference](resp, http.StatusCreated)
}

// https://developers.platform.fiant.io/reference/createwallet
func (uh *userHandler) CreateWallet(ctx context.Context, user dto.User, wallet dto.Wallet) (dto.Wallet, error) {
	path, err := url.JoinPath(uh.path, user.ID, "wallets")
	if err != nil {
		return dto.Wallet{}, err
	}
	resp, err := uh.ctrl.post(ctx, path, wallet) // POST /users/{userId}/wallets
	if err != nil {
		return dto.Wallet{}, err
	}
	return consumeResponse[dto.Wallet](resp, http.StatusCreated)
}

// https://developers.platform.fiant.io/reference/getwallet
func (uh *userHandler) GetWallet(ctx context.Context, user dto.User, walletID string) (dto.Wallet, error) {
	path, err := url.JoinPath(uh.path, user.ID, "wallets", walletID)
	if err != nil {
		return dto.Wallet{}, err
	}
	resp, err := uh.ctrl.get(ctx, path) // GET /users/{userId}/wallets/{walletId}
	if err != nil {
		return dto.Wallet{}, err
	}
	return consumeResponse[dto.Wallet](resp, http.StatusOK)
}

// https://developers.platform.fiant.io/reference/getwallets
func (uh *userHandler) ListWallets(ctx context.Context, user dto.User) ([]dto.Wallet, error) {
	path, err := url.JoinPath(uh.path, user.ID, "wallets")
	if err != nil {
		return nil, err
	}
	resp, err := uh.ctrl.get(ctx, path) // GET /users/{userId}/wallets
	if err != nil {
		return nil, err
	}
	return consumeResponse[[]dto.Wallet](resp, http.StatusOK)
}
