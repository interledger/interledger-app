package external

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	_http "net/http"
)

var (
	ErrInternal = errors.New("unit client: internal error")
	ErrRequest  = errors.New("unit client: request error")
)

type ErrHttp struct {
	Code int
	Body []byte
}

func (e *ErrHttp) Error() string {
	return fmt.Sprintf("unit client: http error. statusCode=%d, body=%s", e.Code, string(e.Body))
}

type (
	// Context we add so we can match the application back to our user. Be very careful changing
	// this value.
	ApplicationTags struct {
		FynbosUserID string `json:"fynbosUserId,omitempty"`
	}

	Unit interface {
		FilterApplicationFormsByUserID(ctx context.Context, userID string) ([]ApplicationForm, error)
		CreateApplicationForm(ctx context.Context, args *CreateApplicationFormArgs) (*ApplicationForm, error)
		CreateApplication(ctx context.Context, args *CreateApplicationArgs) (*Application, error)
		CreateCounterparty(ctx context.Context, args *CreateCounterpartyArgs) (*Counterparty, error)
	}

	client struct {
		baseUrl string
		http    *_http.Client
	}
)

type basicAuthTransport struct {
	baseTransport _http.RoundTripper
	apiToken      string
}

// This sets the basic auth credentials on every request.
func (t basicAuthTransport) RoundTrip(r *_http.Request) (*_http.Response, error) {
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.apiToken))
	r.Header.Set("Content-Type", "application/vnd.api+json")
	return t.baseTransport.RoundTrip(r)
}

func newBasicAuthTransport(apiToken string) *basicAuthTransport {
	return &basicAuthTransport{
		baseTransport: &_http.Transport{},
		apiToken:      apiToken,
	}
}

func NewClient(baseUrl string, apiToken string) *client {
	return &client{
		baseUrl: baseUrl,
		http: &_http.Client{
			Transport: newBasicAuthTransport(apiToken),
		},
	}
}

func (c *client) FilterApplicationFormsByUserID(ctx context.Context, userID string) ([]ApplicationForm, error) {
	url := fmt.Sprintf(`%s/application-forms?page[limit]=1&filter[tags]={"userId":"%s"}`, c.baseUrl, userID)
	resp, err := c.http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrRequest, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	if !isStatusOkay(resp.StatusCode) {
		return nil, &ErrHttp{
			Code: resp.StatusCode,
			Body: body,
		}
	}

	listFormResponse := &ListApplicationFormRequest{}
	err = json.Unmarshal(body, listFormResponse)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return listFormResponse.Data, nil
}

type CreateApplicationFormArgs struct {
	ID      string `validate:"required"`
	Email   string `validate:"required"`
	Country string `validate:"required"`
}

func (c *client) CreateApplicationForm(
	ctx context.Context,
	args *CreateApplicationFormArgs,
) (*ApplicationForm, error) {
	url := fmt.Sprintf(`%s/application-forms`, c.baseUrl)
	formRequest := &ApplicationFormRequest{
		Data: ApplicationForm{
			Attributes: ApplicationFormAttributes{
				Tags: ApplicationTags{
					FynbosUserID: args.ID,
				},
				AllowedApplicationTypes: []string{"Individual"},
				ApplicationDetails: ApplicationFormPrefill{
					ApplicationType: "Individual",
					Nationality:     args.Country,
					Email:           args.Email,
				},
			},
		},
	}
	payload, err := json.Marshal(formRequest)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrRequest, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	if !isStatusOkay(resp.StatusCode) {
		return nil, &ErrHttp{
			Code: resp.StatusCode,
			Body: body,
		}
	}
	formResponse := &ApplicationFormRequest{}
	err = json.Unmarshal(body, formResponse)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &formResponse.Data, nil
}

type CreateApplicationArgs struct {
	UserID             string
	Email              string
	IpAddress          string
	DateOfBirth        string
	FirstName          string
	LastName           string
	Ssn                string
	Address            Address
	Phone              Phone
	DeviceFingerprints []DeviceFingerprint
}

func (c *client) CreateApplication(
	ctx context.Context,
	args *CreateApplicationArgs,
) (*Application, error) {
	url := fmt.Sprintf(`%s/applications`, c.baseUrl)
	application := &ApplicationRequest{
		Data: ApplicationWithoutRelationships{
			Type: "individualApplication",
			Attributes: ApplicationAttributes{
				Ssn: args.Ssn,
				FullName: &FullName{
					First: args.FirstName,
					Last:  args.LastName,
				},
				DateOfBirth: args.DateOfBirth,
				Address:     &args.Address,
				Email:       args.Email,
				Phone:       &args.Phone,
				IP:          args.IpAddress,
				Tags: &ApplicationTags{
					FynbosUserID: args.UserID,
				},
				DeviceFingerprints: args.DeviceFingerprints,
			},
		},
	}

	rawApplication, err := json.Marshal(application)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(rawApplication))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrRequest, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	if !isStatusOkay(resp.StatusCode) {
		return nil, &ErrHttp{
			Code: resp.StatusCode,
			Body: body,
		}
	}

	applicationResponse := &ApplicationResponse{}
	err = json.Unmarshal(body, applicationResponse)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &applicationResponse.Data, nil
}

type CreateCounterpartyArgs struct {
	Name           string
	UnitCustomerID string
	RoutingNumber  string
	AccountNumber  string
	AccountType    string
	Type           string

	// This idempotency key is valid for 48 hours on Unit's api.
	IdempotencyKey string
}

func (c *client) CreateCounterparty(
	ctx context.Context,
	args *CreateCounterpartyArgs,
) (*Counterparty, error) {
	url := fmt.Sprintf(`%s/counterparties`, c.baseUrl)
	counterpartyRequest := &CounterpartyRequest{
		Data: Counterparty{
			Type: "achCounterparty",
			Attributes: CounterpartyAttributes{
				Name:           args.Name,
				RoutingNumber:  args.RoutingNumber,
				AccountNumber:  args.AccountNumber,
				AccountType:    args.AccountType,
				Type:           args.Type,
				IdempotencyKey: args.IdempotencyKey,
			},
			Relationships: CounterpartyRelationships{
				Customer: Customer{
					Data: TypeData{
						ID:   args.UnitCustomerID,
						Type: "customer",
					},
				},
			},
		},
	}
	rawCounterpartyRequest, err := json.Marshal(counterpartyRequest)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(rawCounterpartyRequest))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrRequest, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	if !isStatusOkay(resp.StatusCode) {
		return nil, &ErrHttp{
			Code: resp.StatusCode,
			Body: body,
		}
	}
	counterpartyResponse := &CounterpartyRequest{}
	err = json.Unmarshal(body, counterpartyResponse)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &counterpartyResponse.Data, nil
}

func isStatusOkay(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
}
