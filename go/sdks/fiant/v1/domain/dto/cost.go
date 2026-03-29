package dto

import "encoding/json"

type CostOption func(*Cost)

func WithCostAmount(amount float64) CostOption {
	return func(c *Cost) {
		c.Amount = amount
	}
}

func WithCostCurrency(currency string) CostOption {
	return func(c *Cost) {
		c.Currency = currency
	}
}

// https://github.com/provenancetech/pti-platform-sdks/blob/master/java/src/main/java/com/pti/sdk/types/Cost.java
type Cost struct {
	Amount   float64 `json:"amount,omitempty"`
	Currency string  `json:"currency,omitempty"`
}

func NewCost(opts ...CostOption) *Cost {
	c := &Cost{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c Cost) MarshalJSON() ([]byte, error) {
	type alias Cost
	return json.Marshal(alias(c))
}

func (c *Cost) UnmarshalJSON(data []byte) error {
	type alias Cost
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*c = Cost(a)
	return nil
}
