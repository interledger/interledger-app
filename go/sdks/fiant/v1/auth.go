package v1

import (
	"context"
	"net/http"
	"net/url"

	"gitlab.com/fynbos/sdks/fiant/v1/domain/dto"
)

type authHandler struct {
	path string
	ctrl *Controller
}

// https://developers.platform.fiant.io/reference/getusertoken
func (ah *authHandler) GetToken(ctx context.Context, URL, method string) (dto.UserToken, error) {
	path, err := url.JoinPath(ah.path, "jwt")
	if err != nil {
		return dto.UserToken{}, err
	}

	resp, err := ah.ctrl.post(ctx, path, struct {
		URL    string `json:"url"`
		Method string `json:"method"`
	}{
		URL:    URL,
		Method: method,
	}) // POST /auth/jwt
	if err != nil {
		return dto.UserToken{}, err
	}

	return consumeResponse[dto.UserToken](resp, http.StatusOK)
}
