package v1

import (
	"context"
	"net/http"

	"gitlab.com/fynbos/sdks/fiant/v1/domain/dto"
)

type authService struct {
	client *Client
}

// https://developers.platform.fiant.io/reference/getusertoken
func (as *authService) GetToken(ctx context.Context, URL, method string) (dto.UserToken, error) {
	path := "auth/jwt" // POST /auth/jwt

	resp, err := as.client.post(ctx, path, struct {
		URL    string `json:"url"`
		Method string `json:"method"`
	}{
		URL:    URL,
		Method: method,
	})
	if err != nil {
		return dto.UserToken{}, err
	}

	return consumeResponse[dto.UserToken](resp, http.StatusOK)
}
