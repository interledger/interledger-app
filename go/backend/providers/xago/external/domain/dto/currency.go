package dto

type Currency struct {
	CurrencyCode     string         `json:"currencyCode"`
	Name             string         `json:"name"`
	Symbol           string         `json:"symbol"`
	DepositEnabled   bool           `json:"depositEnabled"`
	WithdrawEnabled  bool           `json:"withdrawEnabled"`
	MarketEnabled    bool           `json:"marketEnabled"`
	BankingProviders []BankProvider `json:"bankingProviders"`
}

type BankProvider struct {
	Name             string        `json:"name"`
	DepositAvailable bool          `json:"depositAvailable"`
	DepositFields    DepositFields `json:"depositFields"`
}

type DepositFields struct {
	BankName       string `json:"bankName"`
	AccountName    string `json:"accountName"`
	AccountNumber  string `json:"accountNumber"`
	BankAddress    string `json:"bankAddress"`
	AccountAddress string `json:"accountAddress"`
	BranchCode     string `json:"branchCode"`
}

type ConvertCurrencyPairEnum string

func (cc ConvertCurrencyPairEnum) String() string {
	return string(cc)
}

const (
	ZARtoEUR ConvertCurrencyPairEnum = "ZAR/EUR"
	EURtoZAR ConvertCurrencyPairEnum = "EUR/ZAR"
)

type ConvertCurrencyRequest struct {
	ConvertCurrencyPair ConvertCurrencyPairEnum `json:"convertCurrencyPair"`

	Amount              float64 `json:"amount"`
	EstimateCalculation bool    `json:"estimateCalculation"`
}

type ConvertCurrencyResponse struct {
	BuyAveragePrice float64 `json:"buyAveragePrice"`
	BuyOrders       float64 `json:"buyOrders"`
	EstimatedRate   float64 `json:"estimatedRate"`
	FinalBuyAmount  float64 `json:"finalBuyAmount"`
	FinalSellAmount float64 `json:"finalSellAmount"`
	QuoteAmount     int     `json:"quoteAmount"`
	ReceivedAmount  float64 `json:"receivedAmount"`
	SellOrders      float64 `json:"sellOrders"`
}
