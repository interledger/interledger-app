package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"gitlab.com/fynbos/env"

	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func MakeAuditInterceptor(db *sqlx.DB) grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {

		if strings.HasPrefix(info.FullMethod, "/grpc.health.v1.Health/") {
			return handler(ctx, req)
		}

		adminUser, ok := ctx.Value(userCtxKey).(*AdminUser)
		if !ok || adminUser == nil {
			if env.FeatureEnforceAdminAuth() {
				return nil, ErrNoUserFound
			}
			adminUser = &AdminUser{
				Email: "local@admin.com",
			}
		}

		// Ignore the following methods from audit logging.
		if strings.Contains(info.FullMethod, "ListAudit") {
			return handler(ctx, req)
		}

		var walletID sql.NullString
		msg := req.(proto.Message)
		msg.ProtoReflect().Range(func(descriptor protoreflect.FieldDescriptor, value protoreflect.Value) bool {
			if strings.Contains(string(descriptor.FullName()), "WalletID") || strings.Contains(string(descriptor.FullName()), "walletID") {
				walletID = sql.NullString{
					String: value.String(),
					Valid:  true,
				}
				return false
			}
			return true
		})

		reqJson, err := json.Marshal(req)
		if err != nil {
			return nil, fmt.Errorf("%w %s", ErrInternal, err)
		}

		_, err = db.ExecContext(ctx, "INSERT INTO admin_audit_log(admin_user, wallet_id, operation, parameters) VALUES ($1, $2, $3, $4)",
			adminUser.Email, walletID, info.FullMethod, string(reqJson))
		if err != nil {
			return nil, fmt.Errorf("%w %s", ErrInternal, err)
		}

		return handler(ctx, req)
	})
}
