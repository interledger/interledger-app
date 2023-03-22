package external

import (
	"context"
	"encoding/xml"
	"time"

	"github.com/hooklift/gowsdl/soap"
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

type InsertTransaction struct {
	XMLName  xml.Name        `xml:"http://tempuri.org/ InsertTransaction"`
	Alias    string          `xml:"alias,omitempty" json:"alias,omitempty"`
	User     string          `xml:"user,omitempty" json:"user,omitempty"`
	Pass     string          `xml:"pass,omitempty" json:"pass,omitempty"`
	Sender   *WsSender       `json:"sender,omitempty"`
	Receiver *WsReceiver     `json:"receiver,omitempty"`
	Transfer *WsTransferInfo `json:"transfer,omitempty"`
}

type InsertTransactionResponse struct {
	XMLName                 xml.Name    `xml:"http://tempuri.org/ InsertTransactionResponse"`
	InsertTransactionResult *WsResponse `json:"InsertTransactionResult,omitempty"`
}

type GetPaidTransactions struct {
	XMLName xml.Name `xml:"http://tempuri.org/ GetPaidTransactions"`
	Alias   string   `xml:"alias,omitempty" json:"alias,omitempty"`
	User    string   `xml:"user,omitempty" json:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty" json:"pass,omitempty"`
}

type GetPaidTransactionsResponse struct {
	XMLName                   xml.Name                   `xml:"http://tempuri.org/ GetPaidTransactionsResponse"`
	GetPaidTransactionsResult *ArrayOfwsPaidTransactions `xml:"GetPaidTransactionsResult,omitempty" json:"GetPaidTransactionsResult,omitempty"`
}

type ConfirmPayment struct {
	XMLName xml.Name `xml:"http://tempuri.org/ ConfirmPayment"`
	Alias   string   `xml:"alias,omitempty" json:"alias,omitempty"`
	User    string   `xml:"user,omitempty " json:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty" json:"pass,omitempty"`
	Receipt string   `xml:"receipt,omitempty" json:"receipt,omitempty"`
}

type ConfirmPaymentResponse struct {
	XMLName              xml.Name    `xml:"http://tempuri.org/ ConfirmPaymentResponse"`
	ConfirmPaymentResult *WsResponse `xml:"ConfirmPaymentResult,omitempty" json:"ConfirmPaymentResult,omitempty"`
}

type RequestCancellation struct {
	XMLName xml.Name `xml:"http://tempuri.org/ RequestCancellation"`
	Alias   string   `xml:"alias,omitempty" json:"alias,omitempty"`
	User    string   `xml:"user,omitempty" json:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty" json:"pass,omitempty"`
	Receipt string   `xml:"receipt,omitempty" json:"receipt,omitempty"`
	Comment string   `xml:"comment,omitempty" json:"comment,omitempty"`
}

type RequestCancellationResponse struct {
	XMLName                   xml.Name    `xml:"http://tempuri.org/ RequestCancellationResponse"`
	RequestCancellationResult *WsResponse `xml:"RequestCancellationResult,omitempty" json:"RequestCancellationResult,omitempty"`
}

type GetCancelledTransactions struct {
	XMLName xml.Name `xml:"http://tempuri.org/ GetCancelledTransactions"`
	Alias   string   `xml:"alias,omitempty" json:"alias,omitempty"`
	User    string   `xml:"user,omitempty" json:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty" json:"pass,omitempty"`
}

type GetCancelledTransactionsResponse struct {
	XMLName                        xml.Name                   `xml:"http://tempuri.org/ GetCancelledTransactionsResponse"`
	GetCancelledTransactionsResult *ArrayOfwsVoidTransactions `xml:"GetCancelledTransactionsResult,omitempty" json:"GetCancelledTransactionsResult,omitempty"`
}

type ConfirmCancellation struct {
	XMLName xml.Name `xml:"http://tempuri.org/ ConfirmCancellation"`
	Alias   string   `xml:"alias,omitempty" json:"alias,omitempty"`
	User    string   `xml:"user,omitempty" json:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty" json:"pass,omitempty"`
	Receipt string   `xml:"receipt,omitempty" json:"receipt,omitempty"`
}

type ConfirmCancellationResponse struct {
	XMLName                   xml.Name    `xml:"http://tempuri.org/ ConfirmCancellationResponse"`
	ConfirmCancellationResult *WsResponse `xml:"ConfirmCancellationResult,omitempty" json:"ConfirmCancellationResult,omitempty"`
}

type RequestModification struct {
	XMLName xml.Name             `xml:"http://tempuri.org/ RequestModification"`
	Alias   string               `xml:"alias,omitempty" json:"alias,omitempty"`
	User    string               `xml:"user,omitempty" json:"user,omitempty"`
	Pass    string               `xml:"pass,omitempty" json:"pass,omitempty"`
	Receipt string               `xml:"receipt,omitempty" json:"receipt,omitempty"`
	Comment string               `xml:"comment,omitempty" json:"comment,omitempty"`
	Data    *WsChangeRequestData `xml:"data,omitempty" json:"data,omitempty"`
}

type RequestModificationResponse struct {
	XMLName                   xml.Name    `xml:"http://tempuri.org/ RequestModificationResponse"`
	RequestModificationResult *WsResponse `xml:"RequestModificationResult,omitempty" json:"RequestModificationResult,omitempty"`
}

type GetModifiedTransactions struct {
	XMLName xml.Name `xml:"http://tempuri.org/ GetModifiedTransactions"`
	Alias   string   `xml:"alias,omitempty" json:"alias,omitempty"`
	User    string   `xml:"user,omitempty" json:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty" json:"pass,omitempty"`
}

type GetModifiedTransactionsResponse struct {
	XMLName                       xml.Name                       `xml:"http://tempuri.org/ GetModifiedTransactionsResponse"`
	GetModifiedTransactionsResult *ArrayOfwsModifiedTransactions `xml:"GetModifiedTransactionsResult,omitempty" json:"GetModifiedTransactionsResult,omitempty"`
}

type ConfirmModification struct {
	XMLName xml.Name `xml:"http://tempuri.org/ ConfirmModification"`
	Alias   string   `xml:"alias,omitempty" json:"alias,omitempty"`
	User    string   `xml:"user,omitempty" json:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty" json:"pass,omitempty"`
	Receipt string   `xml:"receipt,omitempty" json:"receipt,omitempty"`
}

type ConfirmModificationResponse struct {
	XMLName                   xml.Name    `xml:"http://tempuri.org/ ConfirmModificationResponse"`
	ConfirmModificationResult *WsResponse `xml:"ConfirmModificationResult,omitempty" json:"ConfirmModificationResult,omitempty"`
}

type GetReleasedTransactions struct {
	XMLName xml.Name `xml:"http://tempuri.org/ GetReleasedTransactions"`
	Alias   string   `xml:"alias,omitempty" json:"alias,omitempty"`
	User    string   `xml:"user,omitempty" json:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty" json:"pass,omitempty"`
}

type GetReleasedTransactionsResponse struct {
	XMLName                       xml.Name                       `xml:"http://tempuri.org/ GetReleasedTransactionsResponse"`
	GetReleasedTransactionsResult *ArrayOfwsReleasedTransactions `xml:"GetReleasedTransactionsResult,omitempty" json:"GetReleasedTransactionsResult,omitempty"`
}

type ConfirmRelease struct {
	XMLName xml.Name `xml:"http://tempuri.org/ ConfirmRelease"`
	Alias   string   `xml:"alias,omitempty" json:"alias,omitempty"`
	User    string   `xml:"user,omitempty" json:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty" json:"pass,omitempty"`
	Receipt string   `xml:"receipt,omitempty" json:"receipt,omitempty"`
}

type ConfirmReleaseResponse struct {
	XMLName              xml.Name    `xml:"http://tempuri.org/ ConfirmReleaseResponse"`
	ConfirmReleaseResult *WsResponse `xml:"ConfirmReleaseResult,omitempty" json:"ConfirmReleaseResult,omitempty"`
}

type GetClearedAchTransactions struct {
	XMLName xml.Name `xml:"http://tempuri.org/ GetClearedAchTransactions"`
	Alias   string   `xml:"alias,omitempty" json:"alias,omitempty"`
	User    string   `xml:"user,omitempty" json:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty" json:"pass,omitempty"`
}

type GetClearedAchTransactionsResponse struct {
	XMLName                         xml.Name               `xml:"http://tempuri.org/ GetClearedAchTransactionsResponse"`
	GetClearedAchTransactionsResult *ArrayOfwsCollectedAch `xml:"GetClearedAchTransactionsResult,omitempty" json:"GetClearedAchTransactionsResult,omitempty"`
}

type ConfirmCollection struct {
	XMLName xml.Name `xml:"http://tempuri.org/ ConfirmCollection"`
	Alias   string   `xml:"alias,omitempty" json:"alias,omitempty"`
	User    string   `xml:"user,omitempty" json:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty" json:"pass,omitempty"`
	Receipt string   `xml:"receipt,omitempty" json:"receipt,omitempty"`
}

type ConfirmCollectionResponse struct {
	XMLName                 xml.Name    `xml:"http://tempuri.org/ ConfirmCollectionResponse"`
	ConfirmCollectionResult *WsResponse `xml:"ConfirmCollectionResult,omitempty" json:"ConfirmCollectionResult,omitempty"`
}

type GetNotifications struct {
	XMLName xml.Name `xml:"http://tempuri.org/ GetNotifications"`
	Alias   string   `xml:"alias,omitempty" json:"alias,omitempty"`
	User    string   `xml:"user,omitempty" json:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty" json:"pass,omitempty"`
}

type GetNotificationsResponse struct {
	XMLName                xml.Name                `xml:"http://tempuri.org/ GetNotificationsResponse"`
	GetNotificationsResult *ArrayOfwsNotifications `xml:"GetNotificationsResult,omitempty" json:"GetNotificationsResult,omitempty"`
}

type GetAchStatus struct {
	XMLName xml.Name `xml:"http://tempuri.org/ GetAchStatus"`
	Alias   string   `xml:"alias,omitempty" json:"alias,omitempty"`
	User    string   `xml:"user,omitempty" json:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty" json:"pass,omitempty"`
	Receipt string   `xml:"receipt,omitempty" json:"receipt,omitempty"`
}

type GetAchStatusResponse struct {
	XMLName            xml.Name     `xml:"http://tempuri.org/ GetAchStatusResponse"`
	GetAchStatusResult *WsAchStatus `xml:"GetAchStatusResult,omitempty" json:"GetAchStatusResult,omitempty"`
}

type GetTransactionStatus struct {
	XMLName xml.Name `xml:"http://tempuri.org/ GetTransactionStatus"`
	Alias   string   `xml:"alias,omitempty" json:"alias,omitempty"`
	User    string   `xml:"user,omitempty" json:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty" json:"pass,omitempty"`
	Receipt string   `xml:"receipt,omitempty" json:"receipt,omitempty"`
}

type GetTransactionStatusResponse struct {
	XMLName                    xml.Name  `xml:"http://tempuri.org/ GetTransactionStatusResponse"`
	GetTransactionStatusResult *WsStatus `xml:"GetTransactionStatusResult,omitempty" json:"GetTransactionStatusResult,omitempty"`
}

type GetSingleExchangeRate struct {
	XMLName  xml.Name `xml:"http://tempuri.org/ GetSingleExchangeRate"`
	Alias    string   `xml:"alias,omitempty" json:"alias,omitempty"`
	User     string   `xml:"user,omitempty" json:"user,omitempty"`
	Pass     string   `xml:"pass,omitempty" json:"pass,omitempty"`
	Payer    string   `xml:"payer,omitempty" json:"payer,omitempty"`
	Mts      string   `xml:"mts,omitempty" json:"mts,omitempty"`
	Service  string   `xml:"service,omitempty" json:"service,omitempty"`
	Currency string   `xml:"currency,omitempty" json:"currency,omitempty"`
}

type GetSingleExchangeRateResponse struct {
	XMLName                     xml.Name  `xml:"http://tempuri.org/ GetSingleExchangeRateResponse"`
	GetSingleExchangeRateResult *WsExRate `xml:"GetSingleExchangeRateResult,omitempty" json:"GetSingleExchangeRateResult,omitempty"`
}

type GetExchangeRates struct {
	XMLName  xml.Name `xml:"http://tempuri.org/ GetExchangeRates"`
	Alias    string   `xml:"alias,omitempty" json:"alias,omitempty"`
	User     string   `xml:"user,omitempty" json:"user,omitempty"`
	Pass     string   `xml:"pass,omitempty" json:"pass,omitempty"`
	Currency string   `xml:"currency,omitempty" json:"currency,omitempty"`
}

type GetExchangeRatesResponse struct {
	XMLName                xml.Name             `xml:"http://tempuri.org/ GetExchangeRatesResponse"`
	GetExchangeRatesResult *ArrayOfwsExRateList `xml:"GetExchangeRatesResult,omitempty" json:"GetExchangeRatesResult,omitempty"`
}

type OfacVerification struct {
	XMLName   xml.Name `xml:"http://tempuri.org/ OfacVerification"`
	Alias     string   `xml:"alias,omitempty" json:"alias,omitempty"`
	User      string   `xml:"user,omitempty" json:"user,omitempty"`
	Pass      string   `xml:"pass,omitempty" json:"pass,omitempty"`
	LastName  string   `xml:"lastName,omitempty" json:"lastName,omitempty"`
	FirstName string   `xml:"firstName,omitempty" json:"firstName,omitempty"`
}

type OfacVerificationResponse struct {
	XMLName                xml.Name `xml:"http://tempuri.org/ OfacVerificationResponse"`
	OfacVerificationResult *WsOfac  `xml:"OfacVerificationResult,omitempty" json:"OfacVerificationResult,omitempty"`
}

type ComplianceCheck struct {
	XMLName  xml.Name        `xml:"http://tempuri.org/ ComplianceCheck"`
	Alias    string          `xml:"alias,omitempty" json:"alias,omitempty"`
	User     string          `xml:"user,omitempty" json:"user,omitempty"`
	Pass     string          `xml:"pass,omitempty" json:"pass,omitempty"`
	Sender   *WsSender       `xml:"sender,omitempty" json:"sender,omitempty"`
	Receiver *WsReceiver     `xml:"receiver,omitempty" json:"receiver,omitempty"`
	Transfer *WsTransferInfo `xml:"transfer,omitempty" json:"transfer,omitempty"`
}

type ComplianceCheckResponse struct {
	XMLName               xml.Name    `xml:"http://tempuri.org/ ComplianceCheckResponse"`
	ComplianceCheckResult *WsResponse `xml:"ComplianceCheckResult,omitempty" json:"ComplianceCheckResult,omitempty"`
}

type SetVerified struct {
	XMLName xml.Name `xml:"http://tempuri.org/ SetVerified"`
	Alias   string   `xml:"alias,omitempty" json:"alias,omitempty"`
	User    string   `xml:"user,omitempty" json:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty" json:"pass,omitempty"`
	Receipt string   `xml:"receipt,omitempty" json:"receipt,omitempty"`
	Passed  bool     `xml:"passed,omitempty" json:"passed,omitempty"`
	Score   string   `xml:"score,omitempty" json:"score,omitempty"`
	Comment string   `xml:"comment,omitempty" json:"comment,omitempty"`
}

type SetVerifiedResponse struct {
	XMLName           xml.Name  `xml:"http://tempuri.org/ SetVerifiedResponse"`
	SetVerifiedResult *WsResult `xml:"SetVerifiedResult,omitempty" json:"SetVerifiedResult,omitempty"`
}

type PayersConsult struct {
	XMLName     xml.Name `xml:"http://tempuri.org/ PayersConsult"`
	Alias       string   `xml:"alias,omitempty" json:"alias,omitempty"`
	User        string   `xml:"user,omitempty" json:"user,omitempty"`
	Pass        string   `xml:"pass,omitempty" json:"pass,omitempty"`
	CountryCode string   `xml:"countryCode,omitempty" json:"countryCode,omitempty"`
	PayerCode   string   `xml:"payerCode,omitempty" json:"payerCode,omitempty"`
}

type PayersConsultResponse struct {
	XMLName             xml.Name                                    `xml:"http://tempuri.org/ PayersConsultResponse"`
	PayersConsultResult *ArrayOfws_Select_PayersByCountryCodeResult `xml:"PayersConsultResult,omitempty" json:"PayersConsultResult,omitempty"`
}

type GetReceiptData struct {
	XMLName xml.Name `xml:"http://tempuri.org/ GetReceiptData"`
	Alias   string   `xml:"alias,omitempty" json:"alias,omitempty"`
	User    string   `xml:"user,omitempty" json:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty" json:"pass,omitempty"`
	Rem_edo string   `xml:"rem_edo,omitempty" json:"rem_edo,omitempty"`
	Pais    string   `xml:"pais,omitempty" json:"pais,omitempty"`
}

type GetReceiptDataResponse struct {
	XMLName              xml.Name         `xml:"http://tempuri.org/ GetReceiptDataResponse"`
	GetReceiptDataResult *ReceiptTemplate `xml:"GetReceiptDataResult,omitempty" json:"GetReceiptDataResult,omitempty"`
}

type GetCityByZip struct {
	XMLName xml.Name `xml:"http://tempuri.org/ GetCityByZip"`
	Alias   string   `xml:"alias,omitempty" json:"alias,omitempty"`
	User    string   `xml:"user,omitempty" json:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty" json:"pass,omitempty"`
	Zip     string   `xml:"zip,omitempty" json:"zip,omitempty"`
}

type GetCityByZipResponse struct {
	XMLName            xml.Name   `xml:"http://tempuri.org/ GetCityByZipResponse"`
	GetCityByZipResult *CityByZip `xml:"GetCityByZipResult,omitempty" json:"GetCityByZipResult,omitempty"`
}

type CheckPayerLimits struct {
	XMLName     xml.Name `xml:"http://tempuri.org/ CheckPayerLimits"`
	Alias       string   `xml:"alias,omitempty" json:"alias,omitempty"`
	User        string   `xml:"user,omitempty" json:"user,omitempty"`
	Pass        string   `xml:"pass,omitempty" json:"pass,omitempty"`
	Payercode   string   `xml:"payercode,omitempty" json:"payercode,omitempty"`
	Send        float64  `xml:"send,omitempty" json:"send,omitempty"`
	Receive     float64  `xml:"receive,omitempty" json:"receive,omitempty"`
	Servicecode string   `xml:"servicecode,omitempty" json:"servicecode,omitempty"`
	Ben         string   `xml:"ben,omitempty" json:"ben,omitempty"`
	Rem         string   `xml:"rem,omitempty" json:"rem,omitempty"`
}

type CheckPayerLimitsResponse struct {
	XMLName                xml.Name               `xml:"http://tempuri.org/ CheckPayerLimitsResponse"`
	CheckPayerLimitsResult *Ws_LimitByPayerResult `xml:"CheckPayerLimitsResult,omitempty" json:"CheckPayerLimitsResult,omitempty"`
}

type PromotionsCode struct {
	XMLName    xml.Name `xml:"http://tempuri.org/ PromotionsCode"`
	Code_promo string   `xml:"Code_promo,omitempty" json:"Code_promo,omitempty"`
	Fee        float32  `xml:"Fee,omitempty" json:"Fee,omitempty"`
}

type PromotionsCodeResponse struct {
	XMLName              xml.Name               `xml:"http://tempuri.org/ PromotionsCodeResponse"`
	PromotionsCodeResult *Ws_Select_PromoResult `xml:"PromotionsCodeResult,omitempty" json:"PromotionsCodeResult,omitempty"`
}

type RegisterWallet struct {
	XMLName     xml.Name  `xml:"http://tempuri.org/ RegisterWallet"`
	Alias       string    `xml:"alias,omitempty" json:"alias,omitempty"`
	User        string    `xml:"user,omitempty" json:"user,omitempty"`
	Pass        string    `xml:"pass,omitempty" json:"pass,omitempty"`
	Sender      *WsSender `xml:"sender,omitempty" json:"sender,omitempty"`
	External_id string    `xml:"external_id,omitempty" json:"external_id,omitempty"`
}

type RegisterWalletResponse struct {
	XMLName              xml.Name  `xml:"http://tempuri.org/ RegisterWalletResponse"`
	RegisterWalletResult *WsWallet `xml:"RegisterWalletResult,omitempty" json:"RegisterWalletResult,omitempty"`
}

type AddWalletFunds struct {
	XMLName        xml.Name          `xml:"http://tempuri.org/ AddWalletFunds"`
	Alias          string            `xml:"alias,omitempty" json:"alias,omitempty"`
	User           string            `xml:"user,omitempty" json:"user,omitempty"`
	Pass           string            `xml:"pass,omitempty" json:"pass,omitempty"`
	Operation_Info *Wallet_Operation `xml:"Operation_Info,omitempty" json:"Operation_Info,omitempty"`
}

type AddWalletFundsResponse struct {
	XMLName              xml.Name  `xml:"http://tempuri.org/ AddWalletFundsResponse"`
	AddWalletFundsResult *WsWallet `xml:"AddWalletFundsResult,omitempty" json:"AddWalletFundsResult,omitempty"`
}

type WithdrawWalletFunds struct {
	XMLName        xml.Name          `xml:"http://tempuri.org/ WithdrawWalletFunds"`
	Alias          string            `xml:"alias,omitempty" json:"alias,omitempty"`
	User           string            `xml:"user,omitempty" json:"user,omitempty"`
	Pass           string            `xml:"pass,omitempty" json:"pass,omitempty"`
	Operation_Info *Wallet_Operation `xml:"Operation_Info,omitempty" json:"Operation_Info,omitempty"`
}

type WithdrawWalletFundsResponse struct {
	XMLName                   xml.Name  `xml:"http://tempuri.org/ WithdrawWalletFundsResponse"`
	WithdrawWalletFundsResult *WsWallet `xml:"WithdrawWalletFundsResult,omitempty" json:"WithdrawWalletFundsResult,omitempty"`
}

type GetWalletBalance struct {
	XMLName    xml.Name `xml:"http://tempuri.org/ GetWalletBalance"`
	Alias      string   `xml:"alias,omitempty" json:"alias,omitempty"`
	User       string   `xml:"user,omitempty" json:"user,omitempty"`
	Pass       string   `xml:"pass,omitempty" json:"pass,omitempty"`
	SenderId   int32    `xml:"senderId,omitempty" json:"senderId,omitempty"`
	ExternalId string   `xml:"ExternalId,omitempty" json:"ExternalId,omitempty"`
}

type GetWalletBalanceResponse struct {
	XMLName                xml.Name  `xml:"http://tempuri.org/ GetWalletBalanceResponse"`
	GetWalletBalanceResult *WsWallet `xml:"GetWalletBalanceResult,omitempty" json:"GetWalletBalanceResult,omitempty"`
}

type AddDocument struct {
	XMLName  xml.Name    `xml:"http://tempuri.org/ AddDocument"`
	Alias    string      `xml:"alias,omitempty" json:"alias,omitempty"`
	User     string      `xml:"user,omitempty" json:"user,omitempty"`
	Pass     string      `xml:"pass,omitempty" json:"pass,omitempty"`
	Document *WsDocument `xml:"Document,omitempty" json:"Document,omitempty"`
}

type AddDocumentResponse struct {
	XMLName           xml.Name    `xml:"http://tempuri.org/ AddDocumentResponse"`
	AddDocumentResult *WsResponse `xml:"AddDocumentResult,omitempty" json:"AddDocumentResult,omitempty"`
}

type GetOccupations struct {
	XMLName xml.Name `xml:"http://tempuri.org/ GetOccupations"`
	Alias   string   `xml:"alias,omitempty" json:"alias,omitempty"`
	User    string   `xml:"user,omitempty" json:"user,omitempty"`
	Pass    string   `xml:"pass,omitempty" json:"pass,omitempty"`
}

type GetOccupationsResponse struct {
	XMLName              xml.Name              `xml:"http://tempuri.org/ GetOccupationsResponse"`
	GetOccupationsResult *ArrayOfwsOccupations `xml:"GetOccupationsResult,omitempty" json:"GetOccupationsResult,omitempty"`
}

type WsSender struct {
	XMLName                     xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsSender"`
	AgentAlias                  string   `xml:"AgentAlias,omitempty" json:"AgentAlias,omitempty"`
	BusinessName                string   `xml:"BusinessName,omitempty" json:"BusinessName,omitempty"`
	BusinessWebsite             string   `xml:"BusinessWebsite,omitempty" json:"BusinessWebsite,omitempty"`
	Debit                       bool     `xml:"Debit,omitempty" json:"Debit,omitempty"`
	RepresentativeID            string   `xml:"RepresentativeID,omitempty" json:"RepresentativeID,omitempty"`
	SenderAchAccount            string   `xml:"SenderAchAccount,omitempty" json:"SenderAchAccount,omitempty"`
	SenderAchRouting            string   `xml:"SenderAchRouting,omitempty" json:"SenderAchRouting,omitempty"`
	SenderAchType               string   `xml:"SenderAchType,omitempty" json:"SenderAchType,omitempty"`
	SenderActive                bool     `xml:"SenderActive,omitempty" json:"SenderActive,omitempty"`
	SenderAddress               string   `xml:"SenderAddress,omitempty" json:"SenderAddress,omitempty"`
	SenderAddressStreet         string   `xml:"SenderAddressStreet,omitempty" json:"SenderAddressStreet,omitempty"`
	SenderAggregatedTransfers   int32    `xml:"SenderAggregatedTransfers,omitempty" json:"SenderAggregatedTransfers,omitempty"`
	SenderBank                  string   `xml:"SenderBank,omitempty" json:"SenderBank,omitempty"`
	SenderBirthDate             GMTDate  `xml:"SenderBirthDate,omitempty" json:"SenderBirthDate,omitempty"`
	SenderCardBank              string   `xml:"SenderCardBank,omitempty" json:"SenderCardBank,omitempty"`
	SenderCardExpiration        GMTDate  `xml:"SenderCardExpiration,omitempty" json:"SenderCardExpiration,omitempty"`
	SenderCardName              string   `xml:"SenderCardName,omitempty" json:"SenderCardName,omitempty"`
	SenderCardNumber            string   `xml:"SenderCardNumber,omitempty" json:"SenderCardNumber,omitempty"`
	SenderCardType              int      `xml:"SenderCardType,omitempty" json:"SenderCardType,omitempty"`
	SenderCity                  string   `xml:"SenderCity,omitempty" json:"SenderCity,omitempty"`
	SenderComments              string   `xml:"SenderComments,omitempty" json:"SenderComments,omitempty"`
	SenderComments3             string   `xml:"SenderComments3,omitempty" json:"SenderComments3,omitempty"`
	SenderCompany               string   `xml:"SenderCompany,omitempty" json:"SenderCompany,omitempty"`
	SenderCompanyAddress        string   `xml:"SenderCompanyAddress,omitempty" json:"SenderCompanyAddress,omitempty"`
	SenderCompanyPhone          string   `xml:"SenderCompanyPhone,omitempty" json:"SenderCompanyPhone,omitempty"`
	SenderCountryCode           string   `xml:"SenderCountryCode,omitempty" json:"SenderCountryCode,omitempty"`
	SenderCountryNationallity   string   `xml:"SenderCountryNationallity,omitempty" json:"SenderCountryNationallity,omitempty"`
	SenderCountryResidence      string   `xml:"SenderCountryResidence,omitempty" json:"SenderCountryResidence,omitempty"`
	SenderCurrencyCode          string   `xml:"SenderCurrencyCode,omitempty" json:"SenderCurrencyCode,omitempty"`
	SenderDocExpiration         GMTDate  `xml:"SenderDocExpiration,omitempty" json:"SenderDocExpiration,omitempty"`
	SenderEmail                 string   `xml:"SenderEmail,omitempty" json:"SenderEmail,omitempty"`
	SenderFileImg               string   `xml:"SenderFileImg,omitempty" json:"SenderFileImg,omitempty"`
	SenderFileImg2              string   `xml:"SenderFileImg2,omitempty" json:"SenderFileImg2,omitempty"`
	SenderForceNew              bool     `xml:"SenderForceNew,omitempty" json:"SenderForceNew,omitempty"`
	SenderGender                string   `xml:"SenderGender,omitempty" json:"SenderGender,omitempty"`
	SenderIP                    string   `xml:"SenderIP,omitempty" json:"SenderIP,omitempty"`
	SenderId                    int32    `xml:"SenderId,omitempty" json:"SenderId,omitempty"`
	SenderIdCard                string   `xml:"SenderIdCard,omitempty" json:"SenderIdCard,omitempty"`
	SenderIdIssuer              string   `xml:"SenderIdIssuer,omitempty" json:"SenderIdIssuer,omitempty"`
	SenderIdNumber              string   `xml:"SenderIdNumber,omitempty" json:"SenderIdNumber,omitempty"`
	SenderIdNumber2             string   `xml:"SenderIdNumber2,omitempty" json:"SenderIdNumber2,omitempty"`
	SenderIdType                string   `xml:"SenderIdType,omitempty" json:"SenderIdType,omitempty"`
	SenderIdType2               string   `xml:"SenderIdType2,omitempty" json:"SenderIdType2,omitempty"`
	SenderIsBusiness            bool     `xml:"SenderIsBusiness,omitempty" json:"SenderIsBusiness,omitempty"`
	SenderLastName              string   `xml:"SenderLastName,omitempty" json:"SenderLastName,omitempty"`
	SenderMaritalStatus         string   `xml:"SenderMaritalStatus,omitempty" json:"SenderMaritalStatus,omitempty"`
	SenderMobile                string   `xml:"SenderMobile,omitempty" json:"SenderMobile,omitempty"`
	SenderMoneyOrigin           string   `xml:"SenderMoneyOrigin,omitempty" json:"SenderMoneyOrigin,omitempty"`
	SenderMoneyOwn              bool     `xml:"SenderMoneyOwn,omitempty" json:"SenderMoneyOwn,omitempty"`
	SenderMonthAverage          float64  `xml:"SenderMonthAverage,omitempty" json:"SenderMonthAverage,omitempty"`
	SenderName                  string   `xml:"SenderName,omitempty" json:"SenderName,omitempty"`
	SenderNatureOfBusiness      string   `xml:"SenderNatureOfBusiness,omitempty" json:"SenderNatureOfBusiness,omitempty"`
	SenderOccupation            string   `xml:"SenderOccupation,omitempty" json:"SenderOccupation,omitempty"`
	SenderOnBehalfOf            bool     `xml:"SenderOnBehalfOf,omitempty" json:"SenderOnBehalfOf,omitempty"`
	SenderPEP                   bool     `xml:"SenderPEP,omitempty" json:"SenderPEP,omitempty"`
	SenderPOB                   string   `xml:"SenderPOB,omitempty" json:"SenderPOB,omitempty"`
	SenderPassword              string   `xml:"SenderPassword,omitempty" json:"SenderPassword,omitempty"`
	SenderPhone                 string   `xml:"SenderPhone,omitempty" json:"SenderPhone,omitempty"`
	SenderPicture               string   `xml:"SenderPicture,omitempty" json:"SenderPicture,omitempty"`
	SenderPoliticalFamily       bool     `xml:"SenderPoliticalFamily,omitempty" json:"SenderPoliticalFamily,omitempty"`
	SenderResidenceAddress      string   `xml:"SenderResidenceAddress,omitempty" json:"SenderResidenceAddress,omitempty"`
	SenderResidenceAddressExtra string   `xml:"SenderResidenceAddressExtra,omitempty" json:"SenderResidenceAddressExtra,omitempty"`
	SenderResidenceCity         string   `xml:"SenderResidenceCity,omitempty" json:"SenderResidenceCity,omitempty"`
	SenderResidenceCountryCode  string   `xml:"SenderResidenceCountryCode,omitempty" json:"SenderResidenceCountryCode,omitempty"`
	SenderResidenceState        string   `xml:"SenderResidenceState,omitempty" json:"SenderResidenceState,omitempty"`
	SenderResidenceZip          string   `xml:"SenderResidenceZip,omitempty" json:"SenderResidenceZip,omitempty"`
	SenderSendingReason         string   `xml:"SenderSendingReason,omitempty" json:"SenderSendingReason,omitempty"`
	SenderState                 string   `xml:"SenderState,omitempty" json:"SenderState,omitempty"`
	SenderTrackingNumber        string   `xml:"SenderTrackingNumber,omitempty" json:"SenderTrackingNumber,omitempty"`
	SenderZip                   string   `xml:"SenderZip,omitempty" json:"SenderZip,omitempty"`
}

type WsReceiver struct {
	XMLName                     xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsReceiver"`
	ReceiverAchAccount          string   `xml:"ReceiverAchAccount,omitempty" json:"ReceiverAchAccount,omitempty"`
	ReceiverAchRouting          string   `xml:"ReceiverAchRouting,omitempty" json:"ReceiverAchRouting,omitempty"`
	ReceiverAchType             string   `xml:"ReceiverAchType,omitempty" json:"ReceiverAchType,omitempty"`
	ReceiverActive              bool     `xml:"ReceiverActive,omitempty" json:"ReceiverActive,omitempty"`
	ReceiverAddress             string   `xml:"ReceiverAddress,omitempty" json:"ReceiverAddress,omitempty"`
	ReceiverAverageMonth        float64  `xml:"ReceiverAverageMonth,omitempty" json:"ReceiverAverageMonth,omitempty"`
	ReceiverBirthDate           GMTDate  `xml:"ReceiverBirthDate,omitempty" json:"ReceiverBirthDate,omitempty"`
	ReceiverCity                string   `xml:"ReceiverCity,omitempty" json:"ReceiverCity,omitempty"`
	ReceiverCompany             string   `xml:"ReceiverCompany,omitempty" json:"ReceiverCompany,omitempty"`
	ReceiverCountry             string   `xml:"ReceiverCountry,omitempty" json:"ReceiverCountry,omitempty"`
	ReceiverCountryNationallity string   `xml:"ReceiverCountryNationallity,omitempty" json:"ReceiverCountryNationallity,omitempty"`
	ReceiverCurrency            string   `xml:"ReceiverCurrency,omitempty" json:"ReceiverCurrency,omitempty"`
	ReceiverDocExpiration       GMTDate  `xml:"ReceiverDocExpiration,omitempty" json:"ReceiverDocExpiration,omitempty"`
	ReceiverEmail               string   `xml:"ReceiverEmail,omitempty" json:"ReceiverEmail,omitempty"`
	ReceiverFileImg             string   `xml:"ReceiverFileImg,omitempty" json:"ReceiverFileImg,omitempty"`
	ReceiverFileImg2            string   `xml:"ReceiverFileImg2,omitempty" json:"ReceiverFileImg2,omitempty"`
	ReceiverGender              string   `xml:"ReceiverGender,omitempty" json:"ReceiverGender,omitempty"`
	ReceiverId                  int32    `xml:"ReceiverId,omitempty" json:"ReceiverId,omitempty"`
	ReceiverIdIssuer            string   `xml:"ReceiverIdIssuer,omitempty" json:"ReceiverIdIssuer,omitempty"`
	ReceiverIdNumber            string   `xml:"ReceiverIdNumber,omitempty" json:"ReceiverIdNumber,omitempty"`
	ReceiverIdType              string   `xml:"ReceiverIdType,omitempty" json:"ReceiverIdType,omitempty"`
	ReceiverLastName            string   `xml:"ReceiverLastName,omitempty" json:"ReceiverLastName,omitempty"`
	ReceiverLastTransaction     int32    `xml:"ReceiverLastTransaction,omitempty" json:"ReceiverLastTransaction,omitempty"`
	ReceiverMaritalStatus       string   `xml:"ReceiverMaritalStatus,omitempty" json:"ReceiverMaritalStatus,omitempty"`
	ReceiverMobile              string   `xml:"ReceiverMobile,omitempty" json:"ReceiverMobile,omitempty"`
	ReceiverMoneyOrigin         string   `xml:"ReceiverMoneyOrigin,omitempty" json:"ReceiverMoneyOrigin,omitempty"`
	ReceiverName                string   `xml:"ReceiverName,omitempty" json:"ReceiverName,omitempty"`
	ReceiverOccupation          string   `xml:"ReceiverOccupation,omitempty" json:"ReceiverOccupation,omitempty"`
	ReceiverOfficeCode          string   `xml:"ReceiverOfficeCode,omitempty" json:"ReceiverOfficeCode,omitempty"`
	ReceiverPEP                 bool     `xml:"ReceiverPEP,omitempty" json:"ReceiverPEP,omitempty"`
	ReceiverPOB                 string   `xml:"ReceiverPOB,omitempty" json:"ReceiverPOB,omitempty"`
	ReceiverPhone               string   `xml:"ReceiverPhone,omitempty" json:"ReceiverPhone,omitempty"`
	ReceiverPicture             string   `xml:"ReceiverPicture,omitempty" json:"ReceiverPicture,omitempty"`
	ReceiverRemark              string   `xml:"ReceiverRemark,omitempty" json:"ReceiverRemark,omitempty"`
	ReceiverState               string   `xml:"ReceiverState,omitempty" json:"ReceiverState,omitempty"`
	ReceiverZip                 string   `xml:"ReceiverZip,omitempty" json:"ReceiverZip,omitempty"`
	SenderID                    int32    `xml:"SenderID,omitempty" json:"SenderID,omitempty"`
}

type WsTransferInfo struct {
	XMLName                                   xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsTransferInfo"`
	AgenciaCodigo                             string   `xml:"AgenciaCodigo,omitempty" json:"AgenciaCodigo,omitempty"`
	AgencySpecialDiscounts                    string   `xml:"AgencySpecialDiscounts,omitempty" json:"AgencySpecialDiscounts,omitempty"`
	AmountToReceive                           float64  `xml:"AmountToReceive,omitempty" json:"AmountToReceive,omitempty"`
	AttachedFile                              string   `xml:"AttachedFile,omitempty" json:"AttachedFile,omitempty"`
	BancoCuenta                               string   `xml:"BancoCuenta,omitempty" json:"BancoCuenta,omitempty"`
	BancosId                                  string   `xml:"BancosId,omitempty" json:"BancosId,omitempty"`
	BancosNombre                              string   `xml:"BancosNombre,omitempty" json:"BancosNombre,omitempty"`
	BankAccount                               string   `xml:"BankAccount,omitempty" json:"BankAccount,omitempty"`
	BankCode                                  string   `xml:"BankCode,omitempty" json:"BankCode,omitempty"`
	BeneficiarioCelular                       string   `xml:"BeneficiarioCelular,omitempty" json:"BeneficiarioCelular,omitempty"`
	BeneficiarioCiudad                        string   `xml:"BeneficiarioCiudad,omitempty" json:"BeneficiarioCiudad,omitempty"`
	BeneficiarioDialCode                      string   `xml:"BeneficiarioDialCode,omitempty" json:"BeneficiarioDialCode,omitempty"`
	BeneficiarioEnviarSMS                     bool     `xml:"BeneficiarioEnviarSMS,omitempty" json:"BeneficiarioEnviarSMS,omitempty"`
	BeneficiarioEstado                        string   `xml:"BeneficiarioEstado,omitempty" json:"BeneficiarioEstado,omitempty"`
	BeneficiarioIdDescripcion                 string   `xml:"BeneficiarioIdDescripcion,omitempty" json:"BeneficiarioIdDescripcion,omitempty"`
	BeneficiarioIdTipo                        string   `xml:"BeneficiarioIdTipo,omitempty" json:"BeneficiarioIdTipo,omitempty"`
	BeneficiarioMensaje                       string   `xml:"BeneficiarioMensaje,omitempty" json:"BeneficiarioMensaje,omitempty"`
	BeneficiarioNotas                         string   `xml:"BeneficiarioNotas,omitempty" json:"BeneficiarioNotas,omitempty"`
	BeneficiarioZip                           string   `xml:"BeneficiarioZip,omitempty" json:"BeneficiarioZip,omitempty"`
	BeneficiaryID                             int32    `xml:"BeneficiaryID,omitempty" json:"BeneficiaryID,omitempty"`
	CardNumber                                string   `xml:"CardNumber,omitempty" json:"CardNumber,omitempty"`
	CardNumber_CVV                            string   `xml:"CardNumber_CVV,omitempty" json:"CardNumber_CVV,omitempty"`
	CardNumber_ExpDate                        string   `xml:"CardNumber_ExpDate,omitempty" json:"CardNumber_ExpDate,omitempty"`
	CargosAdicionales                         float64  `xml:"CargosAdicionales,omitempty" json:"CargosAdicionales,omitempty"`
	CiudadBeneficiario                        int32    `xml:"CiudadBeneficiario,omitempty" json:"CiudadBeneficiario,omitempty"`
	CiudadNombreBeneficiario                  string   `xml:"CiudadNombreBeneficiario,omitempty" json:"CiudadNombreBeneficiario,omitempty"`
	CiudadNombreRemitente                     string   `xml:"CiudadNombreRemitente,omitempty" json:"CiudadNombreRemitente,omitempty"`
	CiudadRemitente                           string   `xml:"CiudadRemitente,omitempty" json:"CiudadRemitente,omitempty"`
	ComisionAgencia                           float64  `xml:"ComisionAgencia,omitempty" json:"ComisionAgencia,omitempty"`
	ComisionAgenciaFX                         float64  `xml:"ComisionAgenciaFX,omitempty" json:"ComisionAgenciaFX,omitempty"`
	ComisionCompaniaFX                        float64  `xml:"ComisionCompaniaFX,omitempty" json:"ComisionCompaniaFX,omitempty"`
	CompanyComision                           float64  `xml:"CompanyComision,omitempty" json:"CompanyComision,omitempty"`
	ComplianceBanksName                       string   `xml:"ComplianceBanksName,omitempty" json:"ComplianceBanksName,omitempty"`
	ComplianceBirthDate                       GMTDate  `xml:"ComplianceBirthDate,omitempty" json:"ComplianceBirthDate,omitempty"`
	ComplianceDireccion2                      string   `xml:"ComplianceDireccion2,omitempty" json:"ComplianceDireccion2,omitempty"`
	ComplianceEmployerAddress                 string   `xml:"ComplianceEmployerAddress,omitempty" json:"ComplianceEmployerAddress,omitempty"`
	ComplianceEmployersName                   string   `xml:"ComplianceEmployersName,omitempty" json:"ComplianceEmployersName,omitempty"`
	ComplianceHomeType                        string   `xml:"ComplianceHomeType,omitempty" json:"ComplianceHomeType,omitempty"`
	ComplianceIdExpirationDate                GMTDate  `xml:"ComplianceIdExpirationDate,omitempty" json:"ComplianceIdExpirationDate,omitempty"`
	ComplianceIdIssuer                        string   `xml:"ComplianceIdIssuer,omitempty" json:"ComplianceIdIssuer,omitempty"`
	ComplianceIdNumber                        string   `xml:"ComplianceIdNumber,omitempty" json:"ComplianceIdNumber,omitempty"`
	ComplianceIdType                          string   `xml:"ComplianceIdType,omitempty" json:"ComplianceIdType,omitempty"`
	ComplianceOccupation                      string   `xml:"ComplianceOccupation,omitempty" json:"ComplianceOccupation,omitempty"`
	ComplianceOtherSendingReason              string   `xml:"ComplianceOtherSendingReason,omitempty" json:"ComplianceOtherSendingReason,omitempty"`
	ComplianceOverLimitMessage                string   `xml:"ComplianceOverLimitMessage,omitempty" json:"ComplianceOverLimitMessage,omitempty"`
	CompliancePhoneConfirmation               string   `xml:"CompliancePhoneConfirmation,omitempty" json:"CompliancePhoneConfirmation,omitempty"`
	CompliancePhoneType                       string   `xml:"CompliancePhoneType,omitempty" json:"CompliancePhoneType,omitempty"`
	ComplianceRelationshipWithBeneficiary     string   `xml:"ComplianceRelationshipWithBeneficiary,omitempty" json:"ComplianceRelationshipWithBeneficiary,omitempty"`
	ComplianceSSN                             string   `xml:"ComplianceSSN,omitempty" json:"ComplianceSSN,omitempty"`
	ComplianceSendingReason                   string   `xml:"ComplianceSendingReason,omitempty" json:"ComplianceSendingReason,omitempty"`
	ComplianceSourceOfFunds                   string   `xml:"ComplianceSourceOfFunds,omitempty" json:"ComplianceSourceOfFunds,omitempty"`
	ComplianceSpecial                         string   `xml:"ComplianceSpecial,omitempty" json:"ComplianceSpecial,omitempty"`
	ComplianceTextType                        string   `xml:"ComplianceTextType,omitempty" json:"ComplianceTextType,omitempty"`
	ComplianceWorkPhone                       string   `xml:"ComplianceWorkPhone,omitempty" json:"ComplianceWorkPhone,omitempty"`
	CorrespondentCode                         string   `xml:"CorrespondentCode,omitempty" json:"CorrespondentCode,omitempty"`
	CustomerID                                string   `xml:"CustomerID,omitempty" json:"CustomerID,omitempty"`
	DestinatarioApellido                      string   `xml:"DestinatarioApellido,omitempty" json:"DestinatarioApellido,omitempty"`
	DestinatarioNombre                        string   `xml:"DestinatarioNombre,omitempty" json:"DestinatarioNombre,omitempty"`
	DestinationCurrency                       string   `xml:"DestinationCurrency,omitempty" json:"DestinationCurrency,omitempty"`
	DireccionBancoTexto                       string   `xml:"DireccionBancoTexto,omitempty" json:"DireccionBancoTexto,omitempty"`
	DireccionBeneficiario                     string   `xml:"DireccionBeneficiario,omitempty" json:"DireccionBeneficiario,omitempty"`
	DireccionCalleRemitente                   string   `xml:"DireccionCalleRemitente,omitempty" json:"DireccionCalleRemitente,omitempty"`
	DireccionRemitente                        string   `xml:"DireccionRemitente,omitempty" json:"DireccionRemitente,omitempty"`
	EnableSave                                bool     `xml:"EnableSave,omitempty" json:"EnableSave,omitempty"`
	ExchangeRate                              float64  `xml:"ExchangeRate,omitempty" json:"ExchangeRate,omitempty"`
	ExchangeRateEffective                     string   `xml:"ExchangeRateEffective,omitempty" json:"ExchangeRateEffective,omitempty"`
	ExisteTarifa                              bool     `xml:"ExisteTarifa,omitempty" json:"ExisteTarifa,omitempty"`
	Fee                                       float64  `xml:"Fee" json:"Fee,omitempty"`
	FeeDiferencia                             float64  `xml:"FeeDiferencia,omitempty" json:"FeeDiferencia,omitempty"`
	FeeExchangeRateUp                         float64  `xml:"FeeExchangeRateUp,omitempty" json:"FeeExchangeRateUp,omitempty"`
	FeeExchangeRateUpCant                     float64  `xml:"FeeExchangeRateUpCant,omitempty" json:"FeeExchangeRateUpCant,omitempty"`
	FeesByExachangeRate                       float64  `xml:"FeesByExachangeRate,omitempty" json:"FeesByExachangeRate,omitempty"`
	FormaPago                                 int32    `xml:"FormaPago,omitempty" json:"FormaPago,omitempty"`
	FormaPagoCodigo                           string   `xml:"FormaPagoCodigo,omitempty" json:"FormaPagoCodigo,omitempty"`
	ForzarNuevoRemitente                      bool     `xml:"ForzarNuevoRemitente,omitempty" json:"ForzarNuevoRemitente,omitempty"`
	GiroGratis                                bool     `xml:"GiroGratis,omitempty" json:"GiroGratis,omitempty"`
	HDFieldBeneficiary                        string   `xml:"HDFieldBeneficiary,omitempty" json:"HDFieldBeneficiary,omitempty"`
	HDFieldExchRate                           float64  `xml:"HDFieldExchRate,omitempty" json:"HDFieldExchRate,omitempty"`
	HDFieldSales                              float64  `xml:"HDFieldSales,omitempty" json:"HDFieldSales,omitempty"`
	HdFieldCompliance                         string   `xml:"HdFieldCompliance,omitempty" json:"HdFieldCompliance,omitempty"`
	HiloAmount                                float64  `xml:"HiloAmount,omitempty" json:"HiloAmount,omitempty"`
	IDNo                                      string   `xml:"IDNo,omitempty" json:"IDNo,omitempty"`
	InformacionAdicionalRemitenteBeneficiario string   `xml:"InformacionAdicionalRemitenteBeneficiario,omitempty" json:"InformacionAdicionalRemitenteBeneficiario,omitempty"`
	IsBlocked                                 bool     `xml:"IsBlocked,omitempty" json:"IsBlocked,omitempty"`
	IsSuspended                               bool     `xml:"IsSuspended,omitempty" json:"IsSuspended,omitempty"`
	LegalIdBeneficiario                       string   `xml:"LegalIdBeneficiario,omitempty" json:"LegalIdBeneficiario,omitempty"`
	MTSID                                     int32    `xml:"MTSID,omitempty" json:"MTSID,omitempty"`
	MarkUp                                    float64  `xml:"MarkUp,omitempty" json:"MarkUp,omitempty"`
	MovilRemitente                            string   `xml:"MovilRemitente,omitempty" json:"MovilRemitente,omitempty"`
	NetAmount                                 float64  `xml:"NetAmount,omitempty" json:"NetAmount,omitempty"`
	OFACBeneficiario                          bool     `xml:"OFACBeneficiario,omitempty" json:"OFACBeneficiario,omitempty"`
	OFACBeneficiaryBirthDate                  string   `xml:"OFACBeneficiaryBirthDate,omitempty" json:"OFACBeneficiaryBirthDate,omitempty"`
	OFACBeneficiaryPlaceOfBirth               string   `xml:"OFACBeneficiaryPlaceOfBirth,omitempty" json:"OFACBeneficiaryPlaceOfBirth,omitempty"`
	OFACCountryOfNationality                  string   `xml:"OFACCountryOfNationality,omitempty" json:"OFACCountryOfNationality,omitempty"`
	OFACPlaceOfBirth                          string   `xml:"OFACPlaceOfBirth,omitempty" json:"OFACPlaceOfBirth,omitempty"`
	OFACRemitente                             bool     `xml:"OFACRemitente,omitempty" json:"OFACRemitente,omitempty"`
	OFACRemitenteObliga                       string   `xml:"OFACRemitenteObliga,omitempty" json:"OFACRemitenteObliga,omitempty"`
	OFACSenderBirthDate                       string   `xml:"OFACSenderBirthDate,omitempty" json:"OFACSenderBirthDate,omitempty"`
	OfficeCode                                string   `xml:"OfficeCode,omitempty" json:"OfficeCode,omitempty"`
	OfficeNombre                              string   `xml:"OfficeNombre,omitempty" json:"OfficeNombre,omitempty"`
	OriginalCurrency                          string   `xml:"OriginalCurrency,omitempty" json:"OriginalCurrency,omitempty"`
	OriginalPaymentMethod                     string   `xml:"OriginalPaymentMethod,omitempty" json:"OriginalPaymentMethod,omitempty"`
	Others                                    float64  `xml:"Others,omitempty" json:"Others,omitempty"`
	OverLimitMessage                          string   `xml:"OverLimitMessage,omitempty" json:"OverLimitMessage,omitempty"`
	PEPBeneficiarioMessage                    string   `xml:"PEPBeneficiarioMessage,omitempty" json:"PEPBeneficiarioMessage,omitempty"`
	PEPBeneficiarioScore                      float64  `xml:"PEPBeneficiarioScore,omitempty" json:"PEPBeneficiarioScore,omitempty"`
	PEPRemitenteMessage                       string   `xml:"PEPRemitenteMessage,omitempty" json:"PEPRemitenteMessage,omitempty"`
	PEPRemitenteScore                         float64  `xml:"PEPRemitenteScore,omitempty" json:"PEPRemitenteScore,omitempty"`
	POBofacBen                                string   `xml:"POBofacBen,omitempty" json:"POBofacBen,omitempty"`
	PaisBeneficiario                          int32    `xml:"PaisBeneficiario,omitempty" json:"PaisBeneficiario,omitempty"`
	PaisBeneficiarioNombre                    string   `xml:"PaisBeneficiarioNombre,omitempty" json:"PaisBeneficiarioNombre,omitempty"`
	Promotion                                 int32    `xml:"Promotion,omitempty" json:"Promotion,omitempty"`
	PuntosRemitenteIdCard                     string   `xml:"PuntosRemitenteIdCard,omitempty" json:"PuntosRemitenteIdCard,omitempty"`
	RealExchangeRate                          float64  `xml:"RealExchangeRate,omitempty" json:"RealExchangeRate,omitempty"`
	ReceiverCity                              string   `xml:"ReceiverCity,omitempty" json:"ReceiverCity,omitempty"`
	ReceiverState                             string   `xml:"ReceiverState,omitempty" json:"ReceiverState,omitempty"`
	RelationshipWithSenders                   string   `xml:"RelationshipWithSenders,omitempty" json:"RelationshipWithSenders,omitempty"`
	RemitenteApellido                         string   `xml:"RemitenteApellido,omitempty" json:"RemitenteApellido,omitempty"`
	RemitenteEmail                            string   `xml:"RemitenteEmail,omitempty" json:"RemitenteEmail,omitempty"`
	RemitenteEstado                           string   `xml:"RemitenteEstado,omitempty" json:"RemitenteEstado,omitempty"`
	RemitenteNombre                           string   `xml:"RemitenteNombre,omitempty" json:"RemitenteNombre,omitempty"`
	RemitentePais                             string   `xml:"RemitentePais,omitempty" json:"RemitentePais,omitempty"`
	RemitentePaisNombre                       string   `xml:"RemitentePaisNombre,omitempty" json:"RemitentePaisNombre,omitempty"`
	RemitenteTelefono                         string   `xml:"RemitenteTelefono,omitempty" json:"RemitenteTelefono,omitempty"`
	RemitenteZip                              string   `xml:"RemitenteZip,omitempty" json:"RemitenteZip,omitempty"`
	RoutingNumber                             string   `xml:"RoutingNumber,omitempty" json:"RoutingNumber,omitempty"`
	SaveSenderReceiver                        bool     `xml:"SaveSenderReceiver,omitempty" json:"SaveSenderReceiver,omitempty"`
	SenderAchAccount                          string   `xml:"SenderAchAccount,omitempty" json:"SenderAchAccount,omitempty"`
	SenderAchRouting                          string   `xml:"SenderAchRouting,omitempty" json:"SenderAchRouting,omitempty"`
	SenderAchType                             string   `xml:"SenderAchType,omitempty" json:"SenderAchType,omitempty"`
	SenderBirthDate                           GMTDate  `xml:"SenderBirthDate,omitempty" json:"SenderBirthDate,omitempty"`
	SenderID                                  int32    `xml:"SenderID,omitempty" json:"SenderID,omitempty"`
	ServicioCodigo                            string   `xml:"ServicioCodigo,omitempty" json:"ServicioCodigo,omitempty"`
	ServicioId                                int32    `xml:"ServicioId,omitempty" json:"ServicioId,omitempty"`
	SucursalBanco                             string   `xml:"SucursalBanco,omitempty" json:"SucursalBanco,omitempty"`
	SuspendMessage                            string   `xml:"SuspendMessage,omitempty" json:"SuspendMessage,omitempty"`
	SuspendUserType                           string   `xml:"SuspendUserType,omitempty" json:"SuspendUserType,omitempty"`
	TasaError                                 string   `xml:"TasaError,omitempty" json:"TasaError,omitempty"`
	TelefonoBeneficiario                      string   `xml:"TelefonoBeneficiario,omitempty" json:"TelefonoBeneficiario,omitempty"`
	TempCompliance                            string   `xml:"TempCompliance,omitempty" json:"TempCompliance,omitempty"`
	TempGiroRepetido                          string   `xml:"TempGiroRepetido,omitempty" json:"TempGiroRepetido,omitempty"`
	ThirdPartyReceipt                         string   `xml:"ThirdPartyReceipt,omitempty" json:"ThirdPartyReceipt,omitempty"`
	Ticket                                    string   `xml:"Ticket,omitempty" json:"Ticket,omitempty"`
	TipoCuentaCodigo                          string   `xml:"TipoCuentaCodigo,omitempty" json:"TipoCuentaCodigo,omitempty"`
	TotalAditionalCharges                     float64  `xml:"TotalAditionalCharges,omitempty" json:"TotalAditionalCharges,omitempty"`
	TotalAmount                               float64  `xml:"TotalAmount,omitempty" json:"TotalAmount,omitempty"`
	NewTransaction_TipoCalculo                int32    `xml:"newTransaction_TipoCalculo,omitempty" json:"newTransaction_TipoCalculo,omitempty"`
}

type WsResponse struct {
	XMLName            xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsResponse"`
	Error              int32    `xml:"Error,omitempty" json:"Error,omitempty"`
	Message            string   `xml:"Message,omitempty" json:"Message,omitempty"`
	Password           string   `xml:"Password,omitempty" json:"Password,omitempty"`
	Receipt            string   `xml:"Receipt,omitempty" json:"Receipt,omitempty"`
	Receipt_Contact    string   `xml:"Receipt_Contact,omitempty" json:"Receipt_Contact,omitempty"`
	Receipt_Contact_EN string   `xml:"Receipt_Contact_EN,omitempty" json:"Receipt_Contact_EN,omitempty"`
	Receipt_Error      string   `xml:"Receipt_Error,omitempty" json:"Receipt_Error,omitempty"`
	Receipt_Error_EN   string   `xml:"Receipt_Error_EN,omitempty" json:"Receipt_Error_EN,omitempty"`
	Receipt_License    string   `xml:"Receipt_License,omitempty" json:"Receipt_License,omitempty"`
	Receipt_RTR        string   `xml:"Receipt_RTR,omitempty" json:"Receipt_RTR,omitempty"`
	Receipt_RTR_EN     string   `xml:"Receipt_RTR_EN,omitempty" json:"Receipt_RTR_EN,omitempty"`
	ReceiverID         int32    `xml:"ReceiverID,omitempty" json:"ReceiverID,omitempty"`
	SenderID           int32    `xml:"SenderID,omitempty" json:"SenderID,omitempty"`
	Status             string   `xml:"Status,omitempty" json:"Status,omitempty"`
	Valid              bool     `xml:"Valid,omitempty" json:"Valid,omitempty"`
}

type ArrayOfwsPaidTransactions struct {
	WsPaidTransactions []*WsPaidTransactions `xml:"wsPaidTransactions,omitempty" json:"wsPaidTransactions,omitempty"`
}

type WsPaidTransactions struct {
	XMLName           xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsPaidTransactions"`
	Password          string   `xml:"Password,omitempty" json:"Password,omitempty"`
	PaymentDate       GMTDate  `xml:"PaymentDate,omitempty" json:"PaymentDate,omitempty"`
	Receipt           string   `xml:"Receipt,omitempty" json:"Receipt,omitempty"`
	ThirdPartyReceipt string   `xml:"ThirdPartyReceipt,omitempty" json:"ThirdPartyReceipt,omitempty"`
}

type ArrayOfwsVoidTransactions struct {
	WsVoidTransactions []*WsVoidTransactions `xml:"wsVoidTransactions,omitempty" json:"wsVoidTransactions,omitempty"`
}

type WsVoidTransactions struct {
	XMLName           xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsVoidTransactions"`
	CancelDate        GMTDate  `xml:"CancelDate,omitempty" json:"CancelDate,omitempty"`
	Password          string   `xml:"Password,omitempty" json:"Password,omitempty"`
	Receipt           string   `xml:"Receipt,omitempty" json:"Receipt,omitempty"`
	ThirdPartyReceipt string   `xml:"ThirdPartyReceipt,omitempty" json:"ThirdPartyReceipt,omitempty"`
}

type WsChangeRequestData struct {
	XMLName                   xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsChangeRequestData"`
	ReceiverAccount           string   `xml:"ReceiverAccount,omitempty" json:"ReceiverAccount,omitempty"`
	ReceiverAccountType       string   `xml:"ReceiverAccountType,omitempty" json:"ReceiverAccountType,omitempty"`
	ReceiverAddress           string   `xml:"ReceiverAddress,omitempty" json:"ReceiverAddress,omitempty"`
	ReceiverBankBranOrRouting string   `xml:"ReceiverBankBranOrRouting,omitempty" json:"ReceiverBankBranOrRouting,omitempty"`
	ReceiverBankName          string   `xml:"ReceiverBankName,omitempty" json:"ReceiverBankName,omitempty"`
	ReceiverBirthDate         GMTDate  `xml:"ReceiverBirthDate,omitempty" json:"ReceiverBirthDate,omitempty"`
	ReceiverEmail             string   `xml:"ReceiverEmail,omitempty" json:"ReceiverEmail,omitempty"`
	ReceiverIdNumber          string   `xml:"ReceiverIdNumber,omitempty" json:"ReceiverIdNumber,omitempty"`
	ReceiverIdType            string   `xml:"ReceiverIdType,omitempty" json:"ReceiverIdType,omitempty"`
	ReceiverLastName          string   `xml:"ReceiverLastName,omitempty" json:"ReceiverLastName,omitempty"`
	ReceiverMobile            string   `xml:"ReceiverMobile,omitempty" json:"ReceiverMobile,omitempty"`
	ReceiverName              string   `xml:"ReceiverName,omitempty" json:"ReceiverName,omitempty"`
	ReceiverPhone             string   `xml:"ReceiverPhone,omitempty" json:"ReceiverPhone,omitempty"`
	ReceiverZip               string   `xml:"ReceiverZip,omitempty" json:"ReceiverZip,omitempty"`
}

type ArrayOfwsModifiedTransactions struct {
	WsModifiedTransactions []*WsModifiedTransactions `xml:"wsModifiedTransactions,omitempty" json:"wsModifiedTransactions,omitempty"`
}

type WsModifiedTransactions struct {
	XMLName                     xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsModifiedTransactions"`
	ModifyDate                  GMTDate  `xml:"ModifyDate,omitempty" json:"ModifyDate,omitempty"`
	Password                    string   `xml:"Password,omitempty" json:"Password,omitempty"`
	Receipt                     string   `xml:"Receipt,omitempty" json:"Receipt,omitempty"`
	ReceiverAccount             string   `xml:"ReceiverAccount,omitempty" json:"ReceiverAccount,omitempty"`
	ReceiverAccountType         string   `xml:"ReceiverAccountType,omitempty" json:"ReceiverAccountType,omitempty"`
	ReceiverAddress             string   `xml:"ReceiverAddress,omitempty" json:"ReceiverAddress,omitempty"`
	ReceiverBankBranOrRouting   string   `xml:"ReceiverBankBranOrRouting,omitempty" json:"ReceiverBankBranOrRouting,omitempty"`
	ReceiverBankName            string   `xml:"ReceiverBankName,omitempty" json:"ReceiverBankName,omitempty"`
	ReceiverCountryNationallity string   `xml:"ReceiverCountryNationallity,omitempty" json:"ReceiverCountryNationallity,omitempty"`
	ReceiverEmail               string   `xml:"ReceiverEmail,omitempty" json:"ReceiverEmail,omitempty"`
	ReceiverIdNumber            string   `xml:"ReceiverIdNumber,omitempty" json:"ReceiverIdNumber,omitempty"`
	ReceiverIdType              string   `xml:"ReceiverIdType,omitempty" json:"ReceiverIdType,omitempty"`
	ReceiverLastName            string   `xml:"ReceiverLastName,omitempty" json:"ReceiverLastName,omitempty"`
	ReceiverMobile              string   `xml:"ReceiverMobile,omitempty" json:"ReceiverMobile,omitempty"`
	ReceiverName                string   `xml:"ReceiverName,omitempty" json:"ReceiverName,omitempty"`
	ReceiverPOB                 string   `xml:"ReceiverPOB,omitempty" json:"ReceiverPOB,omitempty"`
	ReceiverPhone               string   `xml:"ReceiverPhone,omitempty" json:"ReceiverPhone,omitempty"`
	ReceiverZip                 string   `xml:"ReceiverZip,omitempty" json:"ReceiverZip,omitempty"`
	ThirdPartyReceipt           string   `xml:"ThirdPartyReceipt,omitempty" json:"ThirdPartyReceipt,omitempty"`
}

type ArrayOfwsReleasedTransactions struct {
	WsReleasedTransactions []*WsReleasedTransactions `xml:"wsReleasedTransactions,omitempty" json:"wsReleasedTransactions,omitempty"`
}

type WsReleasedTransactions struct {
	XMLName           xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsReleasedTransactions"`
	HoldDate          GMTDate  `xml:"HoldDate,omitempty" json:"HoldDate,omitempty"`
	Password          string   `xml:"Password,omitempty" json:"Password,omitempty"`
	Receipt           string   `xml:"Receipt,omitempty" json:"Receipt,omitempty"`
	ThirdPartyReceipt string   `xml:"ThirdPartyReceipt,omitempty" json:"ThirdPartyReceipt,omitempty"`
}

type ArrayOfwsCollectedAch struct {
	WsCollectedAch []*WsCollectedAch `xml:"wsCollectedAch,omitempty" json:"wsCollectedAch,omitempty"`
}

type WsCollectedAch struct {
	XMLName           xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsCollectedAch"`
	ClearedDate       GMTDate  `xml:"ClearedDate,omitempty" json:"ClearedDate,omitempty"`
	Password          string   `xml:"Password,omitempty" json:"Password,omitempty"`
	Receipt           string   `xml:"Receipt,omitempty" json:"Receipt,omitempty"`
	ThirdPartyReceipt string   `xml:"ThirdPartyReceipt,omitempty" json:"ThirdPartyReceipt,omitempty"`
}

type ArrayOfwsNotifications struct {
	WsNotifications []*WsNotifications `xml:"wsNotifications,omitempty" json:"wsNotifications,omitempty"`
}

type WsNotifications struct {
	XMLName           xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsNotifications"`
	Password          string   `xml:"Password,omitempty" json:"Password,omitempty"`
	Receipt           string   `xml:"Receipt,omitempty" json:"Receipt,omitempty"`
	Status            string   `xml:"Status,omitempty" json:"Status,omitempty"`
	StatusDate        GMTDate  `xml:"StatusDate,omitempty" json:"StatusDate,omitempty"`
	ThirdPartyReceipt string   `xml:"ThirdPartyReceipt,omitempty" json:"ThirdPartyReceipt,omitempty"`
	WireDate          GMTDate  `xml:"WireDate,omitempty" json:"WireDate,omitempty"`
}

type WsAchStatus struct {
	XMLName  xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsAchStatus"`
	Amount   *float64 `xml:"Amount,omitempty" json:"Amount,omitempty"`
	Name     string   `xml:"Name,omitempty" json:"Name,omitempty"`
	Password string   `xml:"Password,omitempty" json:"Password,omitempty"`
	Status   string   `xml:"Status,omitempty" json:"Status,omitempty"`
}

type WsStatus struct {
	XMLName           xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsStatus"`
	LastComment       string   `xml:"LastComment,omitempty" json:"LastComment,omitempty"`
	PartnerReference  string   `xml:"PartnerReference,omitempty" json:"PartnerReference,omitempty"`
	Password          string   `xml:"Password,omitempty" json:"Password,omitempty"`
	PayMethod         string   `xml:"PayMethod,omitempty" json:"PayMethod,omitempty"`
	Receipt           string   `xml:"Receipt,omitempty" json:"Receipt,omitempty"`
	Status            string   `xml:"Status,omitempty" json:"Status,omitempty"`
	ThirdPartyReceipt string   `xml:"ThirdPartyReceipt,omitempty" json:"ThirdPartyReceipt,omitempty"`
}

type WsExRate struct {
	XMLName xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsExRate"`
	ExRate  *float64 `xml:"ExRate,omitempty" json:"ExRate,omitempty"`
}

type ArrayOfwsExRateList struct {
	WsExRateList []*WsExRateList `xml:"wsExRateList,omitempty" json:"wsExRateList,omitempty"`
}

type WsExRateList struct {
	XMLName  xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsExRateList"`
	ExRate   *float64 `xml:"ExRate,omitempty" json:"ExRate,omitempty"`
	Cantup   float64  `xml:"cantup,omitempty" json:"cantup,omitempty"`
	Code     string   `xml:"code,omitempty" json:"code,omitempty"`
	Country  string   `xml:"country,omitempty" json:"country,omitempty"`
	Currency string   `xml:"currency,omitempty" json:"currency,omitempty"`
	Payer    string   `xml:"payer,omitempty" json:"payer,omitempty"`
}

type WsOfac struct {
	XMLName         xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsOfac"`
	Error           int32    `xml:"Error,omitempty" json:"Error,omitempty"`
	Message         string   `xml:"Message,omitempty" json:"Message,omitempty"`
	OfacAddress     string   `xml:"OfacAddress,omitempty" json:"OfacAddress,omitempty"`
	OfacAka         string   `xml:"OfacAka,omitempty" json:"OfacAka,omitempty"`
	OfacCitizenship string   `xml:"OfacCitizenship,omitempty" json:"OfacCitizenship,omitempty"`
	OfacComment     string   `xml:"OfacComment,omitempty" json:"OfacComment,omitempty"`
	OfacDob         string   `xml:"OfacDob,omitempty" json:"OfacDob,omitempty"`
	OfacFirst       string   `xml:"OfacFirst,omitempty" json:"OfacFirst,omitempty"`
	OfacIds         string   `xml:"OfacIds,omitempty" json:"OfacIds,omitempty"`
	OfacLast        string   `xml:"OfacLast,omitempty" json:"OfacLast,omitempty"`
	OfacName        string   `xml:"OfacName,omitempty" json:"OfacName,omitempty"`
	OfacNationality string   `xml:"OfacNationality,omitempty" json:"OfacNationality,omitempty"`
	OfacPob         string   `xml:"OfacPob,omitempty" json:"OfacPob,omitempty"`
	OfacProgram     string   `xml:"OfacProgram,omitempty" json:"OfacProgram,omitempty"`
	OfacRemark      string   `xml:"OfacRemark,omitempty" json:"OfacRemark,omitempty"`
	OfacScore       float64  `xml:"OfacScore,omitempty" json:"OfacScore,omitempty"`
	OfacSource      string   `xml:"OfacSource,omitempty" json:"OfacSource,omitempty"`
	OfacTitle       string   `xml:"OfacTitle,omitempty" json:"OfacTitle,omitempty"`
	OfacType        string   `xml:"OfacType,omitempty" json:"OfacType,omitempty"`
	Valid           bool     `xml:"Valid,omitempty" json:"Valid,omitempty"`
}

type WsResult struct {
	XMLName xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsResult"`
	Error   int32    `xml:"Error,omitempty" json:"Error,omitempty"`
	Message string   `xml:"Message,omitempty" json:"Message,omitempty"`
	Valid   bool     `xml:"Valid,omitempty" json:"Valid,omitempty"`
}

type ArrayOfws_Select_PayersByCountryCodeResult struct {
	Ws_Select_PayersByCountryCodeResult []*Ws_Select_PayersByCountryCodeResult `xml:"ws_Select_PayersByCountryCodeResult,omitempty" json:"ws_Select_PayersByCountryCodeResult,omitempty"`
}

type Ws_Select_PayersByCountryCodeResult struct {
	XMLName           xml.Name                                  `xml:"http://schemas.datacontract.org/2004/07/gmtpay ws_Select_PayersByCountryCodeResult"`
	NeedsBank         bool                                      `xml:"NeedsBank,omitempty" json:"NeedsBank,omitempty"`
	NeedsBenID        bool                                      `xml:"NeedsBenID,omitempty" json:"NeedsBenID,omitempty"`
	NeedsBranch       bool                                      `xml:"NeedsBranch,omitempty" json:"NeedsBranch,omitempty"`
	NeedsIBAN         bool                                      `xml:"NeedsIBAN,omitempty" json:"NeedsIBAN,omitempty"`
	PayerAccountTypes *ArrayOfws_Select_PayerAccountTypesResult `xml:"PayerAccountTypes,omitempty" json:"PayerAccountTypes,omitempty"`
	PayerBanks        *ArrayOfws_Select_PayerBanksResult        `xml:"PayerBanks,omitempty" json:"PayerBanks,omitempty"`
	PayerCode         string                                    `xml:"PayerCode,omitempty" json:"PayerCode,omitempty"`
	PayerCountryName  string                                    `xml:"PayerCountryName,omitempty" json:"PayerCountryName,omitempty"`
	PayerCurrencies   *ArrayOfws_Select_PayerCurrenciesResult   `xml:"PayerCurrencies,omitempty" json:"PayerCurrencies,omitempty"`
	PayerLimits       *ArrayOfws_Select_PayerLimitsResult       `xml:"PayerLimits,omitempty" json:"PayerLimits,omitempty"`
	PayerName         string                                    `xml:"PayerName,omitempty" json:"PayerName,omitempty"`
	PayerOffices      *ArrayOfws_Select_PayerOfficesResult      `xml:"PayerOffices,omitempty" json:"PayerOffices,omitempty"`
	PayerServices     *ArrayOfws_Select_PayerServicesResult     `xml:"PayerServices,omitempty" json:"PayerServices,omitempty"`
	PayerType         string                                    `xml:"PayerType,omitempty" json:"PayerType,omitempty"`
}

type ArrayOfws_Select_PayerAccountTypesResult struct {
	Ws_Select_PayerAccountTypesResult []*Ws_Select_PayerAccountTypesResult `xml:"ws_Select_PayerAccountTypesResult,omitempty" json:"ws_Select_PayerAccountTypesResult,omitempty"`
}

type Ws_Select_PayerAccountTypesResult struct {
	XMLName         xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay ws_Select_PayerAccountTypesResult"`
	AccountTypeCode string   `xml:"AccountTypeCode,omitempty" json:"AccountTypeCode,omitempty"`
	AccountTypeName string   `xml:"AccountTypeName,omitempty" json:"AccountTypeName,omitempty"`
}

type ArrayOfws_Select_PayerBanksResult struct {
	Ws_Select_PayerBanksResult []*Ws_Select_PayerBanksResult `xml:"ws_Select_PayerBanksResult,omitempty" json:"ws_Select_PayerBanksResult,omitempty"`
}

type Ws_Select_PayerBanksResult struct {
	XMLName  xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay ws_Select_PayerBanksResult"`
	BankCode string   `xml:"BankCode,omitempty" json:"BankCode,omitempty"`
	BankName string   `xml:"BankName,omitempty" json:"BankName,omitempty"`
}

type ArrayOfws_Select_PayerCurrenciesResult struct {
	Ws_Select_PayerCurrenciesResult []*Ws_Select_PayerCurrenciesResult `xml:"ws_Select_PayerCurrenciesResult,omitempty" json:"ws_Select_PayerCurrenciesResult,omitempty"`
}

type Ws_Select_PayerCurrenciesResult struct {
	XMLName      xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay ws_Select_PayerCurrenciesResult"`
	CurrencyCode string   `xml:"CurrencyCode,omitempty" json:"CurrencyCode,omitempty"`
	CurrencyName string   `xml:"CurrencyName,omitempty" json:"CurrencyName,omitempty"`
}

type ArrayOfws_Select_PayerLimitsResult struct {
	Ws_Select_PayerLimitsResult []*Ws_Select_PayerLimitsResult `xml:"ws_Select_PayerLimitsResult,omitempty" json:"ws_Select_PayerLimitsResult,omitempty"`
}

type Ws_Select_PayerLimitsResult struct {
	XMLName         xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay ws_Select_PayerLimitsResult"`
	Amount          float64  `xml:"Amount,omitempty" json:"Amount,omitempty"`
	CurrencyCode    string   `xml:"CurrencyCode,omitempty" json:"CurrencyCode,omitempty"`
	Days            int32    `xml:"Days,omitempty" json:"Days,omitempty"`
	OfficeGroupName string   `xml:"OfficeGroupName,omitempty" json:"OfficeGroupName,omitempty"`
	ServiceCode     string   `xml:"ServiceCode,omitempty" json:"ServiceCode,omitempty"`
	Type            string   `xml:"Type,omitempty" json:"Type,omitempty"`
}

type ArrayOfws_Select_PayerOfficesResult struct {
	Ws_Select_PayerOfficesResult []*Ws_Select_PayerOfficesResult `xml:"ws_Select_PayerOfficesResult,omitempty" json:"ws_Select_PayerOfficesResult,omitempty"`
}

type Ws_Select_PayerOfficesResult struct {
	XMLName         xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay ws_Select_PayerOfficesResult"`
	City            string   `xml:"City,omitempty" json:"City,omitempty"`
	CountryCode     string   `xml:"CountryCode,omitempty" json:"CountryCode,omitempty"`
	OfficeCode      string   `xml:"OfficeCode,omitempty" json:"OfficeCode,omitempty"`
	OfficeGroup     int32    `xml:"OfficeGroup,omitempty" json:"OfficeGroup,omitempty"`
	OfficeGroupName string   `xml:"OfficeGroupName,omitempty" json:"OfficeGroupName,omitempty"`
	State           string   `xml:"State,omitempty" json:"State,omitempty"`
}

type ArrayOfws_Select_PayerServicesResult struct {
	Ws_Select_PayerServicesResult []*Ws_Select_PayerServicesResult `xml:"ws_Select_PayerServicesResult,omitempty" json:"ws_Select_PayerServicesResult,omitempty"`
}

type Ws_Select_PayerServicesResult struct {
	XMLName     xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay ws_Select_PayerServicesResult"`
	ServiceCode string   `xml:"ServiceCode,omitempty" json:"ServiceCode,omitempty"`
	ServiceName string   `xml:"ServiceName,omitempty" json:"ServiceName,omitempty"`
}

type ReceiptTemplate struct {
	Rec_contact    string `xml:"rec_contact,omitempty" json:"rec_contact,omitempty"`
	Rec_contact_en string `xml:"rec_contact_en,omitempty" json:"rec_contact_en,omitempty"`
	Rec_error      string `xml:"rec_error,omitempty" json:"rec_error,omitempty"`
	Rec_error_en   string `xml:"rec_error_en,omitempty" json:"rec_error_en,omitempty"`
	Rec_idioma_id  int32  `xml:"rec_idioma_id,omitempty" json:"rec_idioma_id,omitempty"`
	Rec_license    string `xml:"rec_license,omitempty" json:"rec_license,omitempty"`
	Rec_pdf        string `xml:"rec_pdf,omitempty" json:"rec_pdf,omitempty"`
	Rec_rtr        string `xml:"rec_rtr,omitempty" json:"rec_rtr,omitempty"`
	Rec_rtr_en     string `xml:"rec_rtr_en,omitempty" json:"rec_rtr_en,omitempty"`
}

type CityByZip struct {
	CioCountryCode string `xml:"CioCountryCode,omitempty" json:"CioCountryCode,omitempty"`
	CioName        string `xml:"CioName,omitempty" json:"CioName,omitempty"`
	CioState       string `xml:"CioState,omitempty" json:"CioState,omitempty"`
	CioStateCode   string `xml:"CioStateCode,omitempty" json:"CioStateCode,omitempty"`
}

type Ws_LimitByPayerResult struct {
	XMLName        xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay ws_LimitByPayerResult"`
	Amount         *float64 `xml:"Amount,omitempty" json:"Amount,omitempty"`
	ContinueSaving *bool    `xml:"ContinueSaving,omitempty" json:"ContinueSaving,omitempty"`
	Limit          *float64 `xml:"Limit,omitempty" json:"Limit,omitempty"`
	Message        string   `xml:"Message,omitempty" json:"Message,omitempty"`
}

type Ws_Select_PromoResult struct {
	XMLName  xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay ws_Select_PromoResult"`
	MSG      string   `xml:"MSG,omitempty" json:"MSG,omitempty"`
	PRO_CANT *float64 `xml:"PRO_CANT,omitempty" json:"PRO_CANT,omitempty"`
	PRO_CODE string   `xml:"PRO_CODE,omitempty" json:"PRO_CODE,omitempty"`
	PRO_DESC *float64 `xml:"PRO_DESC,omitempty" json:"PRO_DESC,omitempty"`
	PRO_FEE  *float64 `xml:"PRO_FEE,omitempty" json:"PRO_FEE,omitempty"`
}

type WsWallet struct {
	XMLName    xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsWallet"`
	Balance    float64  `xml:"Balance,omitempty" json:"Balance,omitempty"`
	Error      int32    `xml:"Error,omitempty" json:"Error,omitempty"`
	ExternalId string   `xml:"ExternalId,omitempty" json:"ExternalId,omitempty"`
	Message    string   `xml:"Message,omitempty" json:"Message,omitempty"`
	SenderId   int32    `xml:"SenderId,omitempty" json:"SenderId,omitempty"`
}

type Wallet_Operation struct {
	Amount           float64 `xml:"Amount,omitempty" json:"Amount,omitempty"`
	External_ID      string  `xml:"External_ID,omitempty" json:"External_ID,omitempty"`
	SenderAchAccount string  `xml:"SenderAchAccount,omitempty" json:"SenderAchAccount,omitempty"`
	SenderAchRouting string  `xml:"SenderAchRouting,omitempty" json:"SenderAchRouting,omitempty"`
	SenderAchType    string  `xml:"SenderAchType,omitempty" json:"SenderAchType,omitempty"`
	SenderId         int32   `xml:"SenderId,omitempty" json:"SenderId,omitempty"`
}

type WsDocument struct {
	XMLName      xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsDocument"`
	Comments     string   `xml:"Comments,omitempty" json:"Comments,omitempty"`
	DocumentType string   `xml:"DocumentType,omitempty" json:"DocumentType,omitempty"`
	FileContent  *[]byte  `xml:"FileContent,omitempty" json:"FileContent,omitempty"`
	Filename     string   `xml:"Filename,omitempty" json:"Filename,omitempty"`
	Password     string   `xml:"Password,omitempty" json:"Password,omitempty"`
	SenderId     string   `xml:"SenderId,omitempty" json:"SenderId,omitempty"`
}

type ArrayOfwsOccupations struct {
	WsOccupations []*WsOccupations `xml:"wsOccupations,omitempty" json:"wsOccupations,omitempty"`
}

type WsOccupations struct {
	XMLName  xml.Name `xml:"http://schemas.datacontract.org/2004/07/gmtpay wsOccupations"`
	Job_Name string   `xml:"Job_Name,omitempty" json:"Job_Name,omitempty"`
}

type Service interface {
	InsertTransaction(request *InsertTransaction) (*InsertTransactionResponse, error)

	InsertTransactionContext(ctx context.Context, request *InsertTransaction) (*InsertTransactionResponse, error)

	GetPaidTransactions(request *GetPaidTransactions) (*GetPaidTransactionsResponse, error)

	GetPaidTransactionsContext(ctx context.Context, request *GetPaidTransactions) (*GetPaidTransactionsResponse, error)

	ConfirmPayment(request *ConfirmPayment) (*ConfirmPaymentResponse, error)

	ConfirmPaymentContext(ctx context.Context, request *ConfirmPayment) (*ConfirmPaymentResponse, error)

	RequestCancellation(request *RequestCancellation) (*RequestCancellationResponse, error)

	RequestCancellationContext(ctx context.Context, request *RequestCancellation) (*RequestCancellationResponse, error)

	GetCancelledTransactions(request *GetCancelledTransactions) (*GetCancelledTransactionsResponse, error)

	GetCancelledTransactionsContext(ctx context.Context, request *GetCancelledTransactions) (*GetCancelledTransactionsResponse, error)

	ConfirmCancellation(request *ConfirmCancellation) (*ConfirmCancellationResponse, error)

	ConfirmCancellationContext(ctx context.Context, request *ConfirmCancellation) (*ConfirmCancellationResponse, error)

	RequestModification(request *RequestModification) (*RequestModificationResponse, error)

	RequestModificationContext(ctx context.Context, request *RequestModification) (*RequestModificationResponse, error)

	GetModifiedTransactions(request *GetModifiedTransactions) (*GetModifiedTransactionsResponse, error)

	GetModifiedTransactionsContext(ctx context.Context, request *GetModifiedTransactions) (*GetModifiedTransactionsResponse, error)

	ConfirmModification(request *ConfirmModification) (*ConfirmModificationResponse, error)

	ConfirmModificationContext(ctx context.Context, request *ConfirmModification) (*ConfirmModificationResponse, error)

	GetReleasedTransactions(request *GetReleasedTransactions) (*GetReleasedTransactionsResponse, error)

	GetReleasedTransactionsContext(ctx context.Context, request *GetReleasedTransactions) (*GetReleasedTransactionsResponse, error)

	ConfirmRelease(request *ConfirmRelease) (*ConfirmReleaseResponse, error)

	ConfirmReleaseContext(ctx context.Context, request *ConfirmRelease) (*ConfirmReleaseResponse, error)

	GetClearedAchTransactions(request *GetClearedAchTransactions) (*GetClearedAchTransactionsResponse, error)

	GetClearedAchTransactionsContext(ctx context.Context, request *GetClearedAchTransactions) (*GetClearedAchTransactionsResponse, error)

	ConfirmCollection(request *ConfirmCollection) (*ConfirmCollectionResponse, error)

	ConfirmCollectionContext(ctx context.Context, request *ConfirmCollection) (*ConfirmCollectionResponse, error)

	GetNotifications(request *GetNotifications) (*GetNotificationsResponse, error)

	GetNotificationsContext(ctx context.Context, request *GetNotifications) (*GetNotificationsResponse, error)

	GetAchStatus(request *GetAchStatus) (*GetAchStatusResponse, error)

	GetAchStatusContext(ctx context.Context, request *GetAchStatus) (*GetAchStatusResponse, error)

	GetTransactionStatus(request *GetTransactionStatus) (*GetTransactionStatusResponse, error)

	GetTransactionStatusContext(ctx context.Context, request *GetTransactionStatus) (*GetTransactionStatusResponse, error)

	GetSingleExchangeRate(request *GetSingleExchangeRate) (*GetSingleExchangeRateResponse, error)

	GetSingleExchangeRateContext(ctx context.Context, request *GetSingleExchangeRate) (*GetSingleExchangeRateResponse, error)

	GetExchangeRates(request *GetExchangeRates) (*GetExchangeRatesResponse, error)

	GetExchangeRatesContext(ctx context.Context, request *GetExchangeRates) (*GetExchangeRatesResponse, error)

	OfacVerification(request *OfacVerification) (*OfacVerificationResponse, error)

	OfacVerificationContext(ctx context.Context, request *OfacVerification) (*OfacVerificationResponse, error)

	ComplianceCheck(request *ComplianceCheck) (*ComplianceCheckResponse, error)

	ComplianceCheckContext(ctx context.Context, request *ComplianceCheck) (*ComplianceCheckResponse, error)

	SetVerified(request *SetVerified) (*SetVerifiedResponse, error)

	SetVerifiedContext(ctx context.Context, request *SetVerified) (*SetVerifiedResponse, error)

	PayersConsult(request *PayersConsult) (*PayersConsultResponse, error)

	PayersConsultContext(ctx context.Context, request *PayersConsult) (*PayersConsultResponse, error)

	GetReceiptData(request *GetReceiptData) (*GetReceiptDataResponse, error)

	GetReceiptDataContext(ctx context.Context, request *GetReceiptData) (*GetReceiptDataResponse, error)

	GetCityByZip(request *GetCityByZip) (*GetCityByZipResponse, error)

	GetCityByZipContext(ctx context.Context, request *GetCityByZip) (*GetCityByZipResponse, error)

	CheckPayerLimits(request *CheckPayerLimits) (*CheckPayerLimitsResponse, error)

	CheckPayerLimitsContext(ctx context.Context, request *CheckPayerLimits) (*CheckPayerLimitsResponse, error)

	PromotionsCode(request *PromotionsCode) (*PromotionsCodeResponse, error)

	PromotionsCodeContext(ctx context.Context, request *PromotionsCode) (*PromotionsCodeResponse, error)

	RegisterWallet(request *RegisterWallet) (*RegisterWalletResponse, error)

	RegisterWalletContext(ctx context.Context, request *RegisterWallet) (*RegisterWalletResponse, error)

	AddWalletFunds(request *AddWalletFunds) (*AddWalletFundsResponse, error)

	AddWalletFundsContext(ctx context.Context, request *AddWalletFunds) (*AddWalletFundsResponse, error)

	WithdrawWalletFunds(request *WithdrawWalletFunds) (*WithdrawWalletFundsResponse, error)

	WithdrawWalletFundsContext(ctx context.Context, request *WithdrawWalletFunds) (*WithdrawWalletFundsResponse, error)

	GetWalletBalance(request *GetWalletBalance) (*GetWalletBalanceResponse, error)

	GetWalletBalanceContext(ctx context.Context, request *GetWalletBalance) (*GetWalletBalanceResponse, error)

	AddDocument(request *AddDocument) (*AddDocumentResponse, error)

	AddDocumentContext(ctx context.Context, request *AddDocument) (*AddDocumentResponse, error)

	GetOccupations(request *GetOccupations) (*GetOccupationsResponse, error)

	GetOccupationsContext(ctx context.Context, request *GetOccupations) (*GetOccupationsResponse, error)
}

type iService1 struct {
	client *soap.Client
}

func (service *iService1) InsertTransactionContext(ctx context.Context, request *InsertTransaction) (*InsertTransactionResponse, error) {
	response := new(InsertTransactionResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/InsertTransaction", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) InsertTransaction(request *InsertTransaction) (*InsertTransactionResponse, error) {
	return service.InsertTransactionContext(
		context.Background(),
		request,
	)
}

func (service *iService1) GetPaidTransactionsContext(ctx context.Context, request *GetPaidTransactions) (*GetPaidTransactionsResponse, error) {
	response := new(GetPaidTransactionsResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/GetPaidTransactions", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) GetPaidTransactions(request *GetPaidTransactions) (*GetPaidTransactionsResponse, error) {
	return service.GetPaidTransactionsContext(
		context.Background(),
		request,
	)
}

func (service *iService1) ConfirmPaymentContext(ctx context.Context, request *ConfirmPayment) (*ConfirmPaymentResponse, error) {
	response := new(ConfirmPaymentResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/ConfirmPayment", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) ConfirmPayment(request *ConfirmPayment) (*ConfirmPaymentResponse, error) {
	return service.ConfirmPaymentContext(
		context.Background(),
		request,
	)
}

func (service *iService1) RequestCancellationContext(ctx context.Context, request *RequestCancellation) (*RequestCancellationResponse, error) {
	response := new(RequestCancellationResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/RequestCancellation", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) RequestCancellation(request *RequestCancellation) (*RequestCancellationResponse, error) {
	return service.RequestCancellationContext(
		context.Background(),
		request,
	)
}

func (service *iService1) GetCancelledTransactionsContext(ctx context.Context, request *GetCancelledTransactions) (*GetCancelledTransactionsResponse, error) {
	response := new(GetCancelledTransactionsResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/GetCancelledTransactions", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) GetCancelledTransactions(request *GetCancelledTransactions) (*GetCancelledTransactionsResponse, error) {
	return service.GetCancelledTransactionsContext(
		context.Background(),
		request,
	)
}

func (service *iService1) ConfirmCancellationContext(ctx context.Context, request *ConfirmCancellation) (*ConfirmCancellationResponse, error) {
	response := new(ConfirmCancellationResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/ConfirmCancellation", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) ConfirmCancellation(request *ConfirmCancellation) (*ConfirmCancellationResponse, error) {
	return service.ConfirmCancellationContext(
		context.Background(),
		request,
	)
}

func (service *iService1) RequestModificationContext(ctx context.Context, request *RequestModification) (*RequestModificationResponse, error) {
	response := new(RequestModificationResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/RequestModification", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) RequestModification(request *RequestModification) (*RequestModificationResponse, error) {
	return service.RequestModificationContext(
		context.Background(),
		request,
	)
}

func (service *iService1) GetModifiedTransactionsContext(ctx context.Context, request *GetModifiedTransactions) (*GetModifiedTransactionsResponse, error) {
	response := new(GetModifiedTransactionsResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/GetModifiedTransactions", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) GetModifiedTransactions(request *GetModifiedTransactions) (*GetModifiedTransactionsResponse, error) {
	return service.GetModifiedTransactionsContext(
		context.Background(),
		request,
	)
}

func (service *iService1) ConfirmModificationContext(ctx context.Context, request *ConfirmModification) (*ConfirmModificationResponse, error) {
	response := new(ConfirmModificationResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/ConfirmModification", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) ConfirmModification(request *ConfirmModification) (*ConfirmModificationResponse, error) {
	return service.ConfirmModificationContext(
		context.Background(),
		request,
	)
}

func (service *iService1) GetReleasedTransactionsContext(ctx context.Context, request *GetReleasedTransactions) (*GetReleasedTransactionsResponse, error) {
	response := new(GetReleasedTransactionsResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/GetReleasedTransactions", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) GetReleasedTransactions(request *GetReleasedTransactions) (*GetReleasedTransactionsResponse, error) {
	return service.GetReleasedTransactionsContext(
		context.Background(),
		request,
	)
}

func (service *iService1) ConfirmReleaseContext(ctx context.Context, request *ConfirmRelease) (*ConfirmReleaseResponse, error) {
	response := new(ConfirmReleaseResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/ConfirmRelease", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) ConfirmRelease(request *ConfirmRelease) (*ConfirmReleaseResponse, error) {
	return service.ConfirmReleaseContext(
		context.Background(),
		request,
	)
}

func (service *iService1) GetClearedAchTransactionsContext(ctx context.Context, request *GetClearedAchTransactions) (*GetClearedAchTransactionsResponse, error) {
	response := new(GetClearedAchTransactionsResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/GetClearedAchTransactions", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) GetClearedAchTransactions(request *GetClearedAchTransactions) (*GetClearedAchTransactionsResponse, error) {
	return service.GetClearedAchTransactionsContext(
		context.Background(),
		request,
	)
}

func (service *iService1) ConfirmCollectionContext(ctx context.Context, request *ConfirmCollection) (*ConfirmCollectionResponse, error) {
	response := new(ConfirmCollectionResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/ConfirmCollection", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) ConfirmCollection(request *ConfirmCollection) (*ConfirmCollectionResponse, error) {
	return service.ConfirmCollectionContext(
		context.Background(),
		request,
	)
}

func (service *iService1) GetNotificationsContext(ctx context.Context, request *GetNotifications) (*GetNotificationsResponse, error) {
	response := new(GetNotificationsResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/GetNotifications", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) GetNotifications(request *GetNotifications) (*GetNotificationsResponse, error) {
	return service.GetNotificationsContext(
		context.Background(),
		request,
	)
}

func (service *iService1) GetAchStatusContext(ctx context.Context, request *GetAchStatus) (*GetAchStatusResponse, error) {
	response := new(GetAchStatusResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/GetAchStatus", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) GetAchStatus(request *GetAchStatus) (*GetAchStatusResponse, error) {
	return service.GetAchStatusContext(
		context.Background(),
		request,
	)
}

func (service *iService1) GetTransactionStatusContext(ctx context.Context, request *GetTransactionStatus) (*GetTransactionStatusResponse, error) {
	response := new(GetTransactionStatusResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/GetTransactionStatus", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) GetTransactionStatus(request *GetTransactionStatus) (*GetTransactionStatusResponse, error) {
	return service.GetTransactionStatusContext(
		context.Background(),
		request,
	)
}

func (service *iService1) GetSingleExchangeRateContext(ctx context.Context, request *GetSingleExchangeRate) (*GetSingleExchangeRateResponse, error) {
	response := new(GetSingleExchangeRateResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/GetSingleExchangeRate", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) GetSingleExchangeRate(request *GetSingleExchangeRate) (*GetSingleExchangeRateResponse, error) {
	return service.GetSingleExchangeRateContext(
		context.Background(),
		request,
	)
}

func (service *iService1) GetExchangeRatesContext(ctx context.Context, request *GetExchangeRates) (*GetExchangeRatesResponse, error) {
	response := new(GetExchangeRatesResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/GetExchangeRates", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) GetExchangeRates(request *GetExchangeRates) (*GetExchangeRatesResponse, error) {
	return service.GetExchangeRatesContext(
		context.Background(),
		request,
	)
}

func (service *iService1) OfacVerificationContext(ctx context.Context, request *OfacVerification) (*OfacVerificationResponse, error) {
	response := new(OfacVerificationResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/OfacVerification", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) OfacVerification(request *OfacVerification) (*OfacVerificationResponse, error) {
	return service.OfacVerificationContext(
		context.Background(),
		request,
	)
}

func (service *iService1) ComplianceCheckContext(ctx context.Context, request *ComplianceCheck) (*ComplianceCheckResponse, error) {
	response := new(ComplianceCheckResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/ComplianceCheck", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) ComplianceCheck(request *ComplianceCheck) (*ComplianceCheckResponse, error) {
	return service.ComplianceCheckContext(
		context.Background(),
		request,
	)
}

func (service *iService1) SetVerifiedContext(ctx context.Context, request *SetVerified) (*SetVerifiedResponse, error) {
	response := new(SetVerifiedResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/SetVerified", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) SetVerified(request *SetVerified) (*SetVerifiedResponse, error) {
	return service.SetVerifiedContext(
		context.Background(),
		request,
	)
}

func (service *iService1) PayersConsultContext(ctx context.Context, request *PayersConsult) (*PayersConsultResponse, error) {
	response := new(PayersConsultResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/PayersConsult", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) PayersConsult(request *PayersConsult) (*PayersConsultResponse, error) {
	return service.PayersConsultContext(
		context.Background(),
		request,
	)
}

func (service *iService1) GetReceiptDataContext(ctx context.Context, request *GetReceiptData) (*GetReceiptDataResponse, error) {
	response := new(GetReceiptDataResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/GetReceiptData", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) GetReceiptData(request *GetReceiptData) (*GetReceiptDataResponse, error) {
	return service.GetReceiptDataContext(
		context.Background(),
		request,
	)
}

func (service *iService1) GetCityByZipContext(ctx context.Context, request *GetCityByZip) (*GetCityByZipResponse, error) {
	response := new(GetCityByZipResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/GetCityByZip", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) GetCityByZip(request *GetCityByZip) (*GetCityByZipResponse, error) {
	return service.GetCityByZipContext(
		context.Background(),
		request,
	)
}

func (service *iService1) CheckPayerLimitsContext(ctx context.Context, request *CheckPayerLimits) (*CheckPayerLimitsResponse, error) {
	response := new(CheckPayerLimitsResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/CheckPayerLimits", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) CheckPayerLimits(request *CheckPayerLimits) (*CheckPayerLimitsResponse, error) {
	return service.CheckPayerLimitsContext(
		context.Background(),
		request,
	)
}

func (service *iService1) PromotionsCodeContext(ctx context.Context, request *PromotionsCode) (*PromotionsCodeResponse, error) {
	response := new(PromotionsCodeResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/PromotionsCode", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) PromotionsCode(request *PromotionsCode) (*PromotionsCodeResponse, error) {
	return service.PromotionsCodeContext(
		context.Background(),
		request,
	)
}

func (service *iService1) RegisterWalletContext(ctx context.Context, request *RegisterWallet) (*RegisterWalletResponse, error) {
	response := new(RegisterWalletResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/RegisterWallet", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) RegisterWallet(request *RegisterWallet) (*RegisterWalletResponse, error) {
	return service.RegisterWalletContext(
		context.Background(),
		request,
	)
}

func (service *iService1) AddWalletFundsContext(ctx context.Context, request *AddWalletFunds) (*AddWalletFundsResponse, error) {
	response := new(AddWalletFundsResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/AddWalletFunds", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) AddWalletFunds(request *AddWalletFunds) (*AddWalletFundsResponse, error) {
	return service.AddWalletFundsContext(
		context.Background(),
		request,
	)
}

func (service *iService1) WithdrawWalletFundsContext(ctx context.Context, request *WithdrawWalletFunds) (*WithdrawWalletFundsResponse, error) {
	response := new(WithdrawWalletFundsResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/WithdrawWalletFunds", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) WithdrawWalletFunds(request *WithdrawWalletFunds) (*WithdrawWalletFundsResponse, error) {
	return service.WithdrawWalletFundsContext(
		context.Background(),
		request,
	)
}

func (service *iService1) GetWalletBalanceContext(ctx context.Context, request *GetWalletBalance) (*GetWalletBalanceResponse, error) {
	response := new(GetWalletBalanceResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/GetWalletBalance", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) GetWalletBalance(request *GetWalletBalance) (*GetWalletBalanceResponse, error) {
	return service.GetWalletBalanceContext(
		context.Background(),
		request,
	)
}

func (service *iService1) AddDocumentContext(ctx context.Context, request *AddDocument) (*AddDocumentResponse, error) {
	response := new(AddDocumentResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/AddDocument", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) AddDocument(request *AddDocument) (*AddDocumentResponse, error) {
	return service.AddDocumentContext(
		context.Background(),
		request,
	)
}

func (service *iService1) GetOccupationsContext(ctx context.Context, request *GetOccupations) (*GetOccupationsResponse, error) {
	response := new(GetOccupationsResponse)
	err := service.client.CallContext(ctx, "http://tempuri.org/IService1/GetOccupations", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *iService1) GetOccupations(request *GetOccupations) (*GetOccupationsResponse, error) {
	return service.GetOccupationsContext(
		context.Background(),
		request,
	)
}
