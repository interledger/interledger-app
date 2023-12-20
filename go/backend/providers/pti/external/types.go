package external

import "time"

type (
	CreateUserArgs struct {
		ID            string    `json:"userId,omitempty"`
		Type          string    `json:"type,omitempty"`
		DateOfBirth   string    `json:"dateOfBirth,omitempty"`
		Name          Name      `json:"name,omitempty"`
		Emails        []Email   `json:"emails,omitempty"`
		Addresses     []Address `json:"addresses,omitempty"`
		Phones        []Phone   `json:"phones,omitempty"`
		SourceOfFunds string    `json:"sourceOfFunds,omitempty"`
	}

	CreateUserResponse struct {
		ID   string `json:"id,omitempty"`
		Link string `json:"link,omitempty"`
	}

	CreateWalletArgs struct {
		UserID    string `json:"-"`
		WalletID  string `json:"walletId,omitempty"`
		Currency  string `json:"currency,omitempty"`
		Reference string `json:"reference,omitempty"`
	}
)

type (
	Name struct {
		First  string `json:"firstName,omitempty"`
		Last   string `json:"lastName,omitempty"`
		Middle string `json:"middleName,omitempty"`
	}

	Email struct {
		Address string `json:"address,omitempty"`
		Default bool   `json:"default,omitempty"`
	}

	Phone struct {
		Number  string `json:"number,omitempty"`
		Type    string `json:"type,omitempty"`
		Default bool   `json:"default,omitempty"`
	}

	Address struct {
		Street     string `json:"streetAddress,omitempty"`
		City       string `json:"city,omitempty"`
		PostalCode string `json:"postalCode,omitempty"`
		StateCode  string `json:"stateCode,omitempty"`
		Country    string `json:"country,omitempty"`
		Default    bool   `json:"default,omitempty"`
	}

	Wallet struct {
		WalletID       string  `json:"walletId,omitempty"`
		Currency       string  `json:"currency,omitempty"`
		Reference      string  `json:"reference,omitempty"`
		CreateDateTime string  `json:"createDateTime,omitempty"`
		Balance        float64 `json:"balance"`
	}
)

type SignatureBase struct {
	Method      string
	Payload     []byte
	ContentType string
	Date        time.Time
	ClientID    string
	Path        string
}
