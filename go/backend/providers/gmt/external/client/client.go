package client

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"

	"gitlab.com/fynbos/backend/providers/gmt/external"
	mock_client "gitlab.com/fynbos/backend/providers/gmt/external/client/mock"
	httplog "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/env"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type client struct {
	cl       http.Client
	url      string
	alias    string
	user     string
	password string

	txURL      string
	txUser     string
	txPassword string
	txPartner  string
}

func NewClient(transport http.RoundTripper) external.Client {
	if env.IsLocal() {
		return mock_client.SetupDevMock(nil)
	}

	t := transport
	if t == nil {
		t = otelhttp.NewTransport(http.DefaultTransport)
	}

	var alias, user, pass, url, txURL, txUser, txPassword, txPartner string
	if !env.IsProd() {
		alias = "FYN001"
		user = "Fynbos_api"
		pass = "VUJ6bnkxN2dQVXkwMjZaOA=="
		url = "http://35.166.119.115/gmtpay/Service1.svc"

		txURL = "http://35.166.119.115/gmtupd/Service1.svc"
		txUser = "Fynbos_payer"
		txPassword = "ejV1eGZTY0YzMTBG"
		txPartner = "87"
	}

	return &client{
		cl:       http.Client{Transport: t},
		url:      getEnvDefault("GMT_URL", url),
		alias:    getEnvDefault("GMT_ALIAS", alias),
		user:     getEnvDefault("GMT_USER", user),
		password: getEnvDefault("GMT_PASSWORD", pass),

		txURL:      getEnvDefault("GMT_TX_URL", txURL),
		txUser:     getEnvDefault("GMT_TX_USER", txUser),
		txPassword: getEnvDefault("GMT_TX_PASSWORD", txPassword),
		txPartner:  getEnvDefault("GMT_TX_PARTNER", txPartner),
	}
}

func getEnvDefault(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}

	return val
}

func (c *client) InsertTransaction(ctx context.Context, tx external.InsertTransaction) (*external.WsResponse, error) {
	if tx.Sender != nil {
		tx.Sender.XmlNS = "http://schemas.datacontract.org/2004/07/gmtpay"
	}
	if tx.Receiver != nil {
		tx.Receiver.XmlNS = "http://schemas.datacontract.org/2004/07/gmtpay"
	}
	if tx.Transfer != nil {
		tx.Transfer.XmlNS = "http://schemas.datacontract.org/2004/07/gmtpay"
	}
	type txBody struct {
		Text string                             `xml:",chardata"`
		Resp external.InsertTransactionResponse `xml:"InsertTransactionResponse"`
	}
	response := txBody{}

	tx.Alias = c.alias
	tx.User = c.user
	tx.Pass = c.password

	err := c.call(ctx, "http://tempuri.org/IService1/InsertTransaction", tx, &response)
	if err != nil {
		return nil, err
	}

	return response.Resp.InsertTransactionResult, err
}

func (c *client) ComplianceCheck(ctx context.Context, req external.ComplianceCheck) (*external.WsResponse, error) {
	if req.Sender != nil {
		req.Sender.XmlNS = "http://schemas.datacontract.org/2004/07/gmtpay"
	}
	if req.Receiver != nil {
		req.Receiver.XmlNS = "http://schemas.datacontract.org/2004/07/gmtpay"
	}
	if req.Transfer != nil {
		req.Transfer.XmlNS = "http://schemas.datacontract.org/2004/07/gmtpay"
	}
	type cmpBody struct {
		Text string                           `xml:",chardata"`
		Resp external.ComplianceCheckResponse `xml:"ComplianceCheckResponse"`
	}
	response := cmpBody{}

	req.Alias = c.alias
	req.User = c.user
	req.Pass = c.password

	err := c.call(ctx, "http://tempuri.org/IService1/ComplianceCheck", req, &response)
	if err != nil {
		return nil, err
	}

	return response.Resp.ComplianceCheckResult, err
}

func (c *client) OfacVerification(ctx context.Context, req external.OfacVerification) (*external.WsOfac, error) {
	type ofacBody struct {
		Text string                            `xml:",chardata"`
		Resp external.OfacVerificationResponse `xml:"OfacVerificationResponse"`
	}
	response := ofacBody{}

	req.Alias = c.alias
	req.User = c.user
	req.Pass = c.password

	err := c.call(ctx, "http://tempuri.org/IService1/OfacVerification", req, &response)
	if err != nil {
		return nil, err
	}

	return response.Resp.OfacVerificationResult, err
}

func (c *client) SetVerified(ctx context.Context, req external.SetVerified) (*external.WsResult, error) {
	type verBody struct {
		Text string                       `xml:",chardata"`
		Resp external.SetVerifiedResponse `xml:"SetVerifiedResponse"`
	}
	response := verBody{}

	req.Alias = c.alias
	req.User = c.user
	req.Pass = c.password

	err := c.call(ctx, "http://tempuri.org/IService1/SetVerified", req, &response)
	if err != nil {
		return nil, err
	}

	return response.Resp.SetVerifiedResult, err
}

func (c *client) ConfirmCollection(ctx context.Context, req external.ConfirmCollection) (*external.WsResponse, error) {
	type cnfrmBody struct {
		Text string                             `xml:",chardata"`
		Resp external.ConfirmCollectionResponse `xml:"ConfirmCollectionResponse"`
	}
	response := cnfrmBody{}

	req.Alias = c.alias
	req.User = c.user
	req.Pass = c.password

	err := c.call(ctx, "http://tempuri.org/IService1/ConfirmCollection", req, &response)
	if err != nil {
		return nil, err
	}

	return response.Resp.ConfirmCollectionResult, err
}

func (c *client) ConfirmPayment(ctx context.Context, req external.ConfirmPayment) (*external.WsResponse, error) {
	type cnfrmBody struct {
		Text string                          `xml:",chardata"`
		Resp external.ConfirmPaymentResponse `xml:"ConfirmPaymentResponse"`
	}
	response := cnfrmBody{}

	req.Alias = c.alias
	req.User = c.user
	req.Pass = c.password

	err := c.call(ctx, "http://tempuri.org/IService1/ConfirmPayment", req, &response)
	if err != nil {
		return nil, err
	}

	return response.Resp.ConfirmPaymentResult, err
}

func (c *client) GetNotifications(ctx context.Context) ([]*external.WsNotifications, error) {
	type ntfBody struct {
		Text string                            `xml:",chardata"`
		Resp external.GetNotificationsResponse `xml:"GetNotificationsResponse"`
	}
	response := ntfBody{}
	req := external.GetNotifications{
		Alias: c.alias,
		User:  c.user,
		Pass:  c.password,
	}

	err := c.call(ctx, "http://tempuri.org/IService1/GetNotifications", req, &response)
	if err != nil {
		return nil, err
	}

	if response.Resp.GetNotificationsResult == nil || response.Resp.GetNotificationsResult.WsNotifications == nil {
		return []*external.WsNotifications{}, nil
	}

	return response.Resp.GetNotificationsResult.WsNotifications, err
}

func (c *client) UpdateTransactionStatus(ctx context.Context, tx external.UpdateTransactionStatus) (*external.WsResponse, error) {
	type txBody struct {
		Text string                                   `xml:",chardata"`
		Resp external.UpdateTransactionStatusResponse `xml:"UpdateTransactionStatusResponse"`
	}
	response := txBody{}

	tx.User = c.txUser
	tx.Pass = c.txPassword
	tx.Partner = c.txPartner

	err := c.txCall(ctx, "http://tempuri.org/IService1/UpdateTransactionStatus", tx, &response)
	if err != nil {
		return nil, err
	}

	return response.Resp.UpdateTransactionStatusResult, err
}

func (c *client) txCall(ctx context.Context, action string, request, resp interface{}) error {
	return c.httpCall(ctx, action, c.txURL, request, resp)
}

func (c *client) call(ctx context.Context, action string, request, resp interface{}) error {
	return c.httpCall(ctx, action, c.url, request, resp)
}

func (c *client) httpCall(ctx context.Context, action, url string, request, resp interface{}) error {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = action
	}

	envelope := external.SOAPEnvelope{
		XmlNS: "http://schemas.xmlsoap.org/soap/envelope/",
	}

	envelope.Body.Content = request

	payload, err := xml.Marshal(envelope)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "text/xml; charset=\"utf-8\"")
	req.Header.Add("SOAPAction", action)

	res, err := c.cl.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode >= 400 {
		return fmt.Errorf("http error code (%d) msg (%s)", res.StatusCode, body)
	}

	respEnv := &external.SOAPEnvelopeResponse{
		Body: resp,
	}

	err = xml.Unmarshal(body, respEnv)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) RequestCancellation(ctx context.Context, txID, msg string) (*external.WsResponse, error) {
	req := external.RequestCancellation{
		Alias:   c.alias,
		User:    c.user,
		Pass:    c.password,
		Receipt: txID,
		Comment: msg,
	}

	type cancellationBody struct {
		Text string                               `xml:",chardata"`
		Resp external.RequestCancellationResponse `xml:"RequestCancellationResponse"`
	}
	response := cancellationBody{}

	err := c.call(ctx, "http://tempuri.org/IService1/RequestCancellation", req, &response)
	if err != nil {
		return nil, err
	}

	return response.Resp.RequestCancellationResult, nil
}
