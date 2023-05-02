package external

import (
	"fmt"

	"github.com/Basis-Theory/basistheory-go/v3"
)

type CardData struct {
	TokenizedNumber string `json:"number"`
	ExpirationMonth string `json:"expiration_month"`
	ExpirationYear  string `json:"expiration_year"`
}

func ExtractCardDataFrom(token *basistheory.Token) (*CardData, error) {
	if token == nil {
		return nil, fmt.Errorf("%w No token to extract card data.", ErrInternal)
	}
	if token.GetType() != "card" {
		return nil, fmt.Errorf("%w Token is not a card.", ErrInternal)
	}

	if token.GetData() == nil {
		return nil, fmt.Errorf("%w Token has no data.", ErrInternal)
	}

	tokenData, ok := token.GetData().(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%w Failed to extract card data from token.", ErrInternal)
	}

	return &CardData{
		TokenizedNumber: tokenData["number"].(string),
		ExpirationMonth: fmt.Sprintf("%02d", uint64(tokenData["expiration_month"].(float64))),
		ExpirationYear:  fmt.Sprintf("%d", uint64(tokenData["expiration_year"].(float64))),
	}, nil
}
