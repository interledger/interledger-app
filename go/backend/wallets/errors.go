package wallets

import "errors"

var (
	ErrInternal        = errors.New("wallets: internal error")
	ErrNoWalletFound   = errors.New("wallets: no wallet found")
	ErrDuplicateWallet = errors.New("wallets: duplicate wallet")
)
