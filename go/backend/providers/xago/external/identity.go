package external

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.com/fynbos/backend/providers/xago/external/domain/dto"
)

type identityService struct {
	client *client
}

// checked
func (is *identityService) ListBeneficiaries(ctx context.Context, limit, page uint) (dto.ListBeneficiariesResponse, error) {
	resp, err := is.client.get(ctx, is.client.identityURL, "beneficiaries",
		withQueryParam("limit", fmt.Sprintf("%d", limit)),
		withQueryParam("page", fmt.Sprintf("%d", page)),
	)
	if err != nil {
		return dto.ListBeneficiariesResponse{}, err
	}

	return consumeResponse[dto.ListBeneficiariesResponse](resp, http.StatusOK)
}

func (is *identityService) UpdateSubAccount(ctx context.Context, accountID string, req dto.UpdateSubAccountRequest) error {
	resp, err := is.client.put(ctx, is.client.identityURL, fmt.Sprintf("company/accounts/%v", accountID), req)
	if err != nil {
		return err
	}

	_, err = consumeResponse[struct{}](resp, http.StatusOK)
	return err
}

func (is *identityService) AddBeneficiary(ctx context.Context, req dto.CreateBeneficiaryRequest) (dto.AccountBeneficiaries, error) {
	resp, err := is.client.post(ctx, is.client.identityURL, "beneficiaries", req)
	if err != nil {
		return dto.AccountBeneficiaries{}, err
	}

	return consumeResponse[dto.AccountBeneficiaries](resp, http.StatusOK)
}
