package external

import "fmt"

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

type CardType string

var (
	CardTypeCredit  CardType = "Credit"
	CardTypeDebit   CardType = "Debit"
	CardTypePrepaid CardType = "Prepaid"
)

type DeviceChannelType string

var (
	DeviceChannelBrowser DeviceChannelType = "Browser"
	DeviceChannelSDK     DeviceChannelType = "SDK"
)

type BrowserInfo string

func NewBrowserInfo(fields BrowserInfoFields) BrowserInfo {
	return BrowserInfo(
		fmt.Sprintf(
			"%t|%s|%s|%t|%s|%s|%s|%s|%s|%s",
			fields.JavascriptEnabled,
			fields.UserAgent,
			fields.Header,
			fields.JavaEnabled,
			fields.Language,
			fields.ColorDepth,
			fields.ScreenHeight,
			fields.ScreenWidth,
			fields.TimeZone,
			fields.IpAddress,
		),
	)
}

type ColorDepth string

var (
	ColorDepth1  ColorDepth = "1"
	ColorDepth4  ColorDepth = "4"
	ColorDepth8  ColorDepth = "8"
	ColorDepth15 ColorDepth = "15"
	ColorDepth16 ColorDepth = "16"
	ColorDepth24 ColorDepth = "24"
	ColorDepth32 ColorDepth = "32"
	ColorDepth48 ColorDepth = "48"
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
		NameFI         string     `json:"nameFI,omitempty"`
		Last4          string     `json:"last4"`
		ExpirationDate string     `json:"expirationDate,omitempty"`
		Push           PushObject `json:"push,omitempty"`
		Pull           PullObject `json:"pull,omitempty"`
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

	QueryCardArgs struct {
		Card  *Card  `json:"card,omitempty"`
		Owner *Owner `json:"owner,omitempty"`
	}

	PullObject struct {
		Enabled   string
		Network   string
		Type      CardType
		Regulated bool
		Currency  string
		Country   string
	}

	PushObject struct {
		Enabled      string
		Network      string
		Type         CardType
		Regulated    bool
		Currency     string
		Country      string
		Availability string
	}

	QueryCardResponse struct {
		SC   int          `json:"SC"`
		EC   string       `json:"EC"`
		EM   string       `jsonL:"EM"`
		Card CardResponse `json:"card,omitempty"`
	}

	Account struct {
		AccountID string `json:"accountID"`
		Owner     *Owner `json:"owner,omitempty"`
	}

	Order struct {
		OrderID  string `json:"orderID"`
		Currency string
		Amount   string
	}

	Init3DSArgs struct {
		Account Account `json:"account"`
		Order   Order
	}

	Init3DSResponse struct {
		SC                  int    `json:"SC"`
		EC                  string `json:"EC"`
		EM                  string `jsonL:"EM"`
		ID3DS               string `json:"3dsID"`
		JWT                 string `json:"jwt"`
		DeviceCollectionURL string `json:"deviceCollectionURL"`
	}

	Browser struct {
		BrowserInfo   BrowserInfo       `json:"browserInfo"`
		DeviceChannel DeviceChannelType `json:"deviceChannel"`
	}

	BrowserInfoFields struct {
		JavascriptEnabled bool
		UserAgent         string
		Header            string
		JavaEnabled       bool
		Language          string
		ColorDepth        ColorDepth
		ScreenHeight      string
		ScreenWidth       string
		TimeZone          string
		IpAddress         string
	}

	Lookup3DSArgs struct {
		ID3DS                   string `json:"3dsID"`
		AuthenticationIndicator string `json:"authenticationIndicator"`
		TransactionMode         string `json:"transactionMode"`
		TransactionType         string `json:"transactionType"`
		ProductCode             string `json:"productCode"`
		Account                 Account
		Order                   Order
	}

	Lookup3DSResponse struct {
		SC                     int    `json:"SC"`
		EC                     string `json:"EC"`
		EM                     string `json:"EM"`
		Version3DS             string `json:"3dsVersion"`
		Enrolled               string
		ProcessorTransactionID string `json:"processorTransactionID"`
		DsTransactionID        string `json:"dsTransactionID"`
		Status                 string
		ECI                    string `json:"ECI"`
		UCAF                   string `json:"UCAF"`
		XID                    string `json:"XID"`
		ChallengeURL           string `json:"challengeURL"`
		Payload                string
	}

	Authenticate3DSArgs struct {
		ID3DS string `json:"3dsID"`
		JWT   string `json:"jwt"`
	}

	Authenticate3DSResponse struct {
		SC                     int    `json:"SC"`
		EC                     string `json:"EC"`
		EM                     string `json:"EM"`
		Version3DS             string `json:"3dsVersion"`
		Enrolled               string
		ProcessorTransactionID string `json:"processorTransactionID"`
		DsTransactionID        string `json:"dsTransactionID"`
		Status                 string
		ECI                    string `json:"ECI"`
		UCAF                   string `json:"UCAF"`
		XID                    string `json:"XID"`
	}
)
