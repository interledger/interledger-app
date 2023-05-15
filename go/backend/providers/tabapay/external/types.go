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
	CreateTransactionPullOptions struct {
		ThreeDS      ThreeDS `json:"3dSecure"`
		SecurityCode string  `json:"securityCode,omitempty"`
	}

	ThreeDS struct {
		Version         string `json:"version"`
		ECI             string `json:"ECI"`
		UCAF            string `json:"UCAF"`
		XID             string `json:"XID"`
		DSTransactionID string `json:"dsTransactionID"`
	}

	CreateTransactionArgs struct {
		ReferenceID string                        `json:"referenceID"`
		Type        TransactionType               `json:"type,omitempty"`
		Accounts    CreateTransactionAccounts     `json:"accounts"`
		Currency    string                        `json:"currency,omitempty"`
		Amount      string                        `json:"amount"`
		PullOptions *CreateTransactionPullOptions `json:"pullOptions,omitempty"`
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
		Name    Name     `json:"name"`
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
		SC            int           `json:"SC"`
		EC            string        `json:"EC"`
		TransactionID string        `json:"transactionID"`
		Network       string        `json:"network"`
		NetworkRC     string        `json:"networkRC"`
		Status        string        `json:"status"`
		ApprovalCode  string        `json:"approvalCode"`
		Errors        []string      `json:"errors"`
		Fees          *Fees         `json:"fees,omitempty"`
		Card          *CardResponse `json:"card,omitempty"`
	}

	Fees struct {
		Interchange string `json:"interchange"`
		Network     string `json:"network"`
		Tabapay     string `json:"tabapay"`
	}

	CardResponse struct {
		NameFI         string     `json:"nameFI,omitempty"`
		Last4          string     `json:"last4"`
		ExpirationDate string     `json:"expirationDate,omitempty"`
		Push           PushObject `json:"push,omitempty"`
		Pull           PullObject `json:"pull,omitempty"`
	}

	RetrieveTransactionResponse struct {
		SC             int       `json:"SC"`
		EC             string    `json:"EC"`
		ReferenceID    string    `json:"referenceID"`
		Network        string    `json:"network"`
		NetworkRC      string    `json:"networkRC"`
		Status         string    `json:"status"`
		Originally     string    `json:"originally"`
		Amount         string    `json:"amount"`
		AmountUSD      string    `json:"amountUSD"`
		Fees           *Fees     `json:"fees,omitempty"`
		ReversalStatus string    `json:"reversalStatus"`
		Reversal       *Reversal `json:"reversal,omitempty"`
	}

	Reversal struct {
		NetworkRC  string `json:"networkRC"`
		NetworkRC2 string `json:"networkRC2"`
		Error      string `json:"error"`
	}

	RetrieveAccountResponse struct {
		SC          int           `json:"SC"`
		EC          string        `json:"EC"`
		ReferenceID string        `json:"referenceID"`
		Bank        *Bank         `json:"bank,omitempty"`
		Card        *CardResponse `json:"card,omitempty"`
		Owner       Owner         `json:"owner"`
	}

	CreateAccountArgs struct {
		RejectDuplicateCard  bool   `json:"-"`
		OKToAddDuplicateCard bool   `json:"-"`
		ReferenceID          string `json:"referenceID"`
		Card                 Card   `json:"card"`
		Owner                Owner  `json:"owner"`
	}

	CreateAccountResponse struct {
		SC          int           `json:"SC"`
		EC          string        `json:"EC"`
		AccountID   string        `json:"accountID"`
		ReferenceID string        `json:"-"`
		Card        *CardResponse `json:"card,omitempty"`
		Notices     string
	}

	QueryCardArgs struct {
		Card  *Card  `json:"card,omitempty"`
		Owner *Owner `json:"owner,omitempty"`
	}

	PullObject struct {
		Enabled   bool     `json:"enabled"`
		Network   string   `json:"network"`
		Type      CardType `json:"type"`
		Regulated bool     `json:"regulated"`
		Currency  string   `json:"currency"`
		Country   string   `json:"country"`
	}

	PushObject struct {
		Enabled      bool     `json:"enabled"`
		Network      string   `json:"network"`
		Type         CardType `json:"type"`
		Regulated    bool     `json:"regulated"`
		Currency     string   `json:"currency"`
		Country      string   `json:"country"`
		Availability string   `json:"availability"`
	}

	AVSResponse struct {
		ID               string `json:"avsID"`
		NetworkRC        string `json:"networkRC"`
		NetworkID        string `json:"networkID"`
		AuthorizeID      string `json:"authorizeID"`
		ResultText       string `json:"resultText"`
		CodeAVS          string `json:"codeAVS"`
		CodeSecurityCode string `json:"codeSecurityCode"`
		EC               string `json:"EC"`
	}

	QueryCardResponse struct {
		SC   int          `json:"SC"`
		EC   string       `json:"EC"`
		EM   string       `jsonL:"EM"`
		Card CardResponse `json:"card,omitempty"`
		AVS  AVSResponse  `json:"AVS"`
	}

	Account struct {
		AccountID string `json:"accountID"`
		Owner     *Owner `json:"owner,omitempty"`
	}

	Order struct {
		OrderID  string `json:"orderID"`
		Currency string `json:"currency"`
		Amount   string `json:"amount"`
	}

	Init3DSArgs struct {
		Account Account `json:"account"`
		Order   Order   `json:"order"`
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
		ID3DS                   string  `json:"3dsID"`
		AuthenticationIndicator string  `json:"authenticationIndicator"`
		TransactionMode         string  `json:"transactionMode"`
		TransactionType         string  `json:"transactionType"`
		ProductCode             string  `json:"productCode"`
		Account                 Account `json:"account"`
		Order                   Order   `json:"order"`
		Browser                 Browser `json:"browser"`
	}

	Lookup3DSResponse struct {
		SC                     int    `json:"SC"`
		EC                     string `json:"EC"`
		EM                     string `json:"EM"`
		Version3DS             string `json:"3dsVersion"`
		Enrolled               string `json:"enrolled"`
		ProcessorTransactionID string `json:"processorTransactionID"`
		DsTransactionID        string `json:"dsTransactionID"`
		Status                 string `json:"status"`
		ECI                    string `json:"ECI"`
		UCAF                   string `json:"UCAF"`
		XID                    string `json:"XID"`
		ChallengeURL           string `json:"challengeURL"`
		Payload                string `json:"payload"`
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
		Enrolled               string `json:"enrolled"`
		ProcessorTransactionID string `json:"processorTransactionID"`
		DsTransactionID        string `json:"dsTransactionID"`
		Status                 string `json:"status"`
		ECI                    string `json:"ECI"`
		UCAF                   string `json:"UCAF"`
		XID                    string `json:"XID"`
	}
)

func (t *RetrieveTransactionResponse) GetFees() *Fees {
	if t.Fees != nil {
		return t.Fees
	}

	return &Fees{}
}

func (t *RetrieveTransactionResponse) GetReversal() *Reversal {
	if t.Reversal != nil {
		return t.Reversal
	}

	return &Reversal{}
}

func (t *CreateTransactionResponse) GetFees() *Fees {
	if t.Fees != nil {
		return t.Fees
	}

	return &Fees{}
}

func (t *CreateTransactionResponse) GetCard() *CardResponse {
	if t.Card != nil {
		return t.Card
	}

	return &CardResponse{}
}
