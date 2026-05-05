package dto

type Withdrawal struct {
	ID         string  `json:"id"`
	Total      float64 `json:"total"`
	Amount     float64 `json:"amount"`
	Commission float64 `json:"commission"`
	Status     string  `json:"status"`
	Currency   string  `json:"currencyCode"`
}
