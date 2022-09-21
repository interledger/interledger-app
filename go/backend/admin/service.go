package admin

import (
	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/admin/auth"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
	"go.temporal.io/sdk/client"
)

type AdminRpcService struct {
	backendv1.UnimplementedBackendAdminServiceServer
	Validator   *validator.Validate
	AuthService auth.Service
	Temporal    client.Client
}

//func authorizeAdmin(email string) bool {
//	emails := [...]string{
//		"don@fynbos.dev",
//		"matt@fynbos.dev",
//		"cairin@fynbos.dev",
//		"adrian@fynbos.dev",
//	}
//	for _, e := range emails {
//		if e == email {
//			return true
//		}
//	}
//	return false
//}
