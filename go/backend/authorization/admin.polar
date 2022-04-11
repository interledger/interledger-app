allow(user: AdminUser, action: String, _: Account) if
	user.Email in ["don@fynbos.dev", "matt@fynbos.dev", "cairin@fynbos.dev"] and
	action in ["read"];

actor AdminUser {}
