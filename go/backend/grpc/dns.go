package grpc

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/identities"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
	"net/url"
)

func (s *rpcService) CreateDNSIdentity(
	ctx context.Context,
	request *backendv1.CreateDNSIdentityRequest,
) (*backendv1.CreateDNSIdentityResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	domain, err := getDomain(s.b.Validator(), request.Url)
	if err != nil {
		return nil, toGRPCError(err)
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

	err = s.b.Identities().SetProof(ctx, identity.ID, domain)
	if err != nil {
		return nil, toGRPCError(err)
	}

	sighash := base64.URLEncoding.EncodeToString(identity.SignatureHash)
	txtRecordPrefix := "_fynbos"

	return &backendv1.CreateDNSIdentityResponse{
		TxtRecord: fmt.Sprintf("%s.%s=%s", txtRecordPrefix, identity.Identifier, sighash),
	}, nil
}

func getDomain(validator *validator.Validate, inputUrl string) (string, error) {
	err := validator.Var(inputUrl, "required,fqdn")
	if err == nil {
		return inputUrl, nil
	}

	err = validator.Var(inputUrl, "required,url")
	if err == nil {
		websiteUrl, err := url.Parse(inputUrl)
		if err != nil {
			return "", err
		}

		hostname := websiteUrl.Hostname()
		if len(hostname) > 4 && hostname[:4] == "www." {
			hostname = hostname[4:]
		}

		return hostname, nil
	}

	err = validator.Var(inputUrl, "required,url|fqdn")
	if err != nil {
		return "", err
	}

	return "", InternalError("Invalid domain.")
}
