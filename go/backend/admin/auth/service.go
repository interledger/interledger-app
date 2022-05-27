package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	ErrInvalidArgument = errors.New("admin user: invalid argument.")
	ErrNoUserFound     = errors.New("admin user: invalid user.")
	ErrInvalidToken    = errors.New("admin user: invalid token.")
	ErrInternal        = errors.New("admin user: internal error.")
)

type Service interface {
	GetAdminUser(ctx context.Context) (*AdminUser, error)
	ForContext(ctx context.Context) (*AdminUser, error)
	MakeUnaryInterceptors(serviceName string) grpc.ServerOption
}

type contextKey struct {
	name string
}

var userCtxKey = &contextKey{"adminUser"}

type AdminUser struct {
	Email string
}

type IDTokenClaims struct {
	Iss           string
	Azp           string
	Aud           string
	Sub           string
	Hd            string
	Iat           uint64
	Exp           uint64
	AtHash        string `json:"at_hash"`
	Email         string
	EmailVerified bool `json:"email_verified"`
}

type service struct {
	verifier *oidc.IDTokenVerifier
}

func NewService(oauth2ClientID string) (Service, error) {
	if oauth2ClientID == "" {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, "Oauth2 client ID is required.")
	}

	provider, err := oidc.NewProvider(context.Background(), "https://accounts.google.com")
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &service{
		verifier: provider.Verifier(&oidc.Config{
			ClientID: oauth2ClientID,
		}),
	}, nil
}

func (s *service) GetAdminUser(ctx context.Context) (*AdminUser, error) {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrNoUserFound
	}
	idTokens := meta.Get("idToken") // must match the metadata field key configured in Retool
	if len(idTokens) < 1 {
		return nil, ErrNoUserFound
	}
	idToken, err := s.verifier.Verify(ctx, idTokens[0])
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidToken, err)
	}

	claims := IDTokenClaims{}
	err = idToken.Claims(&claims)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	if !claims.EmailVerified || claims.Email == "" {
		return nil, fmt.Errorf("%w %s", ErrNoUserFound, "Unverified or invalid email address.")
	}

	return &AdminUser{Email: claims.Email}, nil
}

func (s *service) ForContext(ctx context.Context) (*AdminUser, error) {
	raw, ok := ctx.Value(userCtxKey).(*AdminUser)
	if !ok || raw == nil {
		return nil, ErrNoUserFound
	}

	return raw, nil
}

func (s *service) MakeUnaryInterceptors(serviceName string) grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if !strings.Contains(info.FullMethod, "BackendAdminService") {
			return handler(ctx, req)
		}

		user, err := s.GetAdminUser(ctx)
		if err != nil {
			if !errors.Is(err, ErrNoUserFound) {
				return nil, status.Error(codes.Internal, "error parsing id token.")
			}
		}

		newCtx := ctx
		if user != nil {
			newCtx = context.WithValue(ctx, userCtxKey, user)
		}

		return handler(newCtx, req)
	})
}
