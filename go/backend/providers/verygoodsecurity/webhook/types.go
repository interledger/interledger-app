package webhook

type (
	Card struct {
		ID       string `json:"id,omitempty" db:"db"`
		Number   string `json:"card-number,string" db:"number"`
		Expiry   string `json:"exp-date,string" db:"expiry"`
		CVV      string `json:"card-security-code,string" db:"security_code"`
		WalletID string `json:"walletId,string" db:"wallet_id"`
		Last4    string `json:"last4,string" db:"last4"`
		Type     string `json:"cardType,string" db:"type"`
	}
)
