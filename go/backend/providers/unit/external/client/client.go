package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gitlab.com/fynbos/backend/providers/unit/external"
	"gitlab.com/fynbos/env"
)

var _ external.Client = client{}

const (
	sandboxUrl = "https://api.s.unit.sh"
	liveUrl    = "https://api.unit.sh"
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

type client struct {
	baseUrl string
	http    *http.Client
}

func NewClient(apiToken string) external.Client {
	baseUrl := sandboxUrl
	if env.IsProd() {
		baseUrl = liveUrl
	}

	return &client{
		baseUrl: baseUrl,
		http: &http.Client{
			Transport: newBasicAuthTransport(apiToken),
		},
	}
}

func (c client) GetStatementPDF(ctx context.Context, args *external.GetStatementPDFArgs) ([]byte, error) {
	url := fmt.Sprintf(`%s/statements/%s/pdf?filter[customerId]=%s`, c.baseUrl, args.ID, args.CustomerID)
	resp, err := c.http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrRequest, err)
	}

	// can't use parseResponse() here because it expects a json object
	// but we are getting a pdf file. that cause the json parser to
	// fail.
	ret, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	if !isStatusOkay(resp.StatusCode) {
		response := &external.Response{}
		err := json.Unmarshal(ret, response)
		if err != nil {
			return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
		}
		return nil, &external.ErrHttp{
			Code:   resp.StatusCode,
			Errors: response.Errors,
		}
	}

	return ret, nil
}

func (c client) GetStatements(ctx context.Context, customerID string) ([]external.Statement, error) {
	url := fmt.Sprintf(`%s/statements?filter[customerId]=%s`, c.baseUrl, customerID)
	resp, err := c.http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrRequest, err)
	}

	ret := []external.Statement{}
	err = parseResponse(resp, &ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (c client) FilterApplicationFormsByUserID(ctx context.Context, userID string) ([]external.ApplicationForm, error) {
	url := fmt.Sprintf(`%s/application-forms?page[limit]=1&filter[tags]={"userId":"%s"}`, c.baseUrl, userID)
	resp, err := c.http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrRequest, err)
	}

	ret := []external.ApplicationForm{}
	err = parseResponse(resp, &ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (c client) CreateApplicationForm(
	ctx context.Context,
	args *external.CreateApplicationFormArgs,
) (*external.ApplicationForm, error) {
	url := fmt.Sprintf(`%s/application-forms`, c.baseUrl)
	form := &external.ApplicationFormRequest{
		Data: external.ApplicationForm{
			Attributes: external.ApplicationFormAttributes{
				Tags: external.ApplicationTags{
					FynbosUserID: args.ID,
				},
				AllowedApplicationTypes: []string{"Individual"},
				ApplicationDetails: external.ApplicationFormPrefill{
					ApplicationType: "Individual",
					Nationality:     args.Country,
					Email:           args.Email,
				},
			},
		},
	}
	payload, err := json.Marshal(form)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrRequest, err)
	}

	ret := &external.ApplicationForm{}
	err = parseResponse(resp, ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (c client) CreateApplication(
	ctx context.Context,
	args *external.CreateApplicationArgs,
) (*external.Application, error) {
	url := fmt.Sprintf(`%s/applications`, c.baseUrl)
	application := &external.ApplicationRequest{
		Data: external.ApplicationWithoutRelationships{
			Type: "individualApplication",
			Attributes: external.ApplicationAttributes{
				Ssn: args.Ssn,
				FullName: &external.FullName{
					First: args.FirstName,
					Last:  args.LastName,
				},
				DateOfBirth: args.DateOfBirth,
				Address:     &args.Address,
				Email:       args.Email,
				Phone:       &args.Phone,
				IP:          args.IpAddress,
				Tags: &external.ApplicationTags{
					FynbosUserID: args.UserID,
				},
				DeviceFingerprints: args.DeviceFingerprints,
			},
		},
	}

	rawApplication, err := json.Marshal(application)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(rawApplication))
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrRequest, err)
	}

	ret := &external.Application{}
	err = parseResponse(resp, ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (c client) CreateCounterparty(
	ctx context.Context,
	args *external.CreateCounterpartyArgs,
) (*external.Counterparty, error) {
	url := fmt.Sprintf(`%s/counterparties`, c.baseUrl)
	counterparty := &external.CounterpartyRequest{
		Data: external.Counterparty{
			Type: "achCounterparty",
			Attributes: external.CounterpartyAttributes{
				Name:           args.Name,
				RoutingNumber:  args.RoutingNumber,
				AccountNumber:  args.AccountNumber,
				AccountType:    args.AccountType,
				Type:           args.Type,
				IdempotencyKey: args.IdempotencyKey,
			},
			Relationships: external.CounterpartyRelationships{
				Customer: external.Relationship{
					Data: external.TypeData{
						Type: "customer",
						ID:   args.UnitCustomerID,
					},
				},
			},
		},
	}
	rawCounterpartyRequest, err := json.Marshal(counterparty)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(rawCounterpartyRequest))
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrRequest, err)
	}

	ret := &external.Counterparty{}
	err = parseResponse(resp, ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (c client) OriginateAch(ctx context.Context, args *external.OriginateAchArgs) (*external.AchPayment, error) {
	url := fmt.Sprintf("%s/payments", c.baseUrl)
	data := external.AchPaymentRequest{
		Data: external.AchPayment{
			Type: "achPayment",
			Attributes: external.AchPaymentAttributes{
				Amount:         args.Amount,
				Direction:      args.Direction,
				Description:    args.Description,
				IdempotencyKey: args.IdempotencyKey,
			},
			Relationships: external.AchPaymentRelationships{
				Account: external.Relationship{
					Data: external.TypeData{
						Type: "depositAccount",
						ID:   args.DepositAccountID,
					},
				},
				Counterparty: external.Relationship{
					Data: external.TypeData{
						Type: "counterparty",
						ID:   args.CounterpartyID,
					},
				},
			},
		},
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	resp, err := c.http.Do(request)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrRequest, err)
	}

	ret := &external.AchPayment{}
	err = parseResponse(resp, ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (c client) CreateDepositAccount(
	ctx context.Context,
	args *external.CreateDepositAccountArgs,
) (*external.DepositAccount, error) {
	url := fmt.Sprintf(`%s/accounts`, c.baseUrl)
	data := external.CreateDepositAccountRequest{
		Data: external.DepositAccount{
			Type: "depositAccount",
			Attributes: external.DepositAccountAttributes{
				DepositProduct: args.DepositProduct,
				IdempotencyKey: args.IdempotencyKey,
			},
			Relationships: &external.DepositAccountRelationships{
				Customer: external.Relationship{
					Data: external.TypeData{
						ID:   args.CustomerID,
						Type: "customer",
					},
				},
			},
		},
	}
	rawData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(rawData))
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrRequest, err)
	}

	ret := &external.DepositAccount{}
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
		return fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	response := &external.Response{}
	err = json.Unmarshal(body, response)
	if err != nil {
		return fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	if !isStatusOkay(r.StatusCode) || len(response.Errors) > 0 {
		return &external.ErrHttp{
			Code:   r.StatusCode,
			Errors: response.Errors,
		}
	}

	err = json.Unmarshal(response.Data, data)
	if err != nil {
		return fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return nil
}
