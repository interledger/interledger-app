package user

import "errors"

var ErrNoUserFound = errors.New("no user found")
var ErrNoWalletFound = errors.New("no wallet found")
