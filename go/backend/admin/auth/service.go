package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/interledger/interledger-app/go/env"
	"github.com/interledger/interledger-app/go/log"
	"go.uber.org/zap"
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
	ForContext(ctx context.Context) (*AdminUser, error)
	MakeUnaryInterceptors() []grpc.ServerOption
}

type contextKey struct {
	name string
}

var userCtxKey = &contextKey{"adminUser"}

type AdminUser struct {
	Email string
}

type IDTokenClaims struct {
	Email string
}

type service struct {
	verifier *oidc.IDTokenVerifier
	db       *sqlx.DB
}

func NewService(policyAud, teamDomain string, db *sqlx.DB) (Service, error) {

	if !env.IsLocal() {
		if policyAud == "" {
			return nil, fmt.Errorf("%w %s", ErrInvalidArgument, "Policy audience required.")
		}
		if teamDomain == "" {
			return nil, fmt.Errorf("%w %s", ErrInvalidArgument, "Team domain required.")
		}
		ctx := context.Background()

		config := &oidc.Config{
			ClientID: policyAud,
		}
		certsURL := fmt.Sprintf("%s/cdn-cgi/access/certs", teamDomain)
		keySet := oidc.NewRemoteKeySet(ctx, certsURL)
		verifier := oidc.NewVerifier(teamDomain, keySet, config)

		return &service{
			verifier: verifier,
			db:       db,
		}, nil
	}

	return &service{db: db}, nil
}

func (s *service) verifyToken(ctx context.Context) (*oidc.IDToken, error) {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Make sure that the incoming request has our token header
	//  Could also look in the cookies for CF_AUTHORIZATION
	rawCookies := meta.Get("cookies")
	if len(rawCookies) == 0 {
		return nil, ErrInvalidToken
	}

	cookies := strings.Split(rawCookies[0], ";")

	var cfCookie string
	for _, c := range cookies {
		k, v, _ := strings.Cut(c, "=")
		trimmed := strings.Trim(k, " ")
		if trimmed == "CF_Authorization" {
			cfCookie = v
		}
	}
	if cfCookie == "" {
		return nil, ErrInvalidToken
	}

	token, err := s.verifier.Verify(ctx, cfCookie)
	if err != nil {
		return nil, ErrInvalidToken
	}

	return token, nil
}

func (s *service) ForContext(ctx context.Context) (*AdminUser, error) {
	raw, ok := ctx.Value(userCtxKey).(*AdminUser)
	if !ok || raw == nil {
		return nil, ErrNoUserFound
	}

	return raw, nil
}

func (s *service) MakeUnaryInterceptors() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(func(
			ctx context.Context,
			req interface{},
			info *grpc.UnaryServerInfo,
			handler grpc.UnaryHandler,
		) (interface{}, error) {

			if strings.HasPrefix(info.FullMethod, "/grpc.health.v1.Health/") {
				return handler(ctx, req)
			}

			newCtx := ctx
			if !env.IsLocal() {
				token, err := s.verifyToken(ctx)
				if err != nil {
					return nil, status.Error(codes.Unauthenticated, "token not verified")
				}

				var claims IDTokenClaims
				if err := token.Claims(&claims); err != nil {
					return nil, status.Error(codes.Internal, "error parsing claims")
				}

				user := &AdminUser{
					Email: claims.Email,
				}
				newCtx = context.WithValue(ctx, userCtxKey, user)
				log.Info("admin access", zap.String("email", user.Email), zap.String("route", info.FullMethod))
			}

			return handler(newCtx, req)
		}),
		MakeAuditInterceptor(s.db),
	}
}
