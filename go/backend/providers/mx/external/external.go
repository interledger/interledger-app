package external

//go:generate mockgen -destination=./mock.go -package=external -source=./external.go

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	_http "net/http"
)

var (
	ErrInternal = errors.New("mx client: internal")
)

const (
	// Aggregation can be attempted for these.
	CONNECTION_STATUS_CONNECTED    = "CONNECTED"
	CONNECTION_STATUS_CREATED      = "CREATED"
	CONNECTION_STATUS_DEGRADED     = "DEGRADED"
	CONNECTION_STATUS_DISCONNECTED = "DISCONNECTED"
	CONNECTION_STATUS_EXPIRED      = "EXPIRED"
	CONNECTION_STATUS_FAILED       = "FAILED"
	CONNECTION_STATUS_IMPEDED      = "IMPEDED"
	CONNECTION_STATUS_RECONNECTED  = "RECONNECTED"
	CONNECTION_STATUS_UPDATED      = "UPDATED"

	// Aggregation is already ongoing
	CONNECTION_STATUS_CHALLENGED = "CHALLENGED"
	CONNECTION_STATUS_DELAYED    = "DELAYED"
	CONNECTION_STATUS_REJECTED   = "REJECTED"
	CONNECTION_STATUS_RESUMED    = "RESUMED"

	// Update credentials before aggregating
	CONNECTION_STATUS_PREVENTED = "PREVENTED"
	CONNECTION_STATUS_DENIED    = "DENIED"
	CONNECTION_STATUS_IMPAIRED  = "IMPAIRED"
	CONNECTION_STATUS_IMPORTED  = "IMPORTED"

	// Cannot aggregate
	CONNECTION_STATUS_CLOSED       = "CLOSED"
	CONNECTION_STATUS_DISABLED     = "DISABLED"
	CONNECTION_STATUS_DISCONTINUED = "DISCONTINUED"
)

type (
	User struct {
		Guid             string
		ConnectWidgetUrl string `json:"connect_widget_url"`
	}

	Account struct {
		Guid              string
		UserGuid          string `json:"user_guid"`
		MemberGuid        string `json:"member_guid"`
		AccountNumber     string `json:"account_number"`
		InstitutionNumber string `json:"institution_number"`
		RoutingNumber     string `json:"routing_number"`
		TransitNumber     string `json:"transit_number"`
		CurrencyCode      string `json:"currency_code"`
		Type              string
		AvailableBalance  float64 `json:"available_balance"`
		Balance           float64
	}

	AccountNumbers struct {
		AccountGuid string `json:"account_guid"`
		UserGuid    string `json:"user_guid"`
		MemberGuid  string `json:"member_guid"`
	}

	AccountOwner struct {
		AccountGuid string `json:"account_guid"`
		OwnerName   string `json:"owner_name"`
		Country     string
		Email       string
		Phone       string

		// There are more fields (address, state etc.) but we're not using them at the moment.
	}

	Member struct {
		Guid                     string
		UserGuid                 string
		AggregatedAt             string `json:"aggregated_at"`
		IsBeingAggregated        bool   `json:"is_being_aggregated"`
		SuccessfullyAggregatedAt string `json:"successfully_aggregated_at"`
		ConnectionStatus         string `json:"connection_status"`
	}

	Mx interface {
		CreateUser(ctx context.Context) (string, error)
		CheckBalance(ctx context.Context, userGuid, memberGuid string) (*Member, error)
		ReadAccount(ctx context.Context, userGuid, accountGuid string) (*Account, error)
		GetWidgetUrl(ctx context.Context, userGuid string) (string, error)
		GetAccountNumbers(ctx context.Context, userGuid, memberGuid string) ([]AccountNumbers, error)
		GetAccountOwners(ctx context.Context, userGuid, memberGuid string) ([]AccountOwner, error)
		GetMemberStatus(ctx context.Context, userGuid, memberGuid string) (*Member, error)
		AggregateIdentity(ctx context.Context, userGuid, memberGuid string) (*Member, error)
	}

	client struct {
		http    *_http.Client
		baseUrl string
	}
)

type basicAuthTransport struct {
	baseTransport _http.RoundTripper
	userName      string
	password      string
}

// This sets the basic auth credentials on every request.
func (t basicAuthTransport) RoundTrip(r *_http.Request) (*_http.Response, error) {
	r.SetBasicAuth(t.userName, t.password)
	r.Header.Set("Accept", "application/vnd.mx.api.v1+json")
	r.Header.Set("Content-Type", "application/json")
	return t.baseTransport.RoundTrip(r)
}

func newBasicAuthTransport(userName string, password string) *basicAuthTransport {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS12, // pin to tls1.2 as 1.3 doesn't work with MX
	}
	return &basicAuthTransport{
		baseTransport: &_http.Transport{
			TLSClientConfig: tlsConfig,
		},
		userName: userName,
		password: password,
	}
}

func NewClient(baseUrl, clientID, apiKey string) Mx {
	return &client{
		baseUrl: baseUrl,
		http: &_http.Client{
			Transport: newBasicAuthTransport(clientID, apiKey),
		},
	}
}

func (c *client) CreateUser(ctx context.Context) (string, error) {
	payload := `{ "user": { "is_disabled": false } }`
	req, err := _http.NewRequest("POST", fmt.Sprintf("%s/users", c.baseUrl), bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	data := &struct {
		User User
	}{}
	if err = json.Unmarshal(body, data); err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	return data.User.Guid, nil
}

func (c *client) CheckBalance(ctx context.Context, userGuid, memberGuid string) (*Member, error) {
	url := fmt.Sprintf("%s/users/%s/members/%s/check_balance", c.baseUrl, userGuid, memberGuid)
	req, err := _http.NewRequest("GET", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	if resp.StatusCode > 400 {
		return nil, fmt.Errorf("%w Cannot aggregate balance.", ErrInternal)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var member Member
	if err = json.Unmarshal(body, &member); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &member, nil
}

func (c *client) ReadAccount(ctx context.Context, userGuid, accountGuid string) (*Account, error) {
	url := fmt.Sprintf("%s/users/%s/accounts/%s", c.baseUrl, userGuid, accountGuid)
	req, err := _http.NewRequest("GET", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	ret := &struct {
		Account Account
	}{}
	if err = json.Unmarshal(body, ret); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &ret.Account, nil
}

func (c *client) GetWidgetUrl(ctx context.Context, userGuid string) (string, error) {
	url := fmt.Sprintf("%s/users/%s/connect_widget_url", c.baseUrl, userGuid)
	payload := `{
		"config": {
		    "color_scheme": "light",
		    "disable_institution_search": false,
		    "include_transactions": false,
		    "is_mobile_webview": false,
		    "mode": "verification",
		    "ui_message_version": 4,
		    "wait_for_full_aggregation": true,
		    "update_credentials": false
		  }
	}`
	req, err := _http.NewRequest("POST", url, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	data := &struct {
		User User
	}{}
	if err = json.Unmarshal(body, data); err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}
	return data.User.ConnectWidgetUrl, nil
}

func (c *client) GetAccountNumbers(
	ctx context.Context,
	userGuid,
	memberGuid string,
) ([]AccountNumbers, error) {
	url := fmt.Sprintf("%s/users/%s/members/%s/account_numbers?page=1&records_per_page=10", c.baseUrl, userGuid, memberGuid)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	var response struct {
		AccountNumbers []AccountNumbers `json:"account_numbers"`
	}
	if err = json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return response.AccountNumbers, nil
}

func (c *client) GetAccountOwners(ctx context.Context, userGuid, memberGuid string) ([]AccountOwner, error) {
	url := fmt.Sprintf("%s/users/%s/members/%s/account_owners", c.baseUrl, userGuid, memberGuid)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	response := struct {
		AccountOwners []AccountOwner `json:"account_owners"`
	}{}
	if err = json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return response.AccountOwners, nil
}

func (c *client) GetMemberStatus(ctx context.Context, userGuid, memberGuid string) (*Member, error) {
	url := fmt.Sprintf("%s/users/%s/members/%s/status", c.baseUrl, userGuid, memberGuid)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	member := &Member{}
	if err = json.Unmarshal(body, member); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return member, nil
}

func (c *client) AggregateIdentity(ctx context.Context, userGuid, memberGuid string) (*Member, error) {
	url := fmt.Sprintf("%s/users/%s/members/%s/identify", c.baseUrl, userGuid, memberGuid)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	member := &Member{}
	if err = json.Unmarshal(body, member); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return member, nil
}
