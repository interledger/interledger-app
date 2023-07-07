package paymentpointers

import (
	"database/sql/driver"
	"fmt"
	"net/url"
	"strings"
)

type PaymentPointer struct {
	url *url.URL
}

func Parse(rawPaymentPointer string) (PaymentPointer, error) {

	pp := standardize(rawPaymentPointer)

	ppURL, err := url.ParseRequestURI(pp)
	if err != nil {
		return PaymentPointer{}, err
	}

	return PaymentPointer{
		url: ppURL,
	}, nil
}

func (p *PaymentPointer) String() string {
	if p == nil || p.url == nil {
		return ""
	}
	return p.url.String()
}

func (p *PaymentPointer) ShortString() string {
	if p == nil || p.url == nil {
		return ""
	}
	s := p.url.String()
	return strings.Replace(s, "https://", "", 1)
}

// StandardisePaymentPointer takes in a payment pointer in either the forms:
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

	// Payment pointer URLs have at least one slash after the prefix, let the chips fall and the URL parsing fail
	if !strings.Contains(pp, "/") {
		return pp
	}

	// The payment pointer has no prefix assume we need to add https://
	return "https://" + pp
}

func (p PaymentPointer) Value() (driver.Value, error) {
	return p.String(), nil
}

func (p *PaymentPointer) Scan(src interface{}) error {
	if v, ok := src.(string); ok {
		pp, err := Parse(v)
		*p = pp
		return err
	}

	return fmt.Errorf("cannot convert %T to payment pointer", src)
}
