package client

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"gitlab.com/fynbos/env"

	"github.com/Basis-Theory/basistheory-go/v3"
	bt "gitlab.com/fynbos/backend/providers/basistheory"
	"gitlab.com/fynbos/backend/providers/basistheory/external"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var _ external.Client = &Client{}

func New(apiKey string) *Client {
	configuration := basistheory.NewConfiguration()
	configuration.HTTPClient = otelhttp.DefaultClient

	if env.IsLocal() {
		configuration.Servers = basistheory.ServerConfigurations{
			{
				URL:         "http://mockbos:8080/basistheory",
				Description: "No description provided",
			},
		}
	}

	return &Client{
		api:    basistheory.NewAPIClient(configuration),
		apiKey: apiKey,
	}
}

type Client struct {
	api    *basistheory.APIClient
	apiKey string
}

func (c *Client) GetToken(ctx context.Context, id string) (*basistheory.Token, error) {
	apiContext := context.WithValue(ctx, basistheory.ContextAPIKeys, map[string]basistheory.APIKey{
		"ApiKey": {Key: c.apiKey},
	})
	token, resp, err := c.api.TokensApi.GetById(apiContext, id).Execute()
	if err != nil && resp == nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	if err = checkResponseStatusCode(resp); err != nil {
		return nil, err
	}

	return token, nil
}

func (c *Client) CreateCardToken(ctx context.Context, args bt.CreateCardTokenArgs) (*basistheory.CreateTokenResponse, error) {
	apiContext := context.WithValue(ctx, basistheory.ContextAPIKeys, map[string]basistheory.APIKey{
		"ApiKey": {Key: c.apiKey},
	})
	data := map[string]interface{}{
		"number":           args.Number,
		"expiration_year":  args.ExpirationYear,
		"expiration_month": args.ExpirationMonth,
		"cvc":              args.CVC,
	}
	req := basistheory.NewCreateTokenRequest(data)
	req.SetType("card")
	req.SetMetadata(map[string]string{
		"walletID": args.WalletID,
	})
	req.SetDeduplicateToken(true)
	req.SetFingerprintExpression("{{ metadata.walletID }}{{ data.number }}")
	token, resp, err := c.api.TokensApi.Create(apiContext).CreateTokenRequest(*req).Execute()
	if err != nil && resp == nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	if err = checkResponseStatusCode(resp); err != nil {
		return nil, err
	}

	return token, nil
}

func checkResponseStatusCode(r *http.Response) error {
	if r.StatusCode != http.StatusMultiStatus &&
		(http.StatusOK <= r.StatusCode && r.StatusCode < http.StatusMultipleChoices) {
		return nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	switch r.StatusCode {
	case http.StatusMultiStatus:
		return fmt.Errorf("%w %s, path=%s", external.ErrBadRequest, string(body), r.Request.URL.Path)
	case http.StatusBadRequest:
		return fmt.Errorf("%w %s, path=%s", external.ErrBadRequest, string(body), r.Request.URL.Path)
	case http.StatusUnauthorized:
		return fmt.Errorf("%w %s, path=%s", external.ErrUnauthorized, string(body), r.Request.URL.Path)
	case http.StatusForbidden:
		return fmt.Errorf("%w %s, path=%s", external.ErrForbidden, string(body), r.Request.URL.Path)
	case http.StatusNotFound:
		return fmt.Errorf("%w %s, path=%s", external.ErrNotFound, string(body), r.Request.URL.Path)
	case http.StatusMethodNotAllowed:
		return fmt.Errorf("%w %s, path=%s", external.ErrMethodNotAllowed, string(body), r.Request.URL.Path)
	case http.StatusNotAcceptable:
		return fmt.Errorf("%w %s, path=%s", external.ErrNotAcceptable, string(body), r.Request.URL.Path)
	case http.StatusConflict:
		return fmt.Errorf("%w %s, path=%s", external.ErrConflict, string(body), r.Request.URL.Path)
	case http.StatusGone:
		return fmt.Errorf("%w %s, path=%s", external.ErrGone, string(body), r.Request.URL.Path)
	case http.StatusUnsupportedMediaType:
		return fmt.Errorf("%w %s, path=%s", external.ErrUnsupportedMediatype, string(body), r.Request.URL.Path)
	case http.StatusMisdirectedRequest:
		return fmt.Errorf("%w %s, path=%s", external.ErrMisdirectedRequest, string(body), r.Request.URL.Path)
	case http.StatusUnprocessableEntity:
		return fmt.Errorf("%w %s, path=%s", external.ErrUnprocessableEntity, string(body), r.Request.URL.Path)
	case http.StatusLocked:
		return fmt.Errorf("%w %s, path=%s", external.ErrLocked, string(body), r.Request.URL.Path)
	case http.StatusTooManyRequests:
		return fmt.Errorf("%w %s, path=%s", external.ErrTooManyRequests, string(body), r.Request.URL.Path)
	case http.StatusRequestHeaderFieldsTooLarge:
		return fmt.Errorf("%w %s, path=%s", external.ErrRequestHeadersTooLarge, string(body), r.Request.URL.Path)
	case http.StatusInternalServerError:
		return fmt.Errorf("%w %s, path=%s", external.ErrServer, string(body), r.Request.URL.Path)
	case http.StatusBadGateway:
		return fmt.Errorf("%w %s, path=%s", external.ErrBadGateway, string(body), r.Request.URL.Path)
	case http.StatusServiceUnavailable:
		return fmt.Errorf("%w %s, path=%s", external.ErrServiceUnavailable, string(body), r.Request.URL.Path)
	case http.StatusGatewayTimeout:
		return fmt.Errorf("%w %s, path=%s", external.ErrGatewayTimeout, string(body), r.Request.URL.Path)
	default:
		return fmt.Errorf("%w Unknown status code=%d, message=%s, path=%s", external.ErrInternal, r.StatusCode, string(body), r.Request.URL.Path)
	}
}
