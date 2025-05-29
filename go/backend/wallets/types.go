package wallets

import (
	"database/sql/driver"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"gitlab.com/fynbos/backend/country"
)

type CreateArgs struct {
	ID        string `validate:"omitempty,uuid4"`
	UserID    string `validate:"uuid4"`
	Name      string
	Addresses []Address
	Country   country.Country
}

type Wallet struct {
	ID             string `db:"id"`
	Name           string `db:"name"`
	Addresses      []Address
	Country        country.Country `db:"country"`
	ExceededLimits bool            `db:"exceeded_limits"`
}

func (w Wallet) AddressShortString() string {
	if len(w.Addresses) == 0 {
		return ""
	}
	return w.Addresses[0].ShortString()
}

func (w Wallet) AddressString() string {
	if len(w.Addresses) == 0 {
		return ""
	}
	return w.Addresses[0].String()
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
	if p == nil || p.url == nil {
		return ""
	}
	return p.url.String()
}

func (p *Address) ShortString() string {
	if p == nil || p.url == nil {
		return ""
	}
	s := p.url.String()
	return strings.Replace(s, "https://", "", 1)
}

type ValidIdentityLength struct {
	Min int
	Max int
}

var identityLength = ValidIdentityLength{Min: 3, Max: 17}
var pattern = fmt.Sprintf(`^[A-Za-z]{%d}[a-zA-Z0-9_]{0,%d}$`, identityLength.Min, identityLength.Max)
var prefixPattern = fmt.Sprintf(`^[A-Za-z]{%d}$`, identityLength.Min)
var addressRegex = regexp.MustCompile(pattern)
var addressPrefixRegex = regexp.MustCompile(prefixPattern)
var ReservedURLParts = []string{"outgoing", "incoming", "quotes", "jwks.json", "identities"}

// TestAddress creates a Address without any of the validation, only a valid URL is required.
// To be used only for testing.
func TestAddress(_ *testing.T, address *url.URL) Address {
	return Address{url: address}
}

func ParseAddress(rawAddress string) (Address, error) {
	rawAddress = strings.TrimSuffix(rawAddress, "/")

	unescaped, err := url.PathUnescape(rawAddress)
	if err != nil || unescaped != rawAddress {
		// Some URL escapes where added or invalid URL escapes are present
		return Address{}, fmt.Errorf("%w %s", ErrInvalidAddress, "Your wallet address name can only contain letters, numbers and '_'")
	}

	wa := standardize(rawAddress)

	waURL, err := url.ParseRequestURI(wa)
	if err != nil {
		return Address{}, fmt.Errorf("%w %s", ErrInvalidAddress, err)
	}

	if waURL.Scheme == "" || waURL.Host == "" {
		return Address{}, fmt.Errorf("%w %s", ErrInvalidAddress, "Your wallet address name needs to contain a host and a http scheme")
	}

	// Fragments are after a '#' character in the url.
	// Payment pointers do not contain queries.
	if waURL.Fragment != "" || waURL.RawQuery != "" {
		return Address{}, fmt.Errorf("%w %s", ErrInvalidAddress, "Your wallet address name can only contain letters, numbers and '_'")
	}

	waPath := strings.TrimPrefix(waURL.Path, "/")

	if len(waPath) < identityLength.Min {
		return Address{}, fmt.Errorf("%w Your wallet address name must be longer than %d characters", ErrInvalidAddress, identityLength.Min)
	}
	if len(waPath) > identityLength.Max {
		return Address{}, fmt.Errorf("%w Your wallet address name must be shorter than %d characters", ErrInvalidAddress, identityLength.Max)
	}

	if strings.HasSuffix(waPath, "_") {
		return Address{}, fmt.Errorf("%w Your wallet address name can not end with an underscore", ErrInvalidAddress)
	}

	if strings.Contains(waPath, "__") {
		return Address{}, fmt.Errorf("%w Your wallet address name can not contain two or more consecutives underscores", ErrInvalidAddress)
	}

	if !addressPrefixRegex.MatchString(waPath[:3]) {
		return Address{}, fmt.Errorf("%w %s", ErrInvalidAddress, "Your first 3 characters must be letters")
	}

	if !addressRegex.MatchString(waPath) {
		return Address{}, fmt.Errorf("%w %s", ErrInvalidAddress, "Your wallet address name can only contain letters, numbers and '_'")
	}

	pathParts := strings.Split(waPath, "/")
	for _, pp := range pathParts {
		for _, res := range ReservedURLParts {
			if strings.EqualFold(pp, res) {
				return Address{}, fmt.Errorf("%w %s", ErrInvalidAddress, "Your wallet address name cannot contain reserved path parts")
			}
		}
	}

	return Address{
		url: waURL,
	}, nil
}

// standardize takes in a wallet address in either the forms:
// - https://ilp.link/alice
// - ilp.link/alice
// - $ilp.link/alice
// Returns the standard format of : https:///ilp.link/alice
func standardize(wa string) string {
	addr := strings.ToLower(wa)

	if strings.HasPrefix(addr, "https://") {
		return addr
	}

	// Replace the $ with https://
	if strings.HasPrefix(addr, "$") {
		return strings.Replace(addr, "$", "https://", 1)
	}

	// We use https here
	if strings.HasPrefix(addr, "http://") {
		return strings.Replace(addr, "http://", "https://", 1)
	}

	// Payment pointer URLs have at least one slash after the prefix, let the chips fall and the URL parsing fail
	if !strings.Contains(addr, "/") {
		return addr
	}

	// The payment pointer has no prefix assume we need to add https://
	return "https://" + addr
}

const (
	WebMonetizationWalletID = "31db2044-0b83-4aae-9cd8-b5b63cc85414"
	AstraBusinessWalletID   = "2893c8f2-97b0-42ed-a755-c075bdcf9ec5"
)
