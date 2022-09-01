package external

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var (
	ErrInternal = errors.New("unit client: internal error")
	ErrRequest  = errors.New("unit client: request error")
)

type ErrHttp struct {
	Code   int
	Errors []ResponseError
}

func (e *ErrHttp) Error() string {
	return fmt.Sprintf("unit client: http error. statusCode=%d, errors=%+v", e.Code, e.Errors)
}

type (
	// Context we add so we can match the application back to our user. Be very careful changing
	// this value.
	ApplicationTags struct {
		FynbosUserID string `json:"fynbosUserId,omitempty"`
	}

	// Context we add so we can match the ach back to our deposit. Be very careful changing
	// this value.
	DepositTags struct {
		DepositID string `json:"depositID,omitempty"`
	}

	Unit interface {
		CreateDepositAccount(ctx context.Context, args *CreateDepositAccountArgs) (*DepositAccount, error)
		FilterApplicationFormsByUserID(ctx context.Context, userID string) ([]ApplicationForm, error)
		GetStatements(ctx context.Context, customerID string) ([]Statement, error)
		GetStatementPDF(ctx context.Context, args *GetStatementPDFArgs) ([]byte, error)
		CreateApplicationForm(ctx context.Context, args *CreateApplicationFormArgs) (*ApplicationForm, error)
		CreateApplication(ctx context.Context, args *CreateApplicationArgs) (*Application, error)
		CreateCounterparty(ctx context.Context, args *CreateCounterpartyArgs) (*Counterparty, error)
		OriginateAch(ctx context.Context, args *OriginateAchArgs) (*AchPayment, error)
	}

	client struct {
		baseUrl string
		http    *http.Client
	}
)

type basicAuthTransport struct {
	baseTransport http.RoundTripper
	apiToken      string
}

// This sets the basic auth credentials on every request.
func (t basicAuthTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.apiToken))
	r.Header.Set("Content-Type", "application/vnd.api+json")
	return t.baseTransport.RoundTrip(r)
}

func newBasicAuthTransport(apiToken string) *basicAuthTransport {
	return &basicAuthTransport{
		baseTransport: &http.Transport{},
		apiToken:      apiToken,
	}
}

func NewClient(baseUrl string, apiToken string) *client {
	return &client{
		baseUrl: baseUrl,
		http: &http.Client{
			Transport: newBasicAuthTransport(apiToken),
		},
	}
}

type GetStatementPDFArgs struct {
	ID         string
	CustomerID string
}

func (c *client) GetStatementPDF(ctx context.Context, args *GetStatementPDFArgs) ([]byte, error) {
	url := fmt.Sprintf(`%s/statements/%s/pdf?filter[customerId]=%s`, c.baseUrl, args.ID, args.CustomerID)
	resp, err := c.http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrRequest, err)
	}

	// can't use parseResponse() here because it expects a json object
	// but we are getting a pdf file. that cause the json parser to
	// fail.
	ret, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	if !isStatusOkay(resp.StatusCode) {
		response := &Response{}
		err := json.Unmarshal(ret, response)
		if err != nil {
			return nil, fmt.Errorf("%w %s", ErrInternal, err)
		}
		return nil, &ErrHttp{
			Code:   resp.StatusCode,
			Errors: response.Errors,
		}
	}

	return ret, nil
}

func (c *client) GetStatements(ctx context.Context, customerID string) ([]Statement, error) {
	url := fmt.Sprintf(`%s/statements?filter[customerId]=%s`, c.baseUrl, customerID)
	resp, err := c.http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrRequest, err)
	}

	ret := []Statement{}
	err = parseResponse(resp, &ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (c *client) FilterApplicationFormsByUserID(ctx context.Context, userID string) ([]ApplicationForm, error) {
	url := fmt.Sprintf(`%s/application-forms?page[limit]=1&filter[tags]={"userId":"%s"}`, c.baseUrl, userID)
	resp, err := c.http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrRequest, err)
	}

	ret := []ApplicationForm{}
	err = parseResponse(resp, &ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
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
	form := &ApplicationFormRequest{
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
	payload, err := json.Marshal(form)
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

	ret := &ApplicationForm{}
	err = parseResponse(resp, ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
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

	ret := &Application{}
	err = parseResponse(resp, ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
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
	counterparty := &CounterpartyRequest{
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
				Customer: Relationship{
					Data: TypeData{
						Type: "customer",
						ID:   args.UnitCustomerID,
					},
				},
			},
		},
	}
	rawCounterpartyRequest, err := json.Marshal(counterparty)
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

	ret := &Counterparty{}
	err = parseResponse(resp, ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

type OriginateAchArgs struct {
	IdempotencyKey   string `validate:"required"`
	Amount           uint64
	Direction        string `validate:"oneof=credit,debit"`
	CounterpartyID   string `validate:"required"`
	DepositAccountID string `validate:"required"`
	Description      string
	Tags             map[string]string
}

func (c *client) OriginateAch(ctx context.Context, args *OriginateAchArgs) (*AchPayment, error) {
	url := fmt.Sprintf("%s/payments", c.baseUrl)
	data := AchPaymentRequest{
		Data: AchPayment{
			Type: "achPayment",
			Attributes: AchPaymentAttributes{
				Amount:         args.Amount,
				Direction:      args.Direction,
				Description:    args.Description,
				IdempotencyKey: args.IdempotencyKey,
			},
			Relationships: AchPaymentRelationships{
				Account: Relationship{
					Data: TypeData{
						Type: "depositAccount",
						ID:   args.DepositAccountID,
					},
				},
				Counterparty: Relationship{
					Data: TypeData{
						Type: "counterparty",
						ID:   args.CounterpartyID,
					},
				},
			},
		},
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.http.Do(request)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrRequest, err)
	}

	ret := &AchPayment{}
	err = parseResponse(resp, ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

type CreateDepositAccountArgs struct {
	CustomerID     string
	DepositProduct string
	Type           string
	IdempotencyKey string
}

func (c *client) CreateDepositAccount(
	ctx context.Context,
	args *CreateDepositAccountArgs,
) (*DepositAccount, error) {
	url := fmt.Sprintf(`%s/accounts`, c.baseUrl)
	data := CreateDepositAccountRequest{
		Data: DepositAccount{
			Type: "depositAccount",
			Attributes: DepositAccountAttributes{
				DepositProduct: args.DepositProduct,
				IdempotencyKey: args.IdempotencyKey,
			},
			Relationships: &DepositAccountRelationships{
				Customer: Relationship{
					Data: TypeData{
						ID:   args.CustomerID,
						Type: "customer",
					},
				},
			},
		},
	}
	rawData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(rawData))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrRequest, err)
	}

	ret := &DepositAccount{}
	err = parseResponse(resp, ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func isStatusOkay(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
}

func parseResponse(r *http.Response, data any) error {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	response := &Response{}
	err = json.Unmarshal(body, response)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	if !isStatusOkay(r.StatusCode) || len(response.Errors) > 0 {
		return &ErrHttp{
			Code:   r.StatusCode,
			Errors: response.Errors,
		}
	}

	err = json.Unmarshal(response.Data, data)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	return nil
}
