package external

import (
	"encoding/xml"
	"time"
)

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
	v := t.Format("2006/01/02")
	return e.EncodeElement(v, start)
}

func (gd *GMTDate) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	err := d.DecodeElement(&s, &start)
	if err != nil {
		return err
	}
	t, err := time.Parse("2006/01/02", s)
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
	XMLName xml.Name    `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
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
	XMLName                     xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsSender"`
	AgentAlias                  string   `xml:"AgentAlias,omitempty"`
	BusinessName                string   `xml:"BusinessName,omitempty"`
	BusinessWebsite             string   `xml:"BusinessWebsite,omitempty"`
	Debit                       bool     `xml:"Debit,omitempty"`
	RepresentativeID            string   `xml:"RepresentativeID,omitempty"`
	SenderAchAccount            string   `xml:"SenderAchAccount,omitempty"`
	SenderAchRouting            string   `xml:"SenderAchRouting,omitempty"`
	SenderAchType               string   `xml:"SenderAchType,omitempty"`
	SenderActive                bool     `xml:"SenderActive,omitempty"`
	SenderAddress               string   `xml:"SenderAddress,omitempty"`
	SenderAddressStreet         string   `xml:"SenderAddressStreet,omitempty"`
	SenderAggregatedTransfers   int32    `xml:"SenderAggregatedTransfers,omitempty"`
	SenderBank                  string   `xml:"SenderBank,omitempty"`
	SenderBirthDate             GMTDate  `xml:"SenderBirthDate,omitempty"`
	SenderCardBank              string   `xml:"SenderCardBank,omitempty"`
	SenderCardExpiration        GMTDate  `xml:"SenderCardExpiration,omitempty"`
	SenderCardName              string   `xml:"SenderCardName,omitempty"`
	SenderCardNumber            string   `xml:"SenderCardNumber,omitempty"`
	SenderCardType              int      `xml:"SenderCardType,omitempty"`
	SenderCity                  string   `xml:"SenderCity,omitempty"`
	SenderComments              string   `xml:"SenderComments,omitempty"`
	SenderComments3             string   `xml:"SenderComments3,omitempty"`
	SenderCompany               string   `xml:"SenderCompany,omitempty"`
	SenderCompanyAddress        string   `xml:"SenderCompanyAddress,omitempty"`
	SenderCompanyPhone          string   `xml:"SenderCompanyPhone,omitempty"`
	SenderCountryCode           string   `xml:"SenderCountryCode,omitempty"`
	SenderCountryNationallity   string   `xml:"SenderCountryNationallity,omitempty"`
	SenderCountryResidence      string   `xml:"SenderCountryResidence,omitempty"`
	SenderCurrencyCode          string   `xml:"SenderCurrencyCode,omitempty"`
	SenderDocExpiration         GMTDate  `xml:"SenderDocExpiration,omitempty"`
	SenderEmail                 string   `xml:"SenderEmail,omitempty"`
	SenderFileImg               string   `xml:"SenderFileImg,omitempty"`
	SenderFileImg2              string   `xml:"SenderFileImg2,omitempty"`
	SenderForceNew              bool     `xml:"SenderForceNew,omitempty"`
	SenderGender                string   `xml:"SenderGender,omitempty"`
	SenderIP                    string   `xml:"SenderIP,omitempty"`
	SenderId                    int32    `xml:"SenderId,omitempty"`
	SenderIdCard                string   `xml:"SenderIdCard,omitempty"`
	SenderIdIssuer              string   `xml:"SenderIdIssuer,omitempty"`
	SenderIdNumber              string   `xml:"SenderIdNumber,omitempty"`
	SenderIdNumber2             string   `xml:"SenderIdNumber2,omitempty"`
	SenderIdType                string   `xml:"SenderIdType,omitempty"`
	SenderIdType2               string   `xml:"SenderIdType2,omitempty"`
	SenderIsBusiness            bool     `xml:"SenderIsBusiness,omitempty"`
	SenderLastName              string   `xml:"SenderLastName,omitempty"`
	SenderMaritalStatus         string   `xml:"SenderMaritalStatus,omitempty"`
	SenderMobile                string   `xml:"SenderMobile,omitempty"`
	SenderMoneyOrigin           string   `xml:"SenderMoneyOrigin,omitempty"`
	SenderMoneyOwn              bool     `xml:"SenderMoneyOwn,omitempty"`
	SenderMonthAverage          float64  `xml:"SenderMonthAverage,omitempty"`
	SenderName                  string   `xml:"SenderName,omitempty"`
	SenderNatureOfBusiness      string   `xml:"SenderNatureOfBusiness,omitempty"`
	SenderOccupation            string   `xml:"SenderOccupation,omitempty"`
	SenderOnBehalfOf            bool     `xml:"SenderOnBehalfOf,omitempty"`
	SenderPEP                   bool     `xml:"SenderPEP,omitempty"`
	SenderPOB                   string   `xml:"SenderPOB,omitempty"`
	SenderPassword              string   `xml:"SenderPassword,omitempty"`
	SenderPhone                 string   `xml:"SenderPhone,omitempty"`
	SenderPicture               string   `xml:"SenderPicture,omitempty"`
	SenderPoliticalFamily       bool     `xml:"SenderPoliticalFamily,omitempty"`
	SenderResidenceAddress      string   `xml:"SenderResidenceAddress,omitempty"`
	SenderResidenceAddressExtra string   `xml:"SenderResidenceAddressExtra,omitempty"`
	SenderResidenceCity         string   `xml:"SenderResidenceCity,omitempty"`
	SenderResidenceCountryCode  string   `xml:"SenderResidenceCountryCode,omitempty"`
	SenderResidenceState        string   `xml:"SenderResidenceState,omitempty"`
	SenderResidenceZip          string   `xml:"SenderResidenceZip,omitempty"`
	SenderSendingReason         string   `xml:"SenderSendingReason,omitempty"`
	SenderState                 string   `xml:"SenderState,omitempty"`
	SenderTrackingNumber        string   `xml:"SenderTrackingNumber,omitempty"`
	SenderZip                   string   `xml:"SenderZip,omitempty"`
}

type WsReceiver struct {
	XMLName                     xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsReceiver"`
	ReceiverAchAccount          string   `xml:"ReceiverAchAccount,omitempty"`
	ReceiverAchRouting          string   `xml:"ReceiverAchRouting,omitempty"`
	ReceiverAchType             string   `xml:"ReceiverAchType,omitempty"`
	ReceiverActive              bool     `xml:"ReceiverActive,omitempty"`
	ReceiverAddress             string   `xml:"ReceiverAddress,omitempty"`
	ReceiverAverageMonth        float64  `xml:"ReceiverAverageMonth,omitempty"`
	ReceiverBirthDate           GMTDate  `xml:"ReceiverBirthDate,omitempty"`
	ReceiverCity                string   `xml:"ReceiverCity,omitempty"`
	ReceiverCompany             string   `xml:"ReceiverCompany,omitempty"`
	ReceiverCountry             string   `xml:"ReceiverCountry,omitempty"`
	ReceiverCountryNationallity string   `xml:"ReceiverCountryNationallity,omitempty"`
	ReceiverCurrency            string   `xml:"ReceiverCurrency,omitempty"`
	ReceiverDocExpiration       GMTDate  `xml:"ReceiverDocExpiration,omitempty"`
	ReceiverEmail               string   `xml:"ReceiverEmail,omitempty"`
	ReceiverFileImg             string   `xml:"ReceiverFileImg,omitempty"`
	ReceiverFileImg2            string   `xml:"ReceiverFileImg2,omitempty"`
	ReceiverGender              string   `xml:"ReceiverGender,omitempty"`
	ReceiverId                  int32    `xml:"ReceiverId,omitempty"`
	ReceiverIdIssuer            string   `xml:"ReceiverIdIssuer,omitempty"`
	ReceiverIdNumber            string   `xml:"ReceiverIdNumber,omitempty"`
	ReceiverIdType              string   `xml:"ReceiverIdType,omitempty"`
	ReceiverLastName            string   `xml:"ReceiverLastName,omitempty"`
	ReceiverLastTransaction     int32    `xml:"ReceiverLastTransaction,omitempty"`
	ReceiverMaritalStatus       string   `xml:"ReceiverMaritalStatus,omitempty"`
	ReceiverMobile              string   `xml:"ReceiverMobile,omitempty"`
	ReceiverMoneyOrigin         string   `xml:"ReceiverMoneyOrigin,omitempty"`
	ReceiverName                string   `xml:"ReceiverName,omitempty"`
	ReceiverOccupation          string   `xml:"ReceiverOccupation,omitempty"`
	ReceiverOfficeCode          string   `xml:"ReceiverOfficeCode,omitempty"`
	ReceiverPEP                 bool     `xml:"ReceiverPEP,omitempty"`
	ReceiverPOB                 string   `xml:"ReceiverPOB,omitempty"`
	ReceiverPhone               string   `xml:"ReceiverPhone,omitempty"`
	ReceiverPicture             string   `xml:"ReceiverPicture,omitempty"`
	ReceiverRemark              string   `xml:"ReceiverRemark,omitempty"`
	ReceiverState               string   `xml:"ReceiverState,omitempty"`
	ReceiverZip                 string   `xml:"ReceiverZip,omitempty"`
	SenderID                    int32    `xml:"SenderID,omitempty"`
}

type WsTransferInfo struct {
	XMLName                                   xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsTransferInfo"`
	AgenciaCodigo                             string   `xml:"AgenciaCodigo,omitempty"`
	AgencySpecialDiscounts                    string   `xml:"AgencySpecialDiscounts,omitempty"`
	AmountToReceive                           float64  `xml:"AmountToReceive,omitempty"`
	AttachedFile                              string   `xml:"AttachedFile,omitempty"`
	BancoCuenta                               string   `xml:"BancoCuenta,omitempty"`
	BancosId                                  string   `xml:"BancosId,omitempty"`
	BancosNombre                              string   `xml:"BancosNombre,omitempty"`
	BankAccount                               string   `xml:"BankAccount,omitempty"`
	BankCode                                  string   `xml:"BankCode,omitempty"`
	BeneficiarioCelular                       string   `xml:"BeneficiarioCelular,omitempty"`
	BeneficiarioCiudad                        string   `xml:"BeneficiarioCiudad,omitempty"`
	BeneficiarioDialCode                      string   `xml:"BeneficiarioDialCode,omitempty"`
	BeneficiarioEnviarSMS                     bool     `xml:"BeneficiarioEnviarSMS,omitempty"`
	BeneficiarioEstado                        string   `xml:"BeneficiarioEstado,omitempty"`
	BeneficiarioIdDescripcion                 string   `xml:"BeneficiarioIdDescripcion,omitempty"`
	BeneficiarioIdTipo                        string   `xml:"BeneficiarioIdTipo,omitempty"`
	BeneficiarioMensaje                       string   `xml:"BeneficiarioMensaje,omitempty"`
	BeneficiarioNotas                         string   `xml:"BeneficiarioNotas,omitempty"`
	BeneficiarioZip                           string   `xml:"BeneficiarioZip,omitempty"`
	BeneficiaryID                             int32    `xml:"BeneficiaryID,omitempty"`
	CardNumber                                string   `xml:"CardNumber,omitempty"`
	CardNumber_CVV                            string   `xml:"CardNumber_CVV,omitempty"`
	CardNumber_ExpDate                        string   `xml:"CardNumber_ExpDate,omitempty"`
	CargosAdicionales                         float64  `xml:"CargosAdicionales,omitempty"`
	CiudadBeneficiario                        int32    `xml:"CiudadBeneficiario,omitempty"`
	CiudadNombreBeneficiario                  string   `xml:"CiudadNombreBeneficiario,omitempty"`
	CiudadNombreRemitente                     string   `xml:"CiudadNombreRemitente,omitempty"`
	CiudadRemitente                           string   `xml:"CiudadRemitente,omitempty"`
	ComisionAgencia                           float64  `xml:"ComisionAgencia,omitempty"`
	ComisionAgenciaFX                         float64  `xml:"ComisionAgenciaFX,omitempty"`
	ComisionCompaniaFX                        float64  `xml:"ComisionCompaniaFX,omitempty"`
	CompanyComision                           float64  `xml:"CompanyComision,omitempty"`
	ComplianceBanksName                       string   `xml:"ComplianceBanksName,omitempty"`
	ComplianceBirthDate                       GMTDate  `xml:"ComplianceBirthDate,omitempty"`
	ComplianceDireccion2                      string   `xml:"ComplianceDireccion2,omitempty"`
	ComplianceEmployerAddress                 string   `xml:"ComplianceEmployerAddress,omitempty"`
	ComplianceEmployersName                   string   `xml:"ComplianceEmployersName,omitempty"`
	ComplianceHomeType                        string   `xml:"ComplianceHomeType,omitempty"`
	ComplianceIdExpirationDate                GMTDate  `xml:"ComplianceIdExpirationDate,omitempty"`
	ComplianceIdIssuer                        string   `xml:"ComplianceIdIssuer,omitempty"`
	ComplianceIdNumber                        string   `xml:"ComplianceIdNumber,omitempty"`
	ComplianceIdType                          string   `xml:"ComplianceIdType,omitempty"`
	ComplianceOccupation                      string   `xml:"ComplianceOccupation,omitempty"`
	ComplianceOtherSendingReason              string   `xml:"ComplianceOtherSendingReason,omitempty"`
	ComplianceOverLimitMessage                string   `xml:"ComplianceOverLimitMessage,omitempty"`
	CompliancePhoneConfirmation               string   `xml:"CompliancePhoneConfirmation,omitempty"`
	CompliancePhoneType                       string   `xml:"CompliancePhoneType,omitempty"`
	ComplianceRelationshipWithBeneficiary     string   `xml:"ComplianceRelationshipWithBeneficiary,omitempty"`
	ComplianceSSN                             string   `xml:"ComplianceSSN,omitempty"`
	ComplianceSendingReason                   string   `xml:"ComplianceSendingReason,omitempty"`
	ComplianceSourceOfFunds                   string   `xml:"ComplianceSourceOfFunds,omitempty"`
	ComplianceSpecial                         string   `xml:"ComplianceSpecial,omitempty"`
	ComplianceTextType                        string   `xml:"ComplianceTextType,omitempty"`
	ComplianceWorkPhone                       string   `xml:"ComplianceWorkPhone,omitempty"`
	CorrespondentCode                         string   `xml:"CorrespondentCode,omitempty"`
	CustomerID                                string   `xml:"CustomerID,omitempty"`
	DestinatarioApellido                      string   `xml:"DestinatarioApellido,omitempty"`
	DestinatarioNombre                        string   `xml:"DestinatarioNombre,omitempty"`
	DestinationCurrency                       string   `xml:"DestinationCurrency,omitempty"`
	DireccionBancoTexto                       string   `xml:"DireccionBancoTexto,omitempty"`
	DireccionBeneficiario                     string   `xml:"DireccionBeneficiario,omitempty"`
	DireccionCalleRemitente                   string   `xml:"DireccionCalleRemitente,omitempty"`
	DireccionRemitente                        string   `xml:"DireccionRemitente,omitempty"`
	EnableSave                                bool     `xml:"EnableSave,omitempty"`
	ExchangeRate                              float64  `xml:"ExchangeRate,omitempty"`
	ExchangeRateEffective                     string   `xml:"ExchangeRateEffective,omitempty"`
	ExisteTarifa                              bool     `xml:"ExisteTarifa,omitempty"`
	Fee                                       float64  `xml:"Fee"`
	FeeDiferencia                             float64  `xml:"FeeDiferencia,omitempty"`
	FeeExchangeRateUp                         float64  `xml:"FeeExchangeRateUp,omitempty"`
	FeeExchangeRateUpCant                     float64  `xml:"FeeExchangeRateUpCant,omitempty"`
	FeesByExachangeRate                       float64  `xml:"FeesByExachangeRate,omitempty"`
	FormaPago                                 int32    `xml:"FormaPago,omitempty"`
	FormaPagoCodigo                           string   `xml:"FormaPagoCodigo,omitempty"`
	ForzarNuevoRemitente                      bool     `xml:"ForzarNuevoRemitente,omitempty"`
	GiroGratis                                bool     `xml:"GiroGratis,omitempty"`
	HDFieldBeneficiary                        string   `xml:"HDFieldBeneficiary,omitempty"`
	HDFieldExchRate                           float64  `xml:"HDFieldExchRate,omitempty"`
	HDFieldSales                              float64  `xml:"HDFieldSales,omitempty"`
	HdFieldCompliance                         string   `xml:"HdFieldCompliance,omitempty"`
	HiloAmount                                float64  `xml:"HiloAmount,omitempty"`
	IDNo                                      string   `xml:"IDNo,omitempty"`
	InformacionAdicionalRemitenteBeneficiario string   `xml:"InformacionAdicionalRemitenteBeneficiario,omitempty"`
	IsBlocked                                 bool     `xml:"IsBlocked,omitempty"`
	IsSuspended                               bool     `xml:"IsSuspended,omitempty"`
	LegalIdBeneficiario                       string   `xml:"LegalIdBeneficiario,omitempty"`
	MTSID                                     int32    `xml:"MTSID"`
	MarkUp                                    float64  `xml:"MarkUp,omitempty"`
	MovilRemitente                            string   `xml:"MovilRemitente,omitempty"`
	NetAmount                                 float64  `xml:"NetAmount,omitempty"`
	OFACBeneficiario                          bool     `xml:"OFACBeneficiario,omitempty"`
	OFACBeneficiaryBirthDate                  string   `xml:"OFACBeneficiaryBirthDate,omitempty"`
	OFACBeneficiaryPlaceOfBirth               string   `xml:"OFACBeneficiaryPlaceOfBirth,omitempty"`
	OFACCountryOfNationality                  string   `xml:"OFACCountryOfNationality,omitempty"`
	OFACPlaceOfBirth                          string   `xml:"OFACPlaceOfBirth,omitempty"`
	OFACRemitente                             bool     `xml:"OFACRemitente,omitempty"`
	OFACRemitenteObliga                       string   `xml:"OFACRemitenteObliga,omitempty"`
	OFACSenderBirthDate                       string   `xml:"OFACSenderBirthDate,omitempty"`
	OfficeCode                                string   `xml:"OfficeCode,omitempty"`
	OfficeNombre                              string   `xml:"OfficeNombre,omitempty"`
	OriginalCurrency                          string   `xml:"OriginalCurrency,omitempty"`
	OriginalPaymentMethod                     string   `xml:"OriginalPaymentMethod,omitempty"`
	Others                                    float64  `xml:"Others,omitempty"`
	OverLimitMessage                          string   `xml:"OverLimitMessage,omitempty"`
	PEPBeneficiarioMessage                    string   `xml:"PEPBeneficiarioMessage,omitempty"`
	PEPBeneficiarioScore                      float64  `xml:"PEPBeneficiarioScore,omitempty"`
	PEPRemitenteMessage                       string   `xml:"PEPRemitenteMessage,omitempty"`
	PEPRemitenteScore                         float64  `xml:"PEPRemitenteScore,omitempty"`
	POBofacBen                                string   `xml:"POBofacBen,omitempty"`
	PaisBeneficiario                          int32    `xml:"PaisBeneficiario,omitempty"`
	PaisBeneficiarioNombre                    string   `xml:"PaisBeneficiarioNombre,omitempty"`
	Promotion                                 int32    `xml:"Promotion,omitempty"`
	PuntosRemitenteIdCard                     string   `xml:"PuntosRemitenteIdCard,omitempty"`
	RealExchangeRate                          float64  `xml:"RealExchangeRate,omitempty"`
	ReceiverCity                              string   `xml:"ReceiverCity,omitempty"`
	ReceiverState                             string   `xml:"ReceiverState,omitempty"`
	RelationshipWithSenders                   string   `xml:"RelationshipWithSenders,omitempty"`
	RemitenteApellido                         string   `xml:"RemitenteApellido,omitempty"`
	RemitenteEmail                            string   `xml:"RemitenteEmail,omitempty"`
	RemitenteEstado                           string   `xml:"RemitenteEstado,omitempty"`
	RemitenteNombre                           string   `xml:"RemitenteNombre,omitempty"`
	RemitentePais                             string   `xml:"RemitentePais,omitempty"`
	RemitentePaisNombre                       string   `xml:"RemitentePaisNombre,omitempty"`
	RemitenteTelefono                         string   `xml:"RemitenteTelefono,omitempty"`
	RemitenteZip                              string   `xml:"RemitenteZip,omitempty"`
	RoutingNumber                             string   `xml:"RoutingNumber,omitempty"`
	SaveSenderReceiver                        bool     `xml:"SaveSenderReceiver,omitempty"`
	SenderAchAccount                          string   `xml:"SenderAchAccount,omitempty"`
	SenderAchRouting                          string   `xml:"SenderAchRouting,omitempty"`
	SenderAchType                             string   `xml:"SenderAchType,omitempty"`
	SenderBirthDate                           GMTDate  `xml:"SenderBirthDate,omitempty"`
	SenderID                                  int32    `xml:"SenderID,omitempty"`
	ServicioCodigo                            string   `xml:"ServicioCodigo,omitempty"`
	ServicioId                                int32    `xml:"ServicioId,omitempty"`
	SucursalBanco                             string   `xml:"SucursalBanco,omitempty"`
	SuspendMessage                            string   `xml:"SuspendMessage,omitempty"`
	SuspendUserType                           string   `xml:"SuspendUserType,omitempty"`
	TasaError                                 string   `xml:"TasaError,omitempty"`
	TelefonoBeneficiario                      string   `xml:"TelefonoBeneficiario,omitempty"`
	TempCompliance                            string   `xml:"TempCompliance,omitempty"`
	TempGiroRepetido                          string   `xml:"TempGiroRepetido,omitempty"`
	ThirdPartyReceipt                         string   `xml:"ThirdPartyReceipt,omitempty"`
	Ticket                                    string   `xml:"Ticket,omitempty"`
	TipoCuentaCodigo                          string   `xml:"TipoCuentaCodigo,omitempty"`
	TotalAditionalCharges                     float64  `xml:"TotalAditionalCharges,omitempty"`
	TotalAmount                               float64  `xml:"TotalAmount,omitempty"`
	NewTransaction_TipoCalculo                int32    `xml:"newTransaction_TipoCalculo,omitempty"`
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
	XMLName                   xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsChangeRequestData"`
	ReceiverAccount           string   `xml:"ReceiverAccount,omitempty"`
	ReceiverAccountType       string   `xml:"ReceiverAccountType,omitempty"`
	ReceiverAddress           string   `xml:"ReceiverAddress,omitempty"`
	ReceiverBankBranOrRouting string   `xml:"ReceiverBankBranOrRouting,omitempty"`
	ReceiverBankName          string   `xml:"ReceiverBankName,omitempty"`
	ReceiverBirthDate         GMTDate  `xml:"ReceiverBirthDate,omitempty"`
	ReceiverEmail             string   `xml:"ReceiverEmail,omitempty"`
	ReceiverIdNumber          string   `xml:"ReceiverIdNumber,omitempty"`
	ReceiverIdType            string   `xml:"ReceiverIdType,omitempty"`
	ReceiverLastName          string   `xml:"ReceiverLastName,omitempty"`
	ReceiverMobile            string   `xml:"ReceiverMobile,omitempty"`
	ReceiverName              string   `xml:"ReceiverName,omitempty"`
	ReceiverPhone             string   `xml:"ReceiverPhone,omitempty"`
	ReceiverZip               string   `xml:"ReceiverZip,omitempty"`
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
