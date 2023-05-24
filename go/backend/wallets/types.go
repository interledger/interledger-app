package wallets

import (
	"database/sql/driver"
	"fmt"
	"net/url"
	"strings"
)

type CreateArgs struct {
	UserID    string `validate:"uuid4"`
	Name      string
	Addresses []Address
}

type Wallet struct {
	ID        string
	Name      string
	Addresses []Address
}

type Address struct {
	url *url.URL
}

func (p Address) Value() (driver.Value, error) {
	return p.String(), nil
}

func (p *Address) Scan(src interface{}) error {
	if v, ok := src.(string); ok {
		pp, err := parseAddress(v)
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

func parseAddress(rawAddress string) (Address, error) {

	pp := standardize(rawAddress)

	ppURL, err := url.Parse(pp)
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
func standardize(pp string) string {
	if strings.HasPrefix(pp, "https://") {
		return pp
	}

	// Replace the $ with https://
	if strings.HasPrefix(pp, "$") {
		return strings.Replace(pp, "$", "https://", 1)
	}

	// We use https here
	if strings.HasPrefix(pp, "http://") {
		return strings.Replace(pp, "http://", "https://", 1)
	}

	// The payment pointer has no prefix assume we need to add https://
	return "https://" + pp
}
