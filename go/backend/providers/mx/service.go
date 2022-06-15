package mx

//go:generate mockgen -destination=./mock.go -package=mx -source=./service.go

import (
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/identity"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

var (
	ErrInvalidArgument      = errors.New("mx provider: invalid argument.")
	ErrInternal             = errors.New("mx provider: internal error.")
	ErrNotFound             = errors.New("mx provider: not found.")
	ErrDuplicate            = errors.New("mx provider: duplicate.")
	ErrOwnershipCheckFailed = errors.New("mx provider: ownership check failed.")
	ErrUnauthorized         = errors.New("mx provider: unauthorized.")
)

type (
	Service interface {
		CreateUser(ctx context.Context) (string, error)
		GetWidgetUrl(ctx context.Context, mxUserGuid string) (string, error)
		CreateAccount(ctx context.Context, args *CreateAccountArgs) (*Account, error)
		GetAccount(ctx context.Context, guid string) (*Account, error)
		StartIdentityAggregation(ctx context.Context, id string) (*Member, error)
		GetMemberStatus(ctx context.Context, id string) (*Member, error)
		GetAccountOwner(ctx context.Context, id string) (*AccountOwner, error)
		ReadAccount(ctx context.Context, id string) (*MxAccount, error)
		GetSelectedAccountGuid(ctx context.Context, mxUserGuid string, mxMemberGuid string) (string, error)
		GetMxUserByAccountID(ctx context.Context, accountID string) (string, error)
		VerifyOwnership(ctx context.Context, id string) error
		GetConnectWidget(ctx context.Context, accountID string, identityID string) (string, error)
		InitiateCreateAccount(ctx context.Context, args *InitiateCreateAccountArgs) (string, error)
	}

	user struct {
		Guid             string
		ConnectWidgetUrl string `json:"connect_widget_url"`
	}

	Account struct {
		Guid       string //from mx
		UserGuid   string `db:"user_guid"`   // from mx
		MemberGuid string `db:"member_guid"` // from mx
		AccountID  string `db:"account_id"`  // Fynbos account id
		CreatedAt  string `db:"created_at"`
		UpdatedAt  string `db:"updated_at"`
	}

	Member struct {
		Guid                     string
		UserGuid                 string
		AggregatedAt             string `json:"aggregated_at"`
		IsBeingAggregated        bool   `json:"is_being_aggregated"`
		SuccessfullyAggregatedAt string `json:"successfully_aggregated_at"`
		ConnectionStatus         string `json:"connection_status"`
	}

	AccountOwnersResponse struct {
		AccountOwners []AccountOwner `json:"account_owners"`
	}

	AccountOwner struct {
		AccountGuid string `json:"account_guid"`
		OwnerName   string `json:"owner_name"`
		Country     string
		Email       string
		Phone       string

		// There are more fields (address, state etc.) but we wouldn't match on that at the moment.
	}

	ReadAccountResponse struct {
		Account MxAccount
	}

	MxAccount struct {
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

	ServiceArgs struct {
		BaseUrl         string           `validate:"required"`
		Username        string           `validate:"required"`
		Password        string           `validate:"required"`
		Db              *sqlx.DB         `validate:"required"`
		AccountsService accounts.Service `validate:"required"`
		IdentityService identity.Service `validate:"required"`
		Temporal        client.Client    `validate:"required"`
	}

	service struct {
		v               *validator.Validate
		mxClient        *http.Client
		baseUrl         string
		db              *sqlx.DB
		accountsService accounts.Service
		identityService identity.Service
		temporal        client.Client
	}
)

// This sets the basic auth credentials on every request.
type basicAuthTransport struct {
	baseTransport http.RoundTripper
	userName      string
	password      string
}

func (t basicAuthTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.SetBasicAuth(t.userName, t.password)
	r.Header.Set("Accept", "application/vnd.mx.api.v1+json")
	r.Header.Set("Content-Type", "application/json")
	return t.baseTransport.RoundTrip(r)
}

func newBasicAuthTransport(userName string, password string) *basicAuthTransport {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS12,
	}
	return &basicAuthTransport{
		baseTransport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		userName: userName,
		password: password,
	}
}

func NewService(args *ServiceArgs) (Service, error) {
	v := validator.New()
	if err := v.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	return &service{
		v:       v,
		baseUrl: args.BaseUrl,
		mxClient: &http.Client{
			Transport: newBasicAuthTransport(args.Username, args.Password),
		},
		db:              args.Db,
		accountsService: args.AccountsService,
		identityService: args.IdentityService,
		temporal:        args.Temporal,
	}, nil
}

func (s *service) CreateUser(ctx context.Context) (string, error) {
	payload := `{ "user": { "is_disabled": false } }`
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/users", s.baseUrl), bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := s.mxClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	data := &struct {
		User user
	}{}
	if err = json.Unmarshal(body, data); err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	return data.User.Guid, nil
}

func (s *service) GetWidgetUrl(ctx context.Context, mxUserID string) (string, error) {
	url := fmt.Sprintf("%s/users/%s/connect_widget_url", s.baseUrl, mxUserID)
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
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := s.mxClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	data := &struct {
		User user
	}{}
	if err = json.Unmarshal(body, data); err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}
	return data.User.ConnectWidgetUrl, nil
}

type CreateAccountArgs struct {
	Guid       string `validate:"uuid4"`    // from mx
	UserGuid   string `validate:"required"` // from mx
	MemberGuid string `validate:"required"` // from mx
	AccountID  string `validate:"uuid4"`
}

func (s *service) CreateAccount(
	ctx context.Context,
	args *CreateAccountArgs,
) (*Account, error) {
	if err := s.v.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	ret := &Account{}
	err := s.db.GetContext(
		ctx,
		ret,
		`
		INSERT INTO mx_accounts (
			guid,
			user_guid,
			member_guid,
			account_id
		)
		VALUES ($1, $2, $3, $4) RETURNING *;
		`,
		args.Guid,
		args.UserGuid,
		args.MemberGuid,
		args.AccountID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint \"primary\"") {
			return nil, ErrDuplicate
		}
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return ret, nil
}

func (s service) GetAccount(ctx context.Context, guid string) (*Account, error) {
	ret := &Account{}
	err := s.db.GetContext(ctx, ret, "SELECT * FROM mx_accounts WHERE guid=$1", guid)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%w %s", ErrNotFound, fmt.Sprintf("guid=%s", guid))
	} else {
		if err != nil {
			return nil, fmt.Errorf("%w %s", ErrInternal, err)
		}
	}

	return ret, nil
}

func (s *service) StartIdentityAggregation(ctx context.Context, mxFundingSourceID string) (*Member, error) {
	mxAccount, err := s.GetAccount(ctx, mxFundingSourceID)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/users/%s/members/%s/identify", s.baseUrl, mxAccount.UserGuid, mxAccount.MemberGuid)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := s.mxClient.Do(req)
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

func (s *service) GetMemberStatus(ctx context.Context, mxFundingSourceID string) (*Member, error) {
	mxAccount, err := s.GetAccount(ctx, mxFundingSourceID)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/users/%s/members/%s/status", s.baseUrl, mxAccount.UserGuid, mxAccount.MemberGuid)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := s.mxClient.Do(req)
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

func (s service) GetAccountOwner(
	ctx context.Context,
	mxFundingSourceID string,
) (*AccountOwner, error) {
	mxAccount, err := s.GetAccount(ctx, mxFundingSourceID)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/users/%s/members/%s/account_owners", s.baseUrl, mxAccount.UserGuid, mxAccount.MemberGuid)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := s.mxClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	accountOwnersResp := AccountOwnersResponse{}
	if err = json.Unmarshal(body, &accountOwnersResp); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	var ret *AccountOwner = nil
	for _, owner := range accountOwnersResp.AccountOwners {
		if owner.AccountGuid == mxAccount.Guid {
			ret = &owner
			break
		}
	}
	if ret == nil {
		return nil, fmt.Errorf(
			"%w No account owner details found for mx account guid=%s",
			ErrNotFound,
			mxAccount.Guid,
		)
	}

	return ret, nil
}

func (s service) ReadAccount(ctx context.Context, mxFundingSourceID string) (*MxAccount, error) {
	mxAccount, err := s.GetAccount(ctx, mxFundingSourceID)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/users/%s/accounts/%s", s.baseUrl, mxAccount.UserGuid, mxAccount.Guid)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	resp, err := s.mxClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	ret := &ReadAccountResponse{}
	if err = json.Unmarshal(body, ret); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &ret.Account, nil
}

// The mx connect widget will allow the user to log into their bank and select an account.
// They do not pass this to us on the front end and so we need to call out to find out the
// mx account guid of the account that was selected.
// Calling the users/:users/members/:members/account_numbers should only have the account selected
// by the user.
func (s service) GetSelectedAccountGuid(ctx context.Context, mxUserGuid string, mxMemberGuid string) (string, error) {
	url := fmt.Sprintf("%s/users/%s/members/%s/account_numbers?page=1&records_per_page=10", s.baseUrl, mxUserGuid, mxMemberGuid)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}
	resp, err := s.mxClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	var accounts struct {
		AccountNumbers []struct {
			AccountGuid string `json:"account_guid"`
			UserGuid    string `json:"user_guid"`
			MemberGuid  string `json:"member_guid"`
		} `json:"account_numbers"`
	}
	if err = json.Unmarshal(body, &accounts); err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	if len(accounts.AccountNumbers) != 1 {
		return "", fmt.Errorf(
			"%w Unable to find account user selected. %d accounts were returned.",
			ErrInternal,
			len(accounts.AccountNumbers),
		)
	}

	return accounts.AccountNumbers[0].AccountGuid, nil
}

func (s service) GetMxUserByAccountID(ctx context.Context, accountID string) (string, error) {
	mxUserGuids := []string{}
	err := s.db.SelectContext(
		ctx,
		&mxUserGuids,
		"SELECT DISTINCT user_guid FROM mx_accounts WHERE account_id=$1;",
		accountID,
	)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	if len(mxUserGuids) == 0 {
		return "", fmt.Errorf("%w", ErrNotFound)
	}

	if len(mxUserGuids) != 1 {
		return "", fmt.Errorf("%w There are %d mx users linked to accountID=%s", ErrInternal, len(mxUserGuids), accountID)
	}

	return mxUserGuids[0], nil
}

func (s *service) VerifyOwnership(ctx context.Context, id string) error {
	mxAccount, err := s.GetAccount(ctx, id)
	if err != nil {
		return err
	}
	acc, err := s.accountsService.Get(ctx, mxAccount.AccountID)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	user, err := s.identityService.Get(ctx, acc.IdentityID)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	ownerDetails, err := s.GetAccountOwner(ctx, mxAccount.Guid)
	if err != nil {
		return err
	}

	// This verification can be extended in future.
	if strings.TrimSpace(fmt.Sprintf("%s %s", user.FirstName, user.LastName)) != strings.TrimSpace(ownerDetails.OwnerName) {
		return ErrOwnershipCheckFailed
	}

	return nil
}

func (s *service) GetConnectWidget(
	ctx context.Context,
	accountID string,
	identityID string,
) (string, error) {
	acc, err := s.accountsService.Get(ctx, accountID)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	mxUserGuid := ""
	mxUserGuid, err = s.GetMxUserByAccountID(ctx, acc.ID)
	if errors.Is(err, ErrNotFound) {
		mxUserGuid, err = s.CreateUser(ctx)
		if err != nil {
			return "", fmt.Errorf("%w %s", ErrInternal, err)
		}
	} else {
		if err != nil {
			return "", fmt.Errorf("%w %s", ErrInternal, err)
		}

		// mx user found so carry on.
	}

	url, err := s.GetWidgetUrl(ctx, mxUserGuid)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	return url, nil
}

type InitiateCreateAccountArgs struct {
	UserGuid          string `validate:"required"` // from mx
	MemberGuid        string `validate:"required"` // from mx
	AccountID         string `validate:"uuid4"`
	IdentityID        string `validate:"uuid4"`
	FundingsourceName string `validate:"required"`
}

func (s *service) InitiateCreateAccount(
	ctx context.Context,
	args *InitiateCreateAccountArgs,
) (string, error) {
	if err := s.v.Struct(args); err != nil {
		return "", fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	acc, err := s.accountsService.Get(ctx, args.AccountID)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	if acc.IdentityID != args.IdentityID {
		return "", ErrUnauthorized
	}

	workflowUuid := uuid.NewString()
	_, err = s.temporal.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:                    "create_mx_bank_account_" + workflowUuid,
			TaskQueue:             "backend",
			WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
		},
		CreateMxAccountWorkflow,
		&CreateMxAccountWorkflowArgs{
			ID:                workflowUuid,
			IdentityID:        args.IdentityID,
			AccountID:         args.AccountID,
			UserGuid:          args.UserGuid,
			MemberGuid:        args.MemberGuid,
			FundingsourceName: args.FundingsourceName,
		},
	)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	return workflowUuid, nil
}
