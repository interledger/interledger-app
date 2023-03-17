package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"gitlab.com/fynbos/backend/providers/mx/external"
	"gitlab.com/fynbos/env"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var _ external.Client = &client{}

type client struct {
	baseUrl string
	api     *http.Client
}

func New(clientID, apiKey string, opts ...func(*client)) *client {
	c := &client{
		api: &http.Client{
			Transport: otelhttp.NewTransport(newAuthTransport(clientID, apiKey)),
			Timeout:   5 * time.Second,
		},
	}
	if env.IsProd() {
		c.baseUrl = "https://api.mx.com"
	} else {
		c.baseUrl = "https://int-api.mx.com"
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

type authTransport struct {
	clientID string
	apiKey   string
}

func (t authTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/vnd.mx.api.v1+json")
	r.SetBasicAuth(t.clientID, t.apiKey)
	return http.DefaultTransport.RoundTrip(r)
}

func newAuthTransport(clientID, apiKey string) *authTransport {
	return &authTransport{
		clientID: clientID,
		apiKey:   apiKey,
	}
}

func (c *client) CreateUser(ctx context.Context, id string) (*external.User, error) {
	endpoint, err := url.JoinPath(c.baseUrl, "users")
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	payload := map[string]string{
		"id": id,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	defer resp.Body.Close()

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}

	var user external.User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return &user, nil
}

func (c *client) GetWidgetURL(ctx context.Context, args external.GetWidgetURLArgs) (*external.WidgetURL, error) {
	endpoint, err := url.JoinPath(c.baseUrl, "users", args.UserGuid, "widget_urls")
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	jsonPayload, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	defer resp.Body.Close()

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}

	var widget external.GetWidgerURLResponse
	err = json.NewDecoder(resp.Body).Decode(&widget)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return &widget.WidgetURL, nil
}

func (c *client) ListUsersByID(ctx context.Context, id string) (*external.ListUsersResponse, error) {
	endpoint, err := url.JoinPath(c.baseUrl, "users")
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	defer resp.Body.Close()

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}

	var listResponse external.ListUsersResponse
	err = json.NewDecoder(resp.Body).Decode(&listResponse)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return &listResponse, nil
}

func checkResponseStatusCode(r *http.Response) error {
	if http.StatusOK <= r.StatusCode && r.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	switch r.StatusCode {
	case http.StatusBadRequest:
		return external.ErrBadRequest
	case http.StatusUnauthorized:
		return external.ErrUnauthorized
	case http.StatusForbidden:
		return external.ErrForbidden
	case http.StatusNotFound:
		return external.ErrNotFound
	case http.StatusMethodNotAllowed:
		return external.ErrMethodNotAllowed
	case http.StatusNotAcceptable:
		return external.ErrNotAcceptable
	case http.StatusConflict:
		return external.ErrConflict
	case http.StatusUnprocessableEntity:
		return external.ErrUnprocessableEntity
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusGatewayTimeout:
		return external.ErrServer
	case http.StatusServiceUnavailable:
		return external.ErrServiceUnavailable
	default:
		return fmt.Errorf("%w Unknown status code=%d", external.ErrInternal, r.StatusCode)
	}
}
