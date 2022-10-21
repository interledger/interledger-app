package admin

type AdminRpcService struct {
	b Backends
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
