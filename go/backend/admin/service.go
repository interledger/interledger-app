package admin

import (
	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/admin/auth"
	"gitlab.com/fynbos/backend/identity"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
	"go.temporal.io/sdk/client"
)

type AdminRpcService struct {
	backendv1.UnimplementedBackendAdminServiceServer
	Validator       *validator.Validate
	IdentityService identity.Client
	AuthService     auth.Service
	Temporal        client.Client
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
