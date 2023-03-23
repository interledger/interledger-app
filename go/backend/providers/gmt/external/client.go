package external

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type client struct {
	cl  http.Client
	url string
}

type Client interface {
	InsertTransaction(ctx context.Context, tx InsertTransaction) (*WsResponse, error)
	OfacVerification(ctx context.Context, req OfacVerification) (*WsOfac, error)
	ComplianceCheck(ctx context.Context, req ComplianceCheck) (*WsResponse, error)
	SetVerified(ctx context.Context, req SetVerified) (*WsResult, error)
	ConfirmCollection(ctx context.Context, req ConfirmCollection) (*WsResponse, error)
	GetNotifications(ctx context.Context, req GetNotifications) ([]*WsNotifications, error)
}

func NewClient() Client {
	return &client{url: "http://35.166.119.115/gmtpay/Service1.svc", cl: http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}}
}

func (c *client) InsertTransaction(ctx context.Context, tx InsertTransaction) (*WsResponse, error) {
	type txBody struct {
		Text string                    `xml:",chardata"`
		Resp InsertTransactionResponse `xml:"InsertTransactionResponse"`
	}
	response := txBody{}
	err := c.call(ctx, "http://tempuri.org/IService1/InsertTransaction", tx, response)
	if err != nil {
		return nil, err
	}

	return response.Resp.InsertTransactionResult, err
}

func (c *client) ComplianceCheck(ctx context.Context, req ComplianceCheck) (*WsResponse, error) {
	type cmpBody struct {
		Text string                  `xml:",chardata"`
		Resp ComplianceCheckResponse `xml:"ComplianceCheckResponse"`
	}
	response := cmpBody{}
	err := c.call(ctx, "http://tempuri.org/IService1/ComplianceCheck", req, response)
	if err != nil {
		return nil, err
	}

	return response.Resp.ComplianceCheckResult, err
}

func (c *client) OfacVerification(ctx context.Context, req OfacVerification) (*WsOfac, error) {
	type ofacBody struct {
		Text string                   `xml:",chardata"`
		Resp OfacVerificationResponse `xml:"OfacVerificationResponse"`
	}
	response := ofacBody{}
	err := c.call(ctx, "http://tempuri.org/IService1/OfacVerification", req, &response)
	if err != nil {
		return nil, err
	}

	return response.Resp.OfacVerificationResult, err
}

func (c *client) SetVerified(ctx context.Context, req SetVerified) (*WsResult, error) {
	type verBody struct {
		Text string              `xml:",chardata"`
		Resp SetVerifiedResponse `xml:"SetVerifiedResponse"`
	}
	response := verBody{}
	err := c.call(ctx, "http://tempuri.org/IService1/SetVerified", req, response)
	if err != nil {
		return nil, err
	}

	return response.Resp.SetVerifiedResult, err
}

func (c *client) ConfirmCollection(ctx context.Context, req ConfirmCollection) (*WsResponse, error) {
	type cnfrmBody struct {
		Text string                    `xml:",chardata"`
		Resp ConfirmCollectionResponse `xml:"ConfirmCollectionResponse"`
	}
	response := cnfrmBody{}
	err := c.call(ctx, "http://tempuri.org/IService1/ConfirmCollection", req, response)
	if err != nil {
		return nil, err
	}

	return response.Resp.ConfirmCollectionResult, err
}

func (c *client) GetNotifications(ctx context.Context, req GetNotifications) ([]*WsNotifications, error) {
	type ntfBody struct {
		Text string                   `xml:",chardata"`
		Resp GetNotificationsResponse `xml:"GetNotificationsResponse"`
	}
	response := ntfBody{}
	err := c.call(ctx, "http://tempuri.org/IService1/GetNotifications", req, response)
	if err != nil {
		return nil, err
	}

	if response.Resp.GetNotificationsResult == nil || response.Resp.GetNotificationsResult.WsNotifications == nil {
		return []*WsNotifications{}, nil
	}

	return response.Resp.GetNotificationsResult.WsNotifications, err
}

func (c *client) call(ctx context.Context, action string, request, resp interface{}) error {
	envelope := SOAPEnvelope{
		XmlNS: "http://schemas.xmlsoap.org/soap/envelope/",
	}

	envelope.Body.Content = request

	payload, err := xml.Marshal(envelope)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewBuffer(payload))
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

	respEnv := &SOAPEnvelopeResponse{
		Body: resp,
	}

	err = xml.Unmarshal(body, respEnv)
	if err != nil {
		return err
	}

	return nil
}
