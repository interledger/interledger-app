package grpc

import (
	"context"
	"errors"
	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/identities"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
	"net/url"
	"strings"
)

func (s *rpcService) CreateDNSIdentity(
	ctx context.Context,
	request *backendv1.CreateDNSIdentityRequest,
) (*backendv1.CreateDNSIdentityResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	domain, err := getDomain(s.b.Validator(), request.Url)
	if err != nil {
		return nil, err
	}

	identity, err := s.b.Identities().Add(ctx, identities.AddArgs{
		WalletID:   wallet.ID,
		Platform:   identities.PlatformDNS,
		Identifier: domain,
	})
	if err != nil {
		if errors.Is(err, identities.ErrAlreadyExists) {
			return nil, AlreadyExistsError("Identity already exists.")
		}
		return nil, toGRPCError(err)
	}

	return &backendv1.CreateDNSIdentityResponse{
		Id: identity.ID,
	}, nil
}

func getDomain(validator *validator.Validate, inputUrl string) (string, error) {
	err := validator.Var(inputUrl, "required,fqdn")
	if err == nil {
		domain := strings.TrimPrefix(inputUrl, "www.")
		return domain, nil
	}

	err = validator.Var(inputUrl, "required,url")
	if err != nil {
		return "", NewValidationError("Domain", "Please enter a valid domain.")
	}

	websiteUrl, err := url.Parse(inputUrl)
	if err != nil {
		return "", NewValidationError("Domain", "Please enter a valid domain.")
	}

	domain := strings.TrimPrefix(websiteUrl.Hostname(), "www.")

	if domain == "" {
		return "", NewValidationError("Domain", "Please enter a valid domain.")
	}

	return domain, nil
}
