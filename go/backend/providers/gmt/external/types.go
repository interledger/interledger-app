package external

import (
	"encoding/xml"
	"time"
)

var RedactFields = []string{
	"SenderAchAccount",
	"SenderAchRouting",
	"SenderCardExpiration",
	"SenderCardNumber",
	"SenderEmail",
	"SenderIP",
	"SenderIdCard",
	"SenderIdNumber",
	"SenderIdNumber2",
	"SenderPhone",
	"ReceiverAchAccount",
	"ReceiverAchRouting",
	"ReceiverCardExpiration",
	"ReceiverCardNumber",
	"ReceiverEmail",
	"ReceiverIP",
	"ReceiverIdCard",
	"ReceiverIdNumber",
	"ReceiverIdNumber2",
	"ReceiverPhone",
	"BancosNombre",
	"BankAccount",
	"BeneficiarioCelular",
	"CardNumber",
	"CardNumber_CVV",
	"CardNumber_ExpDate",
	"ComplianceIdNumber",
	"ComplianceIdExpirationDate",
	"ComplianceSSN",
	"SucursalBanco",
	"pass",
}

type GMTDate time.Time

/*Created, Hold, Authorized, Transmitted, Paid, Cancelled or Void and Expired*/
const (
	TransactionStatusCreated    = "Created"
	TransactionStatusPaid       = "Paid"
	TransactionStatusHold       = "Hold"
	TransactionStatusAuthorized = "Authorized"
	TransactionStatusCancelled  = "Cancelled"
	TransactionStatusExpired    = "Expired"
)

func (gd GMTDate) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	t := time.Time(gd)
	if t.IsZero() {
		return nil
	}

	v := t.Format("2006-01-02")
	return e.EncodeElement(v, start)
}

func (gd *GMTDate) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	err := d.DecodeElement(&s, &start)
	if err != nil {
		return err
	}
	t, err := time.Parse("2006-01-02T15:04:05", s)
	if err != nil {
		return err
	}
	*gd = GMTDate(t)
	return nil
}

type SOAPEnvelope struct {
	XMLName xml.Name `xml:"soap:Envelope"`
	XmlNS   string   `xml:"xmlns:soap,attr"`
	Body    SOAPBody
}

type SOAPBody struct {
	XMLName xml.Name `xml:"soap:Body"`

	Content interface{} `xml:",omitempty"`
}

type SOAPEnvelopeResponse struct {
	XMLName xml.Name    `xml:"Envelope"`
	Text    string      `xml:",chardata"`
	S       string      `xml:"s,attr"`
	Body    interface{} `xml:"Body"`
}

type InsertTransaction struct {
	XMLName  xml.Name        `xml:"http://tempuri.org/ InsertTransaction"`
	Alias    string          `xml:"alias,omitempty"`
	User     string          `xml:"user,omitempty"`
	Pass     string          `xml:"pass,omitempty"`
	Sender   *WsSender       `json:"sender,omitempty"`
	Receiver *WsReceiver     `json:"receiver,omitempty"`
	Transfer *WsTransferInfo `json:"transfer,omitempty"`
}

type InsertTransactionResponse struct {
	Text                    string      `xml:",chardata"`
	Xmlns                   string      `xml:"xmlns,attr"`
	InsertTransactionResult *WsResponse `xml:"InsertTransactionResult,omitempty"`
}

type UpdateTransactionStatus struct {
	XMLName    xml.Name `xml:"http://tempuri.org/ UpdateTransactionStatus"`
	Partner    string   `xml:"partner,omitempty"`
	Reference  string   `xml:"reference,omitempty"`
	Statuscode string   `xml:"statuscode,omitempty"`
	Date       GMTDate  `xml:"date,omitempty"`
	User       string   `xml:"user,omitempty"`
	Pass       string   `xml:"pass,omitempty"`
	Comment    string   `xml:"comment,omitempty"`
}

type UpdateTransactionStatusResponse struct {
	XMLName                       xml.Name `xml:"http://tempuri.org/ UpdateTransactionStatusResponse"`
	UpdateTransactionStatusResult *WsResponse
}

type GetPaidTransactions struct {
	XMLName xml.Name `xml:"http://tempuri.org/ GetPaidTransactions"`
	Alias   string   `xml:"alias,omitempty"`
	User    string   `xml:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty"`
}

type GetPaidTransactionsResponse struct {
	XMLName                   xml.Name                   `xml:"http://tempuri.org/ GetPaidTransactionsResponse"`
	GetPaidTransactionsResult *ArrayOfwsPaidTransactions `xml:"GetPaidTransactionsResult,omitempty"`
}

type ConfirmPayment struct {
	XMLName xml.Name `xml:"http://tempuri.org/ ConfirmPayment"`
	Alias   string   `xml:"alias,omitempty"`
	User    string   `xml:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty"`
	Receipt string   `xml:"receipt,omitempty"`
}

type ConfirmPaymentResponse struct {
	XMLName              xml.Name    `xml:"http://tempuri.org/ ConfirmPaymentResponse"`
	ConfirmPaymentResult *WsResponse `json:"ConfirmPaymentResult,omitempty"`
}

type RequestCancellation struct {
	XMLName xml.Name `xml:"http://tempuri.org/ RequestCancellation"`
	Alias   string   `xml:"alias,omitempty"`
	User    string   `xml:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty"`
	Receipt string   `xml:"receipt,omitempty"`
	Comment string   `xml:"comment,omitempty"`
}

type RequestCancellationResponse struct {
	XMLName                   xml.Name    `xml:"http://tempuri.org/ RequestCancellationResponse"`
	RequestCancellationResult *WsResponse `json:"RequestCancellationResult,omitempty"`
}

type GetCancelledTransactions struct {
	XMLName xml.Name `xml:"http://tempuri.org/ GetCancelledTransactions"`
	Alias   string   `xml:"alias,omitempty"`
	User    string   `xml:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty"`
}

type GetCancelledTransactionsResponse struct {
	XMLName                        xml.Name                   `xml:"http://tempuri.org/ GetCancelledTransactionsResponse"`
	GetCancelledTransactionsResult *ArrayOfwsVoidTransactions `xml:"GetCancelledTransactionsResult,omitempty"`
}

type ConfirmCancellation struct {
	XMLName xml.Name `xml:"http://tempuri.org/ ConfirmCancellation"`
	Alias   string   `xml:"alias,omitempty"`
	User    string   `xml:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty"`
	Receipt string   `xml:"receipt,omitempty"`
}

type ConfirmCancellationResponse struct {
	XMLName                   xml.Name    `xml:"http://tempuri.org/ ConfirmCancellationResponse"`
	ConfirmCancellationResult *WsResponse `json:"ConfirmCancellationResult,omitempty"`
}

type RequestModification struct {
	XMLName xml.Name             `xml:"http://tempuri.org/ RequestModification"`
	Alias   string               `xml:"alias,omitempty"`
	User    string               `xml:"user,omitempty"`
	Pass    string               `xml:"pass,omitempty"`
	Receipt string               `xml:"receipt,omitempty"`
	Comment string               `xml:"comment,omitempty"`
	Data    *WsChangeRequestData `json:"data,omitempty"`
}

type RequestModificationResponse struct {
	XMLName                   xml.Name    `xml:"http://tempuri.org/ RequestModificationResponse"`
	RequestModificationResult *WsResponse `json:"RequestModificationResult,omitempty"`
}

type GetModifiedTransactions struct {
	XMLName xml.Name `xml:"http://tempuri.org/ GetModifiedTransactions"`
	Alias   string   `xml:"alias,omitempty"`
	User    string   `xml:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty"`
}

type GetModifiedTransactionsResponse struct {
	XMLName                       xml.Name                       `xml:"http://tempuri.org/ GetModifiedTransactionsResponse"`
	GetModifiedTransactionsResult *ArrayOfwsModifiedTransactions `xml:"GetModifiedTransactionsResult,omitempty"`
}

type ConfirmModification struct {
	XMLName xml.Name `xml:"http://tempuri.org/ ConfirmModification"`
	Alias   string   `xml:"alias,omitempty"`
	User    string   `xml:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty"`
	Receipt string   `xml:"receipt,omitempty"`
}

type ConfirmModificationResponse struct {
	XMLName                   xml.Name    `xml:"http://tempuri.org/ ConfirmModificationResponse"`
	ConfirmModificationResult *WsResponse `json:"ConfirmModificationResult,omitempty"`
}

type GetReleasedTransactions struct {
	XMLName xml.Name `xml:"http://tempuri.org/ GetReleasedTransactions"`
	Alias   string   `xml:"alias,omitempty"`
	User    string   `xml:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty"`
}

type GetReleasedTransactionsResponse struct {
	XMLName                       xml.Name                       `xml:"http://tempuri.org/ GetReleasedTransactionsResponse"`
	GetReleasedTransactionsResult *ArrayOfwsReleasedTransactions `xml:"GetReleasedTransactionsResult,omitempty"`
}

type ConfirmRelease struct {
	XMLName xml.Name `xml:"http://tempuri.org/ ConfirmRelease"`
	Alias   string   `xml:"alias,omitempty"`
	User    string   `xml:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty"`
	Receipt string   `xml:"receipt,omitempty"`
}

type ConfirmReleaseResponse struct {
	XMLName              xml.Name    `xml:"http://tempuri.org/ ConfirmReleaseResponse"`
	ConfirmReleaseResult *WsResponse `json:"ConfirmReleaseResult,omitempty"`
}

type GetClearedAchTransactions struct {
	XMLName xml.Name `xml:"http://tempuri.org/ GetClearedAchTransactions"`
	Alias   string   `xml:"alias,omitempty"`
	User    string   `xml:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty"`
}

type GetClearedAchTransactionsResponse struct {
	XMLName                         xml.Name               `xml:"http://tempuri.org/ GetClearedAchTransactionsResponse"`
	GetClearedAchTransactionsResult *ArrayOfwsCollectedAch `xml:"GetClearedAchTransactionsResult,omitempty"`
}

type ConfirmCollection struct {
	XMLName xml.Name `xml:"http://tempuri.org/ ConfirmCollection"`
	Alias   string   `xml:"alias,omitempty"`
	User    string   `xml:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty"`
	Receipt string   `xml:"receipt,omitempty"`
}

type ConfirmCollectionResponse struct {
	Text                    string      `xml:",chardata"`
	Xmlns                   string      `xml:"xmlns,attr"`
	ConfirmCollectionResult *WsResponse `xml:"ConfirmCollectionResult,omitempty"`
}

type GetNotifications struct {
	XMLName xml.Name `xml:"http://tempuri.org/ GetNotifications"`
	Alias   string   `xml:"alias,omitempty"`
	User    string   `xml:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty"`
}

type GetNotificationsResponse struct {
	Text                   string                  `xml:",chardata"`
	Xmlns                  string                  `xml:"xmlns,attr"`
	GetNotificationsResult *ArrayOfwsNotifications `xml:"GetNotificationsResult"`
}

type GetAchStatus struct {
	XMLName xml.Name `xml:"http://tempuri.org/ GetAchStatus"`
	Alias   string   `xml:"alias,omitempty"`
	User    string   `xml:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty"`
	Receipt string   `xml:"receipt,omitempty"`
}

type GetAchStatusResponse struct {
	XMLName            xml.Name     `xml:"http://tempuri.org/ GetAchStatusResponse"`
	GetAchStatusResult *WsAchStatus `json:"GetAchStatusResult,omitempty"`
}

type GetTransactionStatus struct {
	XMLName xml.Name `xml:"http://tempuri.org/ GetTransactionStatus"`
	Alias   string   `xml:"alias,omitempty"`
	User    string   `xml:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty"`
	Receipt string   `xml:"receipt,omitempty"`
}

type GetTransactionStatusResponse struct {
	XMLName                    xml.Name  `xml:"http://tempuri.org/ GetTransactionStatusResponse"`
	GetTransactionStatusResult *WsStatus `json:"GetTransactionStatusResult,omitempty"`
}

type GetSingleExchangeRate struct {
	XMLName  xml.Name `xml:"http://tempuri.org/ GetSingleExchangeRate"`
	Alias    string   `xml:"alias,omitempty"`
	User     string   `xml:"user,omitempty"`
	Pass     string   `xml:"pass,omitempty"`
	Payer    string   `xml:"payer,omitempty"`
	Mts      string   `xml:"mts,omitempty"`
	Service  string   `xml:"service,omitempty"`
	Currency string   `xml:"currency,omitempty"`
}

type GetSingleExchangeRateResponse struct {
	XMLName                     xml.Name  `xml:"http://tempuri.org/ GetSingleExchangeRateResponse"`
	GetSingleExchangeRateResult *WsExRate `json:"GetSingleExchangeRateResult,omitempty"`
}

type GetExchangeRates struct {
	XMLName  xml.Name `xml:"http://tempuri.org/ GetExchangeRates"`
	Alias    string   `xml:"alias,omitempty"`
	User     string   `xml:"user,omitempty"`
	Pass     string   `xml:"pass,omitempty"`
	Currency string   `xml:"currency,omitempty"`
}

type GetExchangeRatesResponse struct {
	XMLName                xml.Name             `xml:"http://tempuri.org/ GetExchangeRatesResponse"`
	GetExchangeRatesResult *ArrayOfwsExRateList `xml:"GetExchangeRatesResult,omitempty"`
}

type OfacVerification struct {
	XMLName   xml.Name `xml:"http://tempuri.org/ OfacVerification"`
	Alias     string   `xml:"alias,omitempty"`
	User      string   `xml:"user,omitempty"`
	Pass      string   `xml:"pass,omitempty"`
	LastName  string   `xml:"lastName,omitempty"`
	FirstName string   `xml:"firstName,omitempty"`
}

type OfacVerificationResponse struct {
	Text                   string  `xml:",chardata"`
	Xmlns                  string  `xml:"xmlns,attr"`
	OfacVerificationResult *WsOfac `xml:"OfacVerificationResult"`
}

type ComplianceCheck struct {
	XMLName  xml.Name        `xml:"http://tempuri.org/ ComplianceCheck"`
	Alias    string          `xml:"alias,omitempty"`
	User     string          `xml:"user,omitempty"`
	Pass     string          `xml:"pass,omitempty"`
	Sender   *WsSender       `json:"sender,omitempty"`
	Receiver *WsReceiver     `json:"receiver,omitempty"`
	Transfer *WsTransferInfo `json:"transfer,omitempty"`
}

type ComplianceCheckResponse struct {
	Text                  string      `xml:",chardata"`
	Xmlns                 string      `xml:"xmlns,attr"`
	ComplianceCheckResult *WsResponse `json:"ComplianceCheckResult,omitempty"`
}

type SetVerified struct {
	XMLName xml.Name `xml:"http://tempuri.org/ SetVerified"`
	Alias   string   `xml:"alias,omitempty"`
	User    string   `xml:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty"`
	Receipt string   `xml:"receipt,omitempty"`
	Passed  bool     `xml:"passed,omitempty"`
	Score   string   `xml:"score,omitempty"`
	Comment string   `xml:"comment,omitempty"`
}

type SetVerifiedResponse struct {
	Text              string    `xml:",chardata"`
	Xmlns             string    `xml:"xmlns,attr"`
	SetVerifiedResult *WsResult `xml:"SetVerifiedResult,omitempty"`
}

type PayersConsult struct {
	XMLName     xml.Name `xml:"http://tempuri.org/ PayersConsult"`
	Alias       string   `xml:"alias,omitempty"`
	User        string   `xml:"user,omitempty"`
	Pass        string   `xml:"pass,omitempty"`
	CountryCode string   `xml:"countryCode,omitempty"`
	PayerCode   string   `xml:"payerCode,omitempty"`
}

type PayersConsultResponse struct {
	XMLName             xml.Name                                    `xml:"http://tempuri.org/ PayersConsultResponse"`
	PayersConsultResult *ArrayOfws_Select_PayersByCountryCodeResult `xml:"PayersConsultResult,omitempty"`
}

type GetReceiptData struct {
	XMLName xml.Name `xml:"http://tempuri.org/ GetReceiptData"`
	Alias   string   `xml:"alias,omitempty"`
	User    string   `xml:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty"`
	Rem_edo string   `xml:"rem_edo,omitempty"`
	Pais    string   `xml:"pais,omitempty"`
}

type GetReceiptDataResponse struct {
	XMLName              xml.Name         `xml:"http://tempuri.org/ GetReceiptDataResponse"`
	GetReceiptDataResult *ReceiptTemplate `xml:"GetReceiptDataResult,omitempty"`
}

type GetCityByZip struct {
	XMLName xml.Name `xml:"http://tempuri.org/ GetCityByZip"`
	Alias   string   `xml:"alias,omitempty"`
	User    string   `xml:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty"`
	Zip     string   `xml:"zip,omitempty"`
}

type GetCityByZipResponse struct {
	XMLName            xml.Name   `xml:"http://tempuri.org/ GetCityByZipResponse"`
	GetCityByZipResult *CityByZip `xml:"GetCityByZipResult,omitempty"`
}

type CheckPayerLimits struct {
	XMLName     xml.Name `xml:"http://tempuri.org/ CheckPayerLimits"`
	Alias       string   `xml:"alias,omitempty"`
	User        string   `xml:"user,omitempty"`
	Pass        string   `xml:"pass,omitempty"`
	Payercode   string   `xml:"payercode,omitempty"`
	Send        float64  `xml:"send,omitempty"`
	Receive     float64  `xml:"receive,omitempty"`
	Servicecode string   `xml:"servicecode,omitempty"`
	Ben         string   `xml:"ben,omitempty"`
	Rem         string   `xml:"rem,omitempty"`
}

type CheckPayerLimitsResponse struct {
	XMLName                xml.Name               `xml:"http://tempuri.org/ CheckPayerLimitsResponse"`
	CheckPayerLimitsResult *Ws_LimitByPayerResult `json:"CheckPayerLimitsResult,omitempty"`
}

type PromotionsCode struct {
	XMLName    xml.Name `xml:"http://tempuri.org/ PromotionsCode"`
	Code_promo string   `xml:"Code_promo,omitempty"`
	Fee        float32  `xml:"Fee,omitempty"`
}

type PromotionsCodeResponse struct {
	XMLName              xml.Name               `xml:"http://tempuri.org/ PromotionsCodeResponse"`
	PromotionsCodeResult *Ws_Select_PromoResult `json:"PromotionsCodeResult,omitempty"`
}

type RegisterWallet struct {
	XMLName     xml.Name  `xml:"http://tempuri.org/ RegisterWallet"`
	Alias       string    `xml:"alias,omitempty"`
	User        string    `xml:"user,omitempty"`
	Pass        string    `xml:"pass,omitempty"`
	Sender      *WsSender `json:"sender,omitempty"`
	External_id string    `xml:"external_id,omitempty"`
}

type RegisterWalletResponse struct {
	XMLName              xml.Name  `xml:"http://tempuri.org/ RegisterWalletResponse"`
	RegisterWalletResult *WsWallet `json:"RegisterWalletResult,omitempty"`
}

type AddWalletFunds struct {
	XMLName        xml.Name          `xml:"http://tempuri.org/ AddWalletFunds"`
	Alias          string            `xml:"alias,omitempty"`
	User           string            `xml:"user,omitempty"`
	Pass           string            `xml:"pass,omitempty"`
	Operation_Info *Wallet_Operation `xml:"Operation_Info,omitempty"`
}

type AddWalletFundsResponse struct {
	XMLName              xml.Name  `xml:"http://tempuri.org/ AddWalletFundsResponse"`
	AddWalletFundsResult *WsWallet `json:"AddWalletFundsResult,omitempty"`
}

type WithdrawWalletFunds struct {
	XMLName        xml.Name          `xml:"http://tempuri.org/ WithdrawWalletFunds"`
	Alias          string            `xml:"alias,omitempty"`
	User           string            `xml:"user,omitempty"`
	Pass           string            `xml:"pass,omitempty"`
	Operation_Info *Wallet_Operation `xml:"Operation_Info,omitempty"`
}

type WithdrawWalletFundsResponse struct {
	XMLName                   xml.Name  `xml:"http://tempuri.org/ WithdrawWalletFundsResponse"`
	WithdrawWalletFundsResult *WsWallet `json:"WithdrawWalletFundsResult,omitempty"`
}

type GetWalletBalance struct {
	XMLName    xml.Name `xml:"http://tempuri.org/ GetWalletBalance"`
	Alias      string   `xml:"alias,omitempty"`
	User       string   `xml:"user,omitempty"`
	Pass       string   `xml:"pass,omitempty"`
	SenderId   int32    `xml:"senderId,omitempty"`
	ExternalId string   `xml:"ExternalId,omitempty"`
}

type GetWalletBalanceResponse struct {
	XMLName                xml.Name  `xml:"http://tempuri.org/ GetWalletBalanceResponse"`
	GetWalletBalanceResult *WsWallet `json:"GetWalletBalanceResult,omitempty"`
}

type AddDocument struct {
	XMLName  xml.Name    `xml:"http://tempuri.org/ AddDocument"`
	Alias    string      `xml:"alias,omitempty"`
	User     string      `xml:"user,omitempty"`
	Pass     string      `xml:"pass,omitempty"`
	Document *WsDocument `json:"Document,omitempty"`
}

type AddDocumentResponse struct {
	XMLName           xml.Name    `xml:"http://tempuri.org/ AddDocumentResponse"`
	AddDocumentResult *WsResponse `json:"AddDocumentResult,omitempty"`
}

type GetOccupations struct {
	XMLName xml.Name `xml:"http://tempuri.org/ GetOccupations"`
	Alias   string   `xml:"alias,omitempty"`
	User    string   `xml:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty"`
}

type GetOccupationsResponse struct {
	XMLName              xml.Name              `xml:"http://tempuri.org/ GetOccupationsResponse"`
	GetOccupationsResult *ArrayOfwsOccupations `xml:"GetOccupationsResult,omitempty"`
}

type WsSender struct {
	XMLName                     xml.Name `xml:"sender"`
	XmlNS                       string   `xml:"xmlns:a,attr"`
	AgentAlias                  string   `xml:"a:AgentAlias,omitempty"`
	BusinessName                string   `xml:"a:BusinessName,omitempty"`
	BusinessWebsite             string   `xml:"a:BusinessWebsite,omitempty"`
	Debit                       bool     `xml:"a:Debit,omitempty"`
	RepresentativeID            string   `xml:"a:RepresentativeID,omitempty"`
	SenderAchAccount            string   `xml:"a:SenderAchAccount,omitempty"`
	SenderAchRouting            string   `xml:"a:SenderAchRouting,omitempty"`
	SenderAchType               string   `xml:"a:SenderAchType,omitempty"`
	SenderActive                bool     `xml:"a:SenderActive,omitempty"`
	SenderAddress               string   `xml:"a:SenderAddress,omitempty"`
	SenderAddressStreet         string   `xml:"a:SenderAddressStreet,omitempty"`
	SenderAggregatedTransfers   int32    `xml:"a:SenderAggregatedTransfers,omitempty"`
	SenderBank                  string   `xml:"a:SenderBank,omitempty"`
	SenderBirthDate             GMTDate  `xml:"a:SenderBirthDate,omitempty"`
	SenderCardBank              string   `xml:"a:SenderCardBank,omitempty"`
	SenderCardExpiration        GMTDate  `xml:"a:SenderCardExpiration,omitempty"`
	SenderCardName              string   `xml:"a:SenderCardName,omitempty"`
	SenderCardNumber            string   `xml:"a:SenderCardNumber,omitempty"`
	SenderCardType              int      `xml:"a:SenderCardType,omitempty"`
	SenderCity                  string   `xml:"a:SenderCity,omitempty"`
	SenderComments              string   `xml:"a:SenderComments,omitempty"`
	SenderComments3             string   `xml:"a:SenderComments3,omitempty"`
	SenderCompany               string   `xml:"a:SenderCompany,omitempty"`
	SenderCompanyAddress        string   `xml:"a:SenderCompanyAddress,omitempty"`
	SenderCompanyPhone          string   `xml:"a:SenderCompanyPhone,omitempty"`
	SenderCountryCode           string   `xml:"a:SenderCountryCode,omitempty"`
	SenderCountryNationallity   string   `xml:"a:SenderCountryNationallity,omitempty"`
	SenderCountryResidence      string   `xml:"a:SenderCountryResidence,omitempty"`
	SenderCurrencyCode          string   `xml:"a:SenderCurrencyCode,omitempty"`
	SenderDocExpiration         GMTDate  `xml:"a:SenderDocExpiration,omitempty"`
	SenderEmail                 string   `xml:"a:SenderEmail,omitempty"`
	SenderFileImg               string   `xml:"a:SenderFileImg,omitempty"`
	SenderFileImg2              string   `xml:"a:SenderFileImg2,omitempty"`
	SenderForceNew              bool     `xml:"a:SenderForceNew,omitempty"`
	SenderGender                string   `xml:"a:SenderGender,omitempty"`
	SenderIP                    string   `xml:"a:SenderIP,omitempty"`
	SenderId                    int32    `xml:"a:SenderId,omitempty"`
	SenderIdCard                string   `xml:"a:SenderIdCard,omitempty"`
	SenderIdIssuer              string   `xml:"a:SenderIdIssuer,omitempty"`
	SenderIdNumber              string   `xml:"a:SenderIdNumber,omitempty"`
	SenderIdNumber2             string   `xml:"a:SenderIdNumber2,omitempty"`
	SenderIdType                string   `xml:"a:SenderIdType,omitempty"`
	SenderIdType2               string   `xml:"a:SenderIdType2,omitempty"`
	SenderIsBusiness            bool     `xml:"a:SenderIsBusiness,omitempty"`
	SenderLastName              string   `xml:"a:SenderLastName,omitempty"`
	SenderMaritalStatus         string   `xml:"a:SenderMaritalStatus,omitempty"`
	SenderMobile                string   `xml:"a:SenderMobile,omitempty"`
	SenderMoneyOrigin           string   `xml:"a:SenderMoneyOrigin,omitempty"`
	SenderMoneyOwn              bool     `xml:"a:SenderMoneyOwn,omitempty"`
	SenderMonthAverage          float64  `xml:"a:SenderMonthAverage,omitempty"`
	SenderName                  string   `xml:"a:SenderName,omitempty"`
	SenderNatureOfBusiness      string   `xml:"a:SenderNatureOfBusiness,omitempty"`
	SenderOccupation            string   `xml:"a:SenderOccupation,omitempty"`
	SenderOnBehalfOf            bool     `xml:"a:SenderOnBehalfOf,omitempty"`
	SenderPEP                   bool     `xml:"a:SenderPEP,omitempty"`
	SenderPOB                   string   `xml:"a:SenderPOB,omitempty"`
	SenderPassword              string   `xml:"a:SenderPassword,omitempty"`
	SenderPhone                 string   `xml:"a:SenderPhone,omitempty"`
	SenderPicture               string   `xml:"a:SenderPicture,omitempty"`
	SenderPoliticalFamily       bool     `xml:"a:SenderPoliticalFamily,omitempty"`
	SenderResidenceAddress      string   `xml:"a:SenderResidenceAddress,omitempty"`
	SenderResidenceAddressExtra string   `xml:"a:SenderResidenceAddressExtra,omitempty"`
	SenderResidenceCity         string   `xml:"a:SenderResidenceCity,omitempty"`
	SenderResidenceCountryCode  string   `xml:"a:SenderResidenceCountryCode,omitempty"`
	SenderResidenceState        string   `xml:"a:SenderResidenceState,omitempty"`
	SenderResidenceZip          string   `xml:"a:SenderResidenceZip,omitempty"`
	SenderSendingReason         string   `xml:"a:SenderSendingReason,omitempty"`
	SenderState                 string   `xml:"a:SenderState,omitempty"`
	SenderTrackingNumber        string   `xml:"a:SenderTrackingNumber,omitempty"`
	SenderZip                   string   `xml:"a:SenderZip,omitempty"`
}

type WsReceiver struct {
	XMLName                     xml.Name `xml:"receiver"`
	XmlNS                       string   `xml:"xmlns:a,attr"`
	ReceiverAchAccount          string   `xml:"a:ReceiverAchAccount,omitempty"`
	ReceiverAchRouting          string   `xml:"a:ReceiverAchRouting,omitempty"`
	ReceiverAchType             string   `xml:"a:ReceiverAchType,omitempty"`
	ReceiverActive              bool     `xml:"a:ReceiverActive,omitempty"`
	ReceiverAddress             string   `xml:"a:ReceiverAddress,omitempty"`
	ReceiverAverageMonth        float64  `xml:"a:ReceiverAverageMonth,omitempty"`
	ReceiverBirthDate           GMTDate  `xml:"a:ReceiverBirthDate,omitempty"`
	ReceiverCity                string   `xml:"a:ReceiverCity,omitempty"`
	ReceiverCompany             string   `xml:"a:ReceiverCompany,omitempty"`
	ReceiverCountry             string   `xml:"a:ReceiverCountry,omitempty"`
	ReceiverCountryNationallity string   `xml:"a:ReceiverCountryNationallity,omitempty"`
	ReceiverCurrency            string   `xml:"a:ReceiverCurrency,omitempty"`
	ReceiverDocExpiration       GMTDate  `xml:"a:ReceiverDocExpiration,omitempty"`
	ReceiverEmail               string   `xml:"a:ReceiverEmail,omitempty"`
	ReceiverFileImg             string   `xml:"a:ReceiverFileImg,omitempty"`
	ReceiverFileImg2            string   `xml:"a:ReceiverFileImg2,omitempty"`
	ReceiverGender              string   `xml:"a:ReceiverGender,omitempty"`
	ReceiverId                  int32    `xml:"a:ReceiverId,omitempty"`
	ReceiverIdIssuer            string   `xml:"a:ReceiverIdIssuer,omitempty"`
	ReceiverIdNumber            string   `xml:"a:ReceiverIdNumber,omitempty"`
	ReceiverIdType              string   `xml:"a:ReceiverIdType,omitempty"`
	ReceiverLastName            string   `xml:"a:ReceiverLastName,omitempty"`
	ReceiverLastTransaction     int32    `xml:"a:ReceiverLastTransaction,omitempty"`
	ReceiverMaritalStatus       string   `xml:"a:ReceiverMaritalStatus,omitempty"`
	ReceiverMobile              string   `xml:"a:ReceiverMobile,omitempty"`
	ReceiverMoneyOrigin         string   `xml:"a:ReceiverMoneyOrigin,omitempty"`
	ReceiverName                string   `xml:"a:ReceiverName,omitempty"`
	ReceiverOccupation          string   `xml:"a:ReceiverOccupation,omitempty"`
	ReceiverOfficeCode          string   `xml:"a:ReceiverOfficeCode,omitempty"`
	ReceiverPEP                 bool     `xml:"a:ReceiverPEP,omitempty"`
	ReceiverPOB                 string   `xml:"a:ReceiverPOB,omitempty"`
	ReceiverPhone               string   `xml:"a:ReceiverPhone,omitempty"`
	ReceiverPicture             string   `xml:"a:ReceiverPicture,omitempty"`
	ReceiverRemark              string   `xml:"a:ReceiverRemark,omitempty"`
	ReceiverState               string   `xml:"a:ReceiverState,omitempty"`
	ReceiverZip                 string   `xml:"a:ReceiverZip,omitempty"`
	SenderID                    int32    `xml:"a:SenderID,omitempty"`
}

type WsTransferInfo struct {
	XMLName                                   xml.Name `xml:"transfer"`
	XmlNS                                     string   `xml:"xmlns:a,attr"`
	AgenciaCodigo                             string   `xml:"a:AgenciaCodigo,omitempty"`
	AgencySpecialDiscounts                    string   `xml:"a:AgencySpecialDiscounts,omitempty"`
	AmountToReceive                           float64  `xml:"a:AmountToReceive,omitempty"`
	AttachedFile                              string   `xml:"a:AttachedFile,omitempty"`
	BancoCuenta                               string   `xml:"a:BancoCuenta,omitempty"`
	BancosId                                  string   `xml:"a:BancosId,omitempty"`
	BancosNombre                              string   `xml:"a:BancosNombre,omitempty"`
	BankAccount                               string   `xml:"a:BankAccount,omitempty"`
	BankCode                                  string   `xml:"a:BankCode,omitempty"`
	BeneficiarioCelular                       string   `xml:"a:BeneficiarioCelular,omitempty"`
	BeneficiarioCiudad                        string   `xml:"a:BeneficiarioCiudad,omitempty"`
	BeneficiarioDialCode                      string   `xml:"a:BeneficiarioDialCode,omitempty"`
	BeneficiarioEnviarSMS                     bool     `xml:"a:BeneficiarioEnviarSMS,omitempty"`
	BeneficiarioEstado                        string   `xml:"a:BeneficiarioEstado,omitempty"`
	BeneficiarioIdDescripcion                 string   `xml:"a:BeneficiarioIdDescripcion,omitempty"`
	BeneficiarioIdTipo                        string   `xml:"a:BeneficiarioIdTipo,omitempty"`
	BeneficiarioMensaje                       string   `xml:"a:BeneficiarioMensaje,omitempty"`
	BeneficiarioNotas                         string   `xml:"a:BeneficiarioNotas,omitempty"`
	BeneficiarioZip                           string   `xml:"a:BeneficiarioZip,omitempty"`
	BeneficiaryID                             int32    `xml:"a:BeneficiaryID,omitempty"`
	CardNumber                                string   `xml:"a:CardNumber,omitempty"`
	CardNumber_CVV                            string   `xml:"a:CardNumber_CVV,omitempty"`
	CardNumber_ExpDate                        string   `xml:"a:CardNumber_ExpDate,omitempty"`
	CargosAdicionales                         float64  `xml:"a:CargosAdicionales,omitempty"`
	CiudadBeneficiario                        int32    `xml:"a:CiudadBeneficiario,omitempty"`
	CiudadNombreBeneficiario                  string   `xml:"a:CiudadNombreBeneficiario,omitempty"`
	CiudadNombreRemitente                     string   `xml:"a:CiudadNombreRemitente,omitempty"`
	CiudadRemitente                           string   `xml:"a:CiudadRemitente,omitempty"`
	ComisionAgencia                           float64  `xml:"a:ComisionAgencia,omitempty"`
	ComisionAgenciaFX                         float64  `xml:"a:ComisionAgenciaFX,omitempty"`
	ComisionCompaniaFX                        float64  `xml:"a:ComisionCompaniaFX,omitempty"`
	CompanyComision                           float64  `xml:"a:CompanyComision,omitempty"`
	ComplianceBanksName                       string   `xml:"a:ComplianceBanksName,omitempty"`
	ComplianceBirthDate                       GMTDate  `xml:"a:ComplianceBirthDate,omitempty"`
	ComplianceDireccion2                      string   `xml:"a:ComplianceDireccion2,omitempty"`
	ComplianceEmployerAddress                 string   `xml:"a:ComplianceEmployerAddress,omitempty"`
	ComplianceEmployersName                   string   `xml:"a:ComplianceEmployersName,omitempty"`
	ComplianceHomeType                        string   `xml:"a:ComplianceHomeType,omitempty"`
	ComplianceIdExpirationDate                GMTDate  `xml:"a:ComplianceIdExpirationDate,omitempty"`
	ComplianceIdIssuer                        string   `xml:"a:ComplianceIdIssuer,omitempty"`
	ComplianceIdNumber                        string   `xml:"a:ComplianceIdNumber,omitempty"`
	ComplianceIdType                          string   `xml:"a:ComplianceIdType,omitempty"`
	ComplianceOccupation                      string   `xml:"a:ComplianceOccupation,omitempty"`
	ComplianceOtherSendingReason              string   `xml:"a:ComplianceOtherSendingReason,omitempty"`
	ComplianceOverLimitMessage                string   `xml:"a:ComplianceOverLimitMessage,omitempty"`
	CompliancePhoneConfirmation               string   `xml:"a:CompliancePhoneConfirmation,omitempty"`
	CompliancePhoneType                       string   `xml:"a:CompliancePhoneType,omitempty"`
	ComplianceRelationshipWithBeneficiary     string   `xml:"a:ComplianceRelationshipWithBeneficiary,omitempty"`
	ComplianceSSN                             string   `xml:"a:ComplianceSSN,omitempty"`
	ComplianceSendingReason                   string   `xml:"a:ComplianceSendingReason,omitempty"`
	ComplianceSourceOfFunds                   string   `xml:"a:ComplianceSourceOfFunds,omitempty"`
	ComplianceSpecial                         string   `xml:"a:ComplianceSpecial,omitempty"`
	ComplianceTextType                        string   `xml:"a:ComplianceTextType,omitempty"`
	ComplianceWorkPhone                       string   `xml:"a:ComplianceWorkPhone,omitempty"`
	CorrespondentCode                         string   `xml:"a:CorrespondentCode,omitempty"`
	CustomerID                                string   `xml:"a:CustomerID,omitempty"`
	DestinatarioApellido                      string   `xml:"a:DestinatarioApellido,omitempty"`
	DestinatarioNombre                        string   `xml:"a:DestinatarioNombre,omitempty"`
	DestinationCurrency                       string   `xml:"a:DestinationCurrency,omitempty"`
	DireccionBancoTexto                       string   `xml:"a:DireccionBancoTexto,omitempty"`
	DireccionBeneficiario                     string   `xml:"a:DireccionBeneficiario,omitempty"`
	DireccionCalleRemitente                   string   `xml:"a:DireccionCalleRemitente,omitempty"`
	DireccionRemitente                        string   `xml:"a:DireccionRemitente,omitempty"`
	EnableSave                                bool     `xml:"a:EnableSave,omitempty"`
	ExchangeRate                              float64  `xml:"a:ExchangeRate,omitempty"`
	ExchangeRateEffective                     string   `xml:"a:ExchangeRateEffective,omitempty"`
	ExisteTarifa                              bool     `xml:"a:ExisteTarifa,omitempty"`
	Fee                                       float64  `xml:"a:Fee"`
	FeeDiferencia                             float64  `xml:"a:FeeDiferencia,omitempty"`
	FeeExchangeRateUp                         float64  `xml:"a:FeeExchangeRateUp,omitempty"`
	FeeExchangeRateUpCant                     float64  `xml:"a:FeeExchangeRateUpCant,omitempty"`
	FeesByExachangeRate                       float64  `xml:"a:FeesByExachangeRate,omitempty"`
	FormaPago                                 int32    `xml:"a:FormaPago,omitempty"`
	FormaPagoCodigo                           string   `xml:"a:FormaPagoCodigo,omitempty"`
	ForzarNuevoRemitente                      bool     `xml:"a:ForzarNuevoRemitente,omitempty"`
	GiroGratis                                bool     `xml:"a:GiroGratis,omitempty"`
	HDFieldBeneficiary                        string   `xml:"a:HDFieldBeneficiary,omitempty"`
	HDFieldExchRate                           float64  `xml:"a:HDFieldExchRate,omitempty"`
	HDFieldSales                              float64  `xml:"a:HDFieldSales,omitempty"`
	HdFieldCompliance                         string   `xml:"a:HdFieldCompliance,omitempty"`
	HiloAmount                                float64  `xml:"a:HiloAmount,omitempty"`
	IDNo                                      string   `xml:"a:IDNo,omitempty"`
	InformacionAdicionalRemitenteBeneficiario string   `xml:"a:InformacionAdicionalRemitenteBeneficiario,omitempty"`
	IsBlocked                                 bool     `xml:"a:IsBlocked,omitempty"`
	IsSuspended                               bool     `xml:"a:IsSuspended,omitempty"`
	LegalIdBeneficiario                       string   `xml:"a:LegalIdBeneficiario,omitempty"`
	MTSID                                     int32    `xml:"a:MTSID,omitempty"`
	MarkUp                                    float64  `xml:"a:MarkUp,omitempty"`
	MovilRemitente                            string   `xml:"a:MovilRemitente,omitempty"`
	NetAmount                                 float64  `xml:"a:NetAmount,omitempty"`
	OFACBeneficiario                          bool     `xml:"a:OFACBeneficiario,omitempty"`
	OFACBeneficiaryBirthDate                  GMTDate  `xml:"a:OFACBeneficiaryBirthDate,omitempty"`
	OFACBeneficiaryPlaceOfBirth               string   `xml:"a:OFACBeneficiaryPlaceOfBirth,omitempty"`
	OFACCountryOfNationality                  string   `xml:"a:OFACCountryOfNationality,omitempty"`
	OFACPlaceOfBirth                          string   `xml:"a:OFACPlaceOfBirth,omitempty"`
	OFACRemitente                             bool     `xml:"a:OFACRemitente,omitempty"`
	OFACRemitenteObliga                       string   `xml:"a:OFACRemitenteObliga,omitempty"`
	OFACSenderBirthDate                       GMTDate  `xml:"a:OFACSenderBirthDate,omitempty"`
	OfficeCode                                string   `xml:"a:OfficeCode,omitempty"`
	OfficeNombre                              string   `xml:"a:OfficeNombre,omitempty"`
	OriginalCurrency                          string   `xml:"a:OriginalCurrency,omitempty"`
	OriginalPaymentMethod                     string   `xml:"a:OriginalPaymentMethod,omitempty"`
	Others                                    float64  `xml:"a:Others,omitempty"`
	OverLimitMessage                          string   `xml:"a:OverLimitMessage,omitempty"`
	PEPBeneficiarioMessage                    string   `xml:"a:PEPBeneficiarioMessage,omitempty"`
	PEPBeneficiarioScore                      float64  `xml:"a:PEPBeneficiarioScore,omitempty"`
	PEPRemitenteMessage                       string   `xml:"a:PEPRemitenteMessage,omitempty"`
	PEPRemitenteScore                         float64  `xml:"a:PEPRemitenteScore,omitempty"`
	POBofacBen                                string   `xml:"a:POBofacBen,omitempty"`
	PaisBeneficiario                          int32    `xml:"a:PaisBeneficiario,omitempty"`
	PaisBeneficiarioNombre                    string   `xml:"a:PaisBeneficiarioNombre,omitempty"`
	Promotion                                 int32    `xml:"a:Promotion,omitempty"`
	PuntosRemitenteIdCard                     string   `xml:"a:PuntosRemitenteIdCard,omitempty"`
	RealExchangeRate                          float64  `xml:"a:RealExchangeRate,omitempty"`
	ReceiverCity                              string   `xml:"a:ReceiverCity,omitempty"`
	ReceiverState                             string   `xml:"a:ReceiverState,omitempty"`
	RelationshipWithSenders                   string   `xml:"a:RelationshipWithSenders,omitempty"`
	RemitenteApellido                         string   `xml:"a:RemitenteApellido,omitempty"`
	RemitenteEmail                            string   `xml:"a:RemitenteEmail,omitempty"`
	RemitenteEstado                           string   `xml:"a:RemitenteEstado,omitempty"`
	RemitenteNombre                           string   `xml:"a:RemitenteNombre,omitempty"`
	RemitentePais                             string   `xml:"a:RemitentePais,omitempty"`
	RemitentePaisNombre                       string   `xml:"a:RemitentePaisNombre,omitempty"`
	RemitenteTelefono                         string   `xml:"a:RemitenteTelefono,omitempty"`
	RemitenteZip                              string   `xml:"a:RemitenteZip,omitempty"`
	RoutingNumber                             string   `xml:"a:RoutingNumber,omitempty"`
	SaveSenderReceiver                        bool     `xml:"a:SaveSenderReceiver,omitempty"`
	SenderAchAccount                          string   `xml:"a:SenderAchAccount,omitempty"`
	SenderAchRouting                          string   `xml:"a:SenderAchRouting,omitempty"`
	SenderAchType                             string   `xml:"a:SenderAchType,omitempty"`
	SenderBirthDate                           GMTDate  `xml:"a:SenderBirthDate,omitempty"`
	SenderID                                  int32    `xml:"a:SenderID,omitempty"`
	ServicioCodigo                            string   `xml:"a:ServicioCodigo,omitempty"`
	ServicioId                                int32    `xml:"a:ServicioId,omitempty"`
	SucursalBanco                             string   `xml:"a:SucursalBanco,omitempty"`
	SuspendMessage                            string   `xml:"a:SuspendMessage,omitempty"`
	SuspendUserType                           string   `xml:"a:SuspendUserType,omitempty"`
	TasaError                                 string   `xml:"a:TasaError,omitempty"`
	TelefonoBeneficiario                      string   `xml:"a:TelefonoBeneficiario,omitempty"`
	TempCompliance                            string   `xml:"a:TempCompliance,omitempty"`
	TempGiroRepetido                          string   `xml:"a:TempGiroRepetido,omitempty"`
	ThirdPartyReceipt                         string   `xml:"a:ThirdPartyReceipt,omitempty"`
	Ticket                                    string   `xml:"a:Ticket,omitempty"`
	TipoCuentaCodigo                          string   `xml:"a:TipoCuentaCodigo,omitempty"`
	TotalAditionalCharges                     float64  `xml:"a:TotalAditionalCharges,omitempty"`
	TotalAmount                               float64  `xml:"a:TotalAmount,omitempty"`
	NewTransaction_TipoCalculo                int32    `xml:"a:newTransaction_TipoCalculo,omitempty"`
}

type WsResponse struct {
	Text               string `xml:",chardata"`
	A                  string `xml:"a,attr"`
	I                  string `xml:"i,attr"`
	Error              int32  `xml:"Error,omitempty"`
	Message            string `xml:"Message,omitempty"`
	Password           string `xml:"Password,omitempty"`
	Receipt            string `xml:"Receipt,omitempty"`
	Receipt_Contact    string `xml:"Receipt_Contact,omitempty"`
	Receipt_Contact_EN string `xml:"Receipt_Contact_EN,omitempty"`
	Receipt_Error      string `xml:"Receipt_Error,omitempty"`
	Receipt_Error_EN   string `xml:"Receipt_Error_EN,omitempty"`
	Receipt_License    string `xml:"Receipt_License,omitempty"`
	Receipt_RTR        string `xml:"Receipt_RTR,omitempty"`
	Receipt_RTR_EN     string `xml:"Receipt_RTR_EN,omitempty"`
	ReceiverID         int32  `xml:"ReceiverID,omitempty"`
	SenderID           int32  `xml:"SenderID,omitempty"`
	Status             string `xml:"Status,omitempty"`
	Valid              bool   `xml:"Valid,omitempty"`
}

type ArrayOfwsPaidTransactions struct {
	WsPaidTransactions []*WsPaidTransactions `xml:"wsPaidTransactions,omitempty"`
}

type WsPaidTransactions struct {
	XMLName           xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsPaidTransactions"`
	Password          string   `xml:"Password,omitempty"`
	PaymentDate       GMTDate  `xml:"PaymentDate,omitempty"`
	Receipt           string   `xml:"Receipt,omitempty"`
	ThirdPartyReceipt string   `xml:"ThirdPartyReceipt,omitempty"`
}

type ArrayOfwsVoidTransactions struct {
	WsVoidTransactions []*WsVoidTransactions `xml:"wsVoidTransactions,omitempty"`
}

type WsVoidTransactions struct {
	XMLName           xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsVoidTransactions"`
	CancelDate        GMTDate  `xml:"CancelDate,omitempty"`
	Password          string   `xml:"Password,omitempty"`
	Receipt           string   `xml:"Receipt,omitempty"`
	ThirdPartyReceipt string   `xml:"ThirdPartyReceipt,omitempty"`
}

type WsChangeRequestData struct {
	XMLName                   xml.Name `xml:"wsChangeRequestData"`
	XmlNS                     string   `xml:"xmlns:a"`
	ReceiverAccount           string   `xml:"a:ReceiverAccount,omitempty"`
	ReceiverAccountType       string   `xml:"a:ReceiverAccountType,omitempty"`
	ReceiverAddress           string   `xml:"a:ReceiverAddress,omitempty"`
	ReceiverBankBranOrRouting string   `xml:"a:ReceiverBankBranOrRouting,omitempty"`
	ReceiverBankName          string   `xml:"a:ReceiverBankName,omitempty"`
	ReceiverBirthDate         GMTDate  `xml:"a:ReceiverBirthDate,omitempty"`
	ReceiverEmail             string   `xml:"a:ReceiverEmail,omitempty"`
	ReceiverIdNumber          string   `xml:"a:ReceiverIdNumber,omitempty"`
	ReceiverIdType            string   `xml:"a:ReceiverIdType,omitempty"`
	ReceiverLastName          string   `xml:"a:ReceiverLastName,omitempty"`
	ReceiverMobile            string   `xml:"a:ReceiverMobile,omitempty"`
	ReceiverName              string   `xml:"a:ReceiverName,omitempty"`
	ReceiverPhone             string   `xml:"a:ReceiverPhone,omitempty"`
	ReceiverZip               string   `xml:"a:ReceiverZip,omitempty"`
}

type ArrayOfwsModifiedTransactions struct {
	WsModifiedTransactions []*WsModifiedTransactions `xml:"wsModifiedTransactions,omitempty"`
}

type WsModifiedTransactions struct {
	XMLName                     xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsModifiedTransactions"`
	ModifyDate                  GMTDate  `xml:"ModifyDate,omitempty"`
	Password                    string   `xml:"Password,omitempty"`
	Receipt                     string   `xml:"Receipt,omitempty"`
	ReceiverAccount             string   `xml:"ReceiverAccount,omitempty"`
	ReceiverAccountType         string   `xml:"ReceiverAccountType,omitempty"`
	ReceiverAddress             string   `xml:"ReceiverAddress,omitempty"`
	ReceiverBankBranOrRouting   string   `xml:"ReceiverBankBranOrRouting,omitempty"`
	ReceiverBankName            string   `xml:"ReceiverBankName,omitempty"`
	ReceiverCountryNationallity string   `xml:"ReceiverCountryNationallity,omitempty"`
	ReceiverEmail               string   `xml:"ReceiverEmail,omitempty"`
	ReceiverIdNumber            string   `xml:"ReceiverIdNumber,omitempty"`
	ReceiverIdType              string   `xml:"ReceiverIdType,omitempty"`
	ReceiverLastName            string   `xml:"ReceiverLastName,omitempty"`
	ReceiverMobile              string   `xml:"ReceiverMobile,omitempty"`
	ReceiverName                string   `xml:"ReceiverName,omitempty"`
	ReceiverPOB                 string   `xml:"ReceiverPOB,omitempty"`
	ReceiverPhone               string   `xml:"ReceiverPhone,omitempty"`
	ReceiverZip                 string   `xml:"ReceiverZip,omitempty"`
	ThirdPartyReceipt           string   `xml:"ThirdPartyReceipt,omitempty"`
}

type ArrayOfwsReleasedTransactions struct {
	WsReleasedTransactions []*WsReleasedTransactions `xml:"wsReleasedTransactions,omitempty"`
}

type WsReleasedTransactions struct {
	XMLName           xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsReleasedTransactions"`
	HoldDate          GMTDate  `xml:"HoldDate,omitempty"`
	Password          string   `xml:"Password,omitempty"`
	Receipt           string   `xml:"Receipt,omitempty"`
	ThirdPartyReceipt string   `xml:"ThirdPartyReceipt,omitempty"`
}

type ArrayOfwsCollectedAch struct {
	WsCollectedAch []*WsCollectedAch `xml:"wsCollectedAch,omitempty"`
}

type WsCollectedAch struct {
	XMLName           xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsCollectedAch"`
	ClearedDate       GMTDate  `xml:"ClearedDate,omitempty"`
	Password          string   `xml:"Password,omitempty"`
	Receipt           string   `xml:"Receipt,omitempty"`
	ThirdPartyReceipt string   `xml:"ThirdPartyReceipt,omitempty"`
}

type ArrayOfwsNotifications struct {
	WsNotifications []*WsNotifications `xml:"wsNotifications"`
}

type WsNotifications struct {
	Text              string  `xml:",chardata"`
	A                 string  `xml:"a,attr"`
	I                 string  `xml:"i,attr"`
	Password          string  `xml:"Password,omitempty"`
	Receipt           string  `xml:"Receipt,omitempty"`
	Status            string  `xml:"Status,omitempty"`
	StatusDate        GMTDate `xml:"StatusDate,omitempty"`
	ThirdPartyReceipt string  `xml:"ThirdPartyReceipt,omitempty"`
	WireDate          GMTDate `xml:"WireDate,omitempty"`
}

type WsAchStatus struct {
	XMLName  xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsAchStatus"`
	Amount   *float64 `xml:"Amount,omitempty"`
	Name     string   `xml:"Name,omitempty"`
	Password string   `xml:"Password,omitempty"`
	Status   string   `xml:"Status,omitempty"`
}

type WsStatus struct {
	XMLName           xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsStatus"`
	LastComment       string   `xml:"LastComment,omitempty"`
	PartnerReference  string   `xml:"PartnerReference,omitempty"`
	Password          string   `xml:"Password,omitempty"`
	PayMethod         string   `xml:"PayMethod,omitempty"`
	Receipt           string   `xml:"Receipt,omitempty"`
	Status            string   `xml:"Status,omitempty"`
	ThirdPartyReceipt string   `xml:"ThirdPartyReceipt,omitempty"`
}

type WsExRate struct {
	XMLName xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsExRate"`
	ExRate  *float64 `xml:"ExRate,omitempty"`
}

type ArrayOfwsExRateList struct {
	WsExRateList []*WsExRateList `xml:"wsExRateList,omitempty"`
}

type WsExRateList struct {
	XMLName  xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsExRateList"`
	ExRate   *float64 `xml:"ExRate,omitempty"`
	Cantup   float64  `xml:"cantup,omitempty"`
	Code     string   `xml:"code,omitempty"`
	Country  string   `xml:"country,omitempty"`
	Currency string   `xml:"currency,omitempty"`
	Payer    string   `xml:"payer,omitempty"`
}

type WsOfac struct {
	Text            string  `xml:",chardata"`
	A               string  `xml:"a,attr"`
	I               string  `xml:"i,attr"`
	Error           int32   `xml:"Error,omitempty"`
	Message         string  `xml:"Message,omitempty"`
	OfacAddress     string  `xml:"OfacAddress,omitempty"`
	OfacAka         string  `xml:"OfacAka,omitempty"`
	OfacCitizenship string  `xml:"OfacCitizenship,omitempty"`
	OfacComment     string  `xml:"OfacComment,omitempty"`
	OfacDob         string  `xml:"OfacDob,omitempty"`
	OfacFirst       string  `xml:"OfacFirst,omitempty"`
	OfacIds         string  `xml:"OfacIds,omitempty"`
	OfacLast        string  `xml:"OfacLast,omitempty"`
	OfacName        string  `xml:"OfacName,omitempty"`
	OfacNationality string  `xml:"OfacNationality,omitempty"`
	OfacPob         string  `xml:"OfacPob,omitempty"`
	OfacProgram     string  `xml:"OfacProgram,omitempty"`
	OfacRemark      string  `xml:"OfacRemark,omitempty"`
	OfacScore       float64 `xml:"OfacScore,omitempty"`
	OfacSource      string  `xml:"OfacSource,omitempty"`
	OfacTitle       string  `xml:"OfacTitle,omitempty"`
	OfacType        string  `xml:"OfacType,omitempty"`
	Valid           bool    `xml:"Valid,omitempty"`
}

type WsResult struct {
	Text    string `xml:",chardata"`
	A       string `xml:"a,attr"`
	I       string `xml:"i,attr"`
	Error   int32  `xml:"Error,omitempty"`
	Message string `xml:"Message,omitempty"`
	Valid   bool   `xml:"Valid,omitempty"`
}

type ArrayOfws_Select_PayersByCountryCodeResult struct {
	Ws_Select_PayersByCountryCodeResult []*Ws_Select_PayersByCountryCodeResult `xml:"ws_Select_PayersByCountryCodeResult,omitempty"`
}

type Ws_Select_PayersByCountryCodeResult struct {
	XMLName           xml.Name                                  `xml:"http://schemas.datacontract.org/2004/07/gmtpay ws_Select_PayersByCountryCodeResult"`
	NeedsBank         bool                                      `xml:"NeedsBank,omitempty"`
	NeedsBenID        bool                                      `xml:"NeedsBenID,omitempty"`
	NeedsBranch       bool                                      `xml:"NeedsBranch,omitempty"`
	NeedsIBAN         bool                                      `xml:"NeedsIBAN,omitempty"`
	PayerAccountTypes *ArrayOfws_Select_PayerAccountTypesResult `xml:"PayerAccountTypes,omitempty"`
	PayerBanks        *ArrayOfws_Select_PayerBanksResult        `xml:"PayerBanks,omitempty"`
	PayerCode         string                                    `xml:"PayerCode,omitempty"`
	PayerCountryName  string                                    `xml:"PayerCountryName,omitempty"`
	PayerCurrencies   *ArrayOfws_Select_PayerCurrenciesResult   `xml:"PayerCurrencies,omitempty"`
	PayerLimits       *ArrayOfws_Select_PayerLimitsResult       `xml:"PayerLimits,omitempty"`
	PayerName         string                                    `xml:"PayerName,omitempty"`
	PayerOffices      *ArrayOfws_Select_PayerOfficesResult      `xml:"PayerOffices,omitempty"`
	PayerServices     *ArrayOfws_Select_PayerServicesResult     `xml:"PayerServices,omitempty"`
	PayerType         string                                    `xml:"PayerType,omitempty"`
}

type ArrayOfws_Select_PayerAccountTypesResult struct {
	Ws_Select_PayerAccountTypesResult []*Ws_Select_PayerAccountTypesResult `xml:"ws_Select_PayerAccountTypesResult,omitempty"`
}

type Ws_Select_PayerAccountTypesResult struct {
	XMLName         xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay ws_Select_PayerAccountTypesResult"`
	AccountTypeCode string   `xml:"AccountTypeCode,omitempty"`
	AccountTypeName string   `xml:"AccountTypeName,omitempty"`
}

type ArrayOfws_Select_PayerBanksResult struct {
	Ws_Select_PayerBanksResult []*Ws_Select_PayerBanksResult `xml:"ws_Select_PayerBanksResult,omitempty"`
}

type Ws_Select_PayerBanksResult struct {
	XMLName  xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay ws_Select_PayerBanksResult"`
	BankCode string   `xml:"BankCode,omitempty"`
	BankName string   `xml:"BankName,omitempty"`
}

type ArrayOfws_Select_PayerCurrenciesResult struct {
	Ws_Select_PayerCurrenciesResult []*Ws_Select_PayerCurrenciesResult `xml:"ws_Select_PayerCurrenciesResult,omitempty"`
}

type Ws_Select_PayerCurrenciesResult struct {
	XMLName      xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay ws_Select_PayerCurrenciesResult"`
	CurrencyCode string   `xml:"CurrencyCode,omitempty"`
	CurrencyName string   `xml:"CurrencyName,omitempty"`
}

type ArrayOfws_Select_PayerLimitsResult struct {
	Ws_Select_PayerLimitsResult []*Ws_Select_PayerLimitsResult `xml:"ws_Select_PayerLimitsResult,omitempty"`
}

type Ws_Select_PayerLimitsResult struct {
	XMLName         xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay ws_Select_PayerLimitsResult"`
	Amount          float64  `xml:"Amount,omitempty"`
	CurrencyCode    string   `xml:"CurrencyCode,omitempty"`
	Days            int32    `xml:"Days,omitempty"`
	OfficeGroupName string   `xml:"OfficeGroupName,omitempty"`
	ServiceCode     string   `xml:"ServiceCode,omitempty"`
	Type            string   `xml:"Type,omitempty"`
}

type ArrayOfws_Select_PayerOfficesResult struct {
	Ws_Select_PayerOfficesResult []*Ws_Select_PayerOfficesResult `xml:"ws_Select_PayerOfficesResult,omitempty"`
}

type Ws_Select_PayerOfficesResult struct {
	XMLName         xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay ws_Select_PayerOfficesResult"`
	City            string   `xml:"City,omitempty"`
	CountryCode     string   `xml:"CountryCode,omitempty"`
	OfficeCode      string   `xml:"OfficeCode,omitempty"`
	OfficeGroup     int32    `xml:"OfficeGroup,omitempty"`
	OfficeGroupName string   `xml:"OfficeGroupName,omitempty"`
	State           string   `xml:"State,omitempty"`
}

type ArrayOfws_Select_PayerServicesResult struct {
	Ws_Select_PayerServicesResult []*Ws_Select_PayerServicesResult `xml:"ws_Select_PayerServicesResult,omitempty"`
}

type Ws_Select_PayerServicesResult struct {
	XMLName     xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay ws_Select_PayerServicesResult"`
	ServiceCode string   `xml:"ServiceCode,omitempty"`
	ServiceName string   `xml:"ServiceName,omitempty"`
}

type ReceiptTemplate struct {
	Rec_contact    string `xml:"rec_contact,omitempty"`
	Rec_contact_en string `xml:"rec_contact_en,omitempty"`
	Rec_error      string `xml:"rec_error,omitempty"`
	Rec_error_en   string `xml:"rec_error_en,omitempty"`
	Rec_idioma_id  int32  `xml:"rec_idioma_id,omitempty"`
	Rec_license    string `xml:"rec_license,omitempty"`
	Rec_pdf        string `xml:"rec_pdf,omitempty"`
	Rec_rtr        string `xml:"rec_rtr,omitempty"`
	Rec_rtr_en     string `xml:"rec_rtr_en,omitempty"`
}

type CityByZip struct {
	CioCountryCode string `xml:"CioCountryCode,omitempty"`
	CioName        string `xml:"CioName,omitempty"`
	CioState       string `xml:"CioState,omitempty"`
	CioStateCode   string `xml:"CioStateCode,omitempty"`
}

type Ws_LimitByPayerResult struct {
	XMLName        xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay ws_LimitByPayerResult"`
	Amount         *float64 `xml:"Amount,omitempty"`
	ContinueSaving *bool    `xml:"ContinueSaving,omitempty"`
	Limit          *float64 `xml:"Limit,omitempty"`
	Message        string   `xml:"Message,omitempty"`
}

type Ws_Select_PromoResult struct {
	XMLName  xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay ws_Select_PromoResult"`
	MSG      string   `xml:"MSG,omitempty"`
	PRO_CANT *float64 `xml:"PRO_CANT,omitempty"`
	PRO_CODE string   `xml:"PRO_CODE,omitempty"`
	PRO_DESC *float64 `xml:"PRO_DESC,omitempty"`
	PRO_FEE  *float64 `xml:"PRO_FEE,omitempty"`
}

type WsWallet struct {
	XMLName    xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsWallet"`
	Balance    float64  `xml:"Balance,omitempty"`
	Error      int32    `xml:"Error,omitempty"`
	ExternalId string   `xml:"ExternalId,omitempty"`
	Message    string   `xml:"Message,omitempty"`
	SenderId   int32    `xml:"SenderId,omitempty"`
}

type Wallet_Operation struct {
	Amount           float64 `xml:"Amount,omitempty"`
	External_ID      string  `xml:"External_ID,omitempty"`
	SenderAchAccount string  `xml:"SenderAchAccount,omitempty"`
	SenderAchRouting string  `xml:"SenderAchRouting,omitempty"`
	SenderAchType    string  `xml:"SenderAchType,omitempty"`
	SenderId         int32   `xml:"SenderId,omitempty"`
}

type WsDocument struct {
	XMLName      xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsDocument"`
	Comments     string   `xml:"Comments,omitempty"`
	DocumentType string   `xml:"DocumentType,omitempty"`
	FileContent  *[]byte  `xml:"FileContent,omitempty"`
	Filename     string   `xml:"Filename,omitempty"`
	Password     string   `xml:"Password,omitempty"`
	SenderId     string   `xml:"SenderId,omitempty"`
}

type ArrayOfwsOccupations struct {
	WsOccupations []*WsOccupations `xml:"wsOccupations,omitempty"`
}

type WsOccupations struct {
	XMLName  xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsOccupations"`
	Job_Name string   `xml:"Job_Name,omitempty"`
}
