package external

type TransactionType string

var (
	TransactionTypePull TransactionType = "pull"
	TransactionTypePush TransactionType = "push"
)

type TransactionStatus string

var (
	TransactionStatusOk        TransactionStatus = "OK"
	TransactionStatusLocked    TransactionStatus = "LOCKED"
	TransactionStatusDeleted   TransactionStatus = "DELETED"
	TransactionStatusPending   TransactionStatus = "PENDING"
	TransactionStatusBatch     TransactionStatus = "BATCH"
	TransactionStatusFailed    TransactionStatus = "FAILED"
	TransactionStatusUnknown   TransactionStatus = "UNKNOWN"
	TransactionStatusError     TransactionStatus = "ERROR"
	TransactionStatusCompleted TransactionStatus = "COMPLETED"
	TransactionStatusReversed  TransactionStatus = "REVERSED"
	TransactionStatusReversal  TransactionStatus = "REVERSAL"
)

type AccountType string

var (
	AccountTypeSavings          AccountType = "S"
	AccountTypeChecking         AccountType = "C"
	AccountTypeBusinessSavings  AccountType = "A"
	AccountTypeBusinessChecking AccountType = "B"
)

type (
	CreateTransactionArgs struct {
		ReferenceID string `json:"referenceID"`
		Type        TransactionType
		Accounts    CreateTransactionAccounts
		Currency    string `json:"currency,omitempty"`
		Amount      string `json:"amount"`
	}

	CreateTransactionAccounts struct {
		SourceAccountID      string         `json:"sourceAccountID,omitempty"`
		SourceAccount        *SourceAccount `json:"sourceAccount,omitempty"`
		DestinationAccountID string         `json:"destinationAccountID,omitempty"`
		DestinationAccount   *SourceAccount `json:"destinationAccount,omitempty"`
	}

	SourceAccount struct {
		Card  *Card `json:"card,omitempty"`
		Bank  *Bank `json:"bank,omitempty"`
		Owner Owner `json:"owner"`
	}

	Card struct {
		AccountNumber  string `json:"accountNumber,omitempty"`
		ExpirationDate string `json:"expirationDate,omitempty"`
		SecurityCode   string `json:"securityCode,omitempty"`
	}

	Bank struct {
		AccountNumber string      `json:"accountNumber"`
		RoutingNumber string      `json:"routingNumber"`
		AccountType   AccountType `json:"accountType"`
	}

	Owner struct {
		Name    Name
		Address *Address `json:"address,omitempty"`
		Phone   *Phone   `json:"phone,omitempty"`
	}

	Name struct {
		Company string `json:"company,omitempty"`
		First   string `json:"first,omitempty"`
		Middle  string `json:"middle,omitempty"`
		Last    string `json:"last,omitempty"`
		Suffix  string `json:"suffix,omitempty"`
	}

	Address struct {
		Line1   string `json:"line1,omitempty"`
		Line2   string `json:"line2,omitempty"`
		City    string `json:"city,omitempty"`
		State   string `json:"state,omitempty"`
		ZipCode string `json:"zipcode,omitempty"`
		Country string `json:"country,omitempty"`
	}

	Phone struct {
		CountryCode string `json:"countryCode,omitempty"`
		Number      string `json:"number"`
	}

	CreateTransactionResponse struct {
		SC            int    `json:"SC"`
		EC            string `json:"EC"`
		TransactionID string `json:"transactionID"`
		Network       string
		NetworkRC     string `json:"networkRC"`
		Status        string
		ApprovalCode  string `json:"approvalCode"`
		Errors        []string
		Fees          *Fees         `json:"fees,omitempty"`
		Card          *CardResponse `json:"card,omitempty"`
	}

	Fees struct {
		Interchange string
		Network     string
		Tabapay     string
	}

	CardResponse struct {
		NameFI         string `json:"nameFI,omitempty"`
		Last4          string `json:"last4"`
		ExpirationDate string `json:"expirationDate"`
	}

	RetrieveTransactionResponse struct {
		SC             int    `json:"SC"`
		EC             string `json:"EC"`
		ReferenceID    string `json:"referenceID"`
		Network        string
		NetworkRC      string `json:"networkRC"`
		Status         string
		Originally     string
		Amount         string
		AmountUSD      string    `json:"amountUSD"`
		Fees           *Fees     `json:"fees,omitempty"`
		ReversalStatus string    `json:"reversalStatus"`
		Reversal       *Reversal `json:"reversal,omitempty"`
	}

	Reversal struct {
		NetworkRC  string `json:"networkRC"`
		NetworkRC2 string `json:"networkRC2"`
		Error      string
	}

	RetrieveAccountResponse struct {
		SC          int           `json:"SC"`
		EC          string        `json:"EC"`
		ReferenceID string        `json:"referenceID"`
		Bank        *Bank         `json:"bank,omitempty"`
		Card        *CardResponse `json:"card,omitempty"`
		Owner       Owner
	}

	CreateAccountArgs struct {
		RejectDuplicateCard  bool   `json:"-"`
		OKToAddDuplicateCard bool   `json:"-"`
		ReferenceID          string `json:"referenceID"`
		Card                 Card
		Owner                Owner
	}

	CreateAccountResponse struct {
		SC        int           `json:"SC"`
		EC        string        `json:"EC"`
		AccountID string        `json:"accountID"`
		Card      *CardResponse `json:"card,omitempty"`
		Notices   string
	}
)
