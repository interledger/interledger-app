package dto

import "encoding/json"

type PaymentMethodTypeEnum string

func (pmt PaymentMethodTypeEnum) String() string {
	return string(pmt)
}

const (
	ACH_METHOD    PaymentMethodTypeEnum = "ACH"
	WALLET_METHOD PaymentMethodTypeEnum = "WALLET"
)

type PaymentMethod struct {
	PaymentMethodType PaymentMethodTypeEnum `json:"paymentMethodType"`
	*ACHPaymentMethod
	*WalletPaymentMethod
}

// MarshalJSON merges PaymentMethodType with the active embedded method's fields,
// avoiding silent field drops caused by name collisions between ACHPaymentMethod
// and WalletPaymentMethod when both are embedded as pointers.
func (p PaymentMethod) MarshalJSON() ([]byte, error) {
	type base struct {
		PaymentMethodType PaymentMethodTypeEnum `json:"paymentMethodType"`
	}

	baseBytes, err := json.Marshal(base{PaymentMethodType: p.PaymentMethodType})
	if err != nil {
		return nil, err
	}

	var innerBytes []byte
	if p.ACHPaymentMethod != nil {
		innerBytes, err = json.Marshal(p.ACHPaymentMethod)
	} else if p.WalletPaymentMethod != nil {
		innerBytes, err = json.Marshal(p.WalletPaymentMethod)
	}
	if err != nil {
		return nil, err
	}

	if len(innerBytes) == 0 || string(innerBytes) == "{}" {
		return baseBytes, nil
	}

	// merge: {"paymentMethodType":"ACH"} + {"currency":"USD",...} → {"paymentMethodType":"ACH","currency":"USD",...}
	result := append(baseBytes[:len(baseBytes)-1], ',')
	result = append(result, innerBytes[1:]...)
	return result, nil
}

func NewPaymentMethod(paymentMethod PaymentMethodTypeEnum, opts ...PaymentMethodOption) *PaymentMethod {
	p := &PaymentMethod{PaymentMethodType: paymentMethod}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

type PaymentMethodOption func(*PaymentMethod)

func WithACHPaymentMethod(achPaymentMethod ACHPaymentMethod) PaymentMethodOption {
	return func(p *PaymentMethod) {
		p.ACHPaymentMethod = &achPaymentMethod
	}
}

func WithWalletPaymentMethod(walletPaymentMethod WalletPaymentMethod) PaymentMethodOption {
	return func(p *PaymentMethod) {
		p.WalletPaymentMethod = &walletPaymentMethod
	}
}

type ACHPaymentMethod struct {
	SameDayACH         bool               `json:"sameDayAch,omitempty"`
	Currency           string             `json:"currency,omitempty"`
	BillingEmail       string             `json:"billingEmail,omitempty"`
	PaymentInformation PaymentInformation `json:"paymentInformation,omitempty"`
}

func (a ACHPaymentMethod) MarshalJSON() ([]byte, error) {
	type Alias ACHPaymentMethod
	return json.Marshal(Alias(a))
}

func (a *ACHPaymentMethod) UnmarshalJSON(data []byte) error {
	type Alias ACHPaymentMethod
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*a = ACHPaymentMethod(alias)
	return nil
}

type ACHPaymentMethodOption func(*ACHPaymentMethod)

func NewACHPaymentMethod(opts ...ACHPaymentMethodOption) *ACHPaymentMethod {
	a := &ACHPaymentMethod{}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func WithSameDayACH(sameDayACH bool) ACHPaymentMethodOption {
	return func(a *ACHPaymentMethod) {
		a.SameDayACH = sameDayACH
	}
}

func WithACHCurrency(currency string) ACHPaymentMethodOption {
	return func(a *ACHPaymentMethod) {
		a.Currency = currency
	}
}

func WithACHBillingEmail(email string) ACHPaymentMethodOption {
	return func(a *ACHPaymentMethod) {
		a.BillingEmail = email
	}
}

func WithACHPaymentInformation(paymentInformation PaymentInformation) ACHPaymentMethodOption {
	return func(a *ACHPaymentMethod) {
		a.PaymentInformation = paymentInformation
	}
}

type WalletPaymentMethodOption func(*WalletPaymentMethod)

func NewWalletPaymentMethod(opts ...WalletPaymentMethodOption) *WalletPaymentMethod {
	w := &WalletPaymentMethod{}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

func WithWalletBillingEmail(email string) WalletPaymentMethodOption {
	return func(w *WalletPaymentMethod) {
		w.BillingEmail = email
	}
}

func WithWalletPaymentInformation(paymentInformation Wallet) WalletPaymentMethodOption {
	return func(w *WalletPaymentMethod) {
		w.PaymentInformation = paymentInformation
	}
}

type WalletPaymentMethod struct {
	BillingEmail       string `json:"billingEmail,omitempty"`
	PaymentInformation Wallet `json:"paymentInformation,omitempty"`
}

func (w WalletPaymentMethod) MarshalJSON() ([]byte, error) {
	type Alias WalletPaymentMethod
	return json.Marshal(Alias(w))
}

func (w *WalletPaymentMethod) UnmarshalJSON(data []byte) error {
	type Alias WalletPaymentMethod
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*w = WalletPaymentMethod(alias)
	return nil
}
