# List available SSH roles
path "ssh-client-signer/roles/*" {
	capabilities = ["list"]
}

# Allow access to SSH admin role
path "ssh-client-signer/sign/admin-role" {
	capabilities = ["create","update"]
}