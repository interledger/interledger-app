package user

import "errors"

var (
	ErrInternal        = errors.New("user: internal error")
	ErrNoUserFound     = errors.New("no user found")
	ErrNoWalletFound   = errors.New("no wallet found")
	ErrDuplicateWallet = errors.New("user: duplicate wallet")
)
