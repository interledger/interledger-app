package unit

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var (
	ErrInternal = errors.New("unit: internal error")
	ErrUnauthorized = errors.New("unit: unauthorized webhook request")
)

type Service interface {
	VerifyWebhook(ctx context.Context, request *http.Request) error
}

type service struct {
	validator *validator.Validate
	webhookToken string
}

type ServiceArgs struct {
	WebhookToken string `validate:"required"`
}

func NewService(args ServiceArgs) (Service, error) {
	validator := validator.New()
	err := validator.Struct(args)
	if err != nil {
		return nil, err
	}

	return &service{
		validator: validator,
		webhookToken: args.WebhookToken,
	}, nil
}

func (self *service) VerifyWebhook(ctx context.Context, request *http.Request) error {
	signature := request.Header.Get("x-unit-signature")
	if signature == "" {
		return ErrInternal
	}

	mac := hmac.New(sha1.New, []byte(self.webhookToken))

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return ErrInternal
	}

	mac.Write(body)
	sha := hex.EncodeToString(mac.Sum(nil))

	if sha != signature {
		return ErrUnauthorized
	}

	return nil
}
