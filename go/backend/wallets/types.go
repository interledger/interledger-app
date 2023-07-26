package wallets

import (
	"database/sql/driver"
	"fmt"
	"net/url"
	"strings"
)

type CreateArgs struct {
	ID        string `validate:"omitempty,uuid4"`
	UserID    string `validate:"uuid4"`
	Name      string
	Addresses []Address
}

type Wallet struct {
	ID        string
	Name      string
	Addresses []Address
}

type WalletCtxKey string

var CtxKey = WalletCtxKey("wallet_key")

type Address struct {
	url *url.URL
}

func (p Address) Value() (driver.Value, error) {
	return p.String(), nil
}

func (p *Address) Scan(src interface{}) error {
	if v, ok := src.(string); ok {
		pp, err := ParseAddress(v)
		*p = pp
		return err
	}

	return fmt.Errorf("cannot convert %T to wallet address", src)
}

func (p *Address) String() string {
	return p.url.String()
}

func (p *Address) ShortString() string {
	s := p.url.String()
	return strings.Replace(s, "https://", "$", 1)
}

func ParseAddress(rawAddress string) (Address, error) {

	pp := standardize(rawAddress)

	ppURL, err := url.ParseRequestURI(pp)
	if err != nil {
		return Address{}, err
	}

	return Address{
		url: ppURL,
	}, nil
}

// standardize takes in a wallet address in either the forms:
// - https://fynbos.me/alice
// - fynbos.me/alice
// - $fynbos.me/alice
// Returns the standard format of : https:///fynbos.me/alice
func standardize(wa string) string {
	if strings.HasPrefix(wa, "https://") {
		return wa
	}

	// Replace the $ with https://
	if strings.HasPrefix(wa, "$") {
		return strings.Replace(wa, "$", "https://", 1)
	}

	// We use https here
	if strings.HasPrefix(wa, "http://") {
		return strings.Replace(wa, "http://", "https://", 1)
	}

	// Payment pointer URLs have at least one slash after the prefix, let the chips fall and the URL parsing fail
	if !strings.Contains(wa, "/") {
		return wa
	}

	// The payment pointer has no prefix assume we need to add https://
	return "https://" + wa
}
