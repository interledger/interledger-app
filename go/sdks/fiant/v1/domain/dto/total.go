package dto

import "encoding/json"

type TotalOption func(*Total)

func WithTotalFee(fee *Cost) TotalOption {
	return func(t *Total) {
		t.Fee = fee
	}
}

func WithTotalTotal(total *Cost) TotalOption {
	return func(t *Total) {
		t.Total = total
	}
}

func WithTotalSubtotal(subtotal *Cost) TotalOption {
	return func(t *Total) {
		t.Subtotal = subtotal
	}
}

// https://github.com/provenancetech/pti-platform-sdks/blob/master/java/src/main/java/com/pti/sdk/types/Total.java
type Total struct {
	Fee      *Cost `json:"fee,omitempty"`
	Total    *Cost `json:"total,omitempty"`
	Subtotal *Cost `json:"subtotal,omitempty"`
}

func NewTotal(opts ...TotalOption) *Total {
	t := &Total{}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func (t Total) MarshalJSON() ([]byte, error) {
	type alias Total
	return json.Marshal(alias(t))
}

func (t *Total) UnmarshalJSON(data []byte) error {
	type alias Total
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*t = Total(a)
	return nil
}
