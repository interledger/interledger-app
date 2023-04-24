package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

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

		adminUser, ok := ctx.Value(userCtxKey).(*AdminUser)
		if !ok || adminUser == nil {
			return nil, ErrNoUserFound
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
