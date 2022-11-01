package machnet

import "errors"

var (
	ErrInternal              = errors.New("machnet: internal error")
	ErrNotFound              = errors.New("machnet: not found")
	ErrIncompleteKYC         = errors.New("machnet: incomplete kyc information")
	ErrInvalidSignature      = errors.New("machnet: invalid signature")
	ErrInvalidArgument       = errors.New("machnet: invalid argument")
	ErrUserHasExistingWallet = errors.New("machnet: user has an existing wallet")
)
