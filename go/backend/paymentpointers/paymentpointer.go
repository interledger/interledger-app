package paymentpointers

import (
	"net/url"
	"strings"
)

type PaymentPointer struct {
	url *url.URL
}

func Parse(rawPaymentPointer string) (*PaymentPointer, error) {

	pp := standardizePaymentPointer(rawPaymentPointer)

	ppURL, err := url.Parse(pp)
	if err != nil {
		return nil, err
	}

	return &PaymentPointer{
		url: ppURL,
	}, nil
}

func (p *PaymentPointer) String() string {
	return p.url.String()
}

func (p *PaymentPointer) ShortString() string {
	s := p.url.String()
	return strings.Replace(s, "https://", "$", 1)
}

// StandardisePaymentPointer takes in a payment pointer in either the forms:
// - https://fynbos.me/alice
// - fynbos.me/alice
// - $fynbos.me/alice
// Returns the standard format of : https:///fynbos.me/alice
func standardizePaymentPointer(pp string) string {
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
