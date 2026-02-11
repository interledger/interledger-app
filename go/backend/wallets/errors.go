package wallets

import "errors"

var (
	ErrInternal        = errors.New("wallets: internal error")
	ErrNoWalletFound   = errors.New("wallets: no wallet found")
	ErrDuplicateWallet = errors.New("wallets: duplicate wallet")
	ErrWalletConflict  = errors.New("wallets: wallet conflict")
	ErrAddressExists   = errors.New("wallets: duplicate wallet")
	ErrInvalidAddress  = errors.New("wallets: invalid address")
)
