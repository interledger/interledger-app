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
	"github.com/jmoiron/sqlx"
)

var (
	ErrInvalidArgument = errors.New("mx provider: invalid argument.")
	ErrInternal        = errors.New("mx provider: internal error.")
	ErrNotFound        = errors.New("mx provider: not found.")
	ErrDuplicate       = errors.New("mx provider: duplicate.")
)

type (
	Service interface {
		CreateUser(ctx context.Context) (string, error)
		GetWidgetUrl(ctx context.Context, mxUserGuid string) (string, error)
		CreateMxFundingSource(ctx context.Context, args *CreateMxFundingSourceArgs) (*MxFundingSource, error)
		GetMxFundingSource(ctx context.Context, id string) (*MxFundingSource, error)
		StartIdentityAggregation(ctx context.Context, mxFundingSourceID string) (*Member, error)
		GetMemberStatus(ctx context.Context, mxFundingSourceID string) (*Member, error)
		GetAccountOwner(ctx context.Context, mxFundingSourceID string) (*AccountOwner, error)
		GetMxAccount(ctx context.Context, mxFundingSourceID string) (*MxAccount, error)
		GetSelectedAccountGuid(ctx context.Context, mxUserGuid string, mxMemberGuid string) (string, error)
		GetMxUserByAccountID(ctx context.Context, accountID string) (string, error)
	}

	user struct {
		Guid             string
		ConnectWidgetUrl string `json:"connect_widget_url"`
	}

	MxFundingSource struct {
		ID              string
		AccountID       string `db:"account_id"`
		MxUserGuid      string `db:"mx_user_guid"`
		MxMemberGuid    string `db:"mx_member_guid"`
		MxAccountGuidID string `db:"mx_account_guid"`
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
		// Db *sqlx.DB
		BaseUrl  string   `validate:"required"`
		Username string   `validate:"required"`
		Password string   `validate:"required"`
		Db       *sqlx.DB `validate:"required"`
	}

	service struct {
		v        *validator.Validate
		mxClient *http.Client
		baseUrl  string
		db       *sqlx.DB
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
		db: args.Db,
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

type CreateMxFundingSourceArgs struct {
	ID            string `validate:"uuid4"`
	AccountID     string `validate:"uuid4"`
	MxUserGuid    string `validate:"required"`
	MxMemberGuid  string `validate:"required"`
	MxAccountGuid string `validate:"required"`
}

func (s *service) CreateMxFundingSource(
	ctx context.Context,
	args *CreateMxFundingSourceArgs,
) (*MxFundingSource, error) {
	if err := s.v.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	ret := &MxFundingSource{}
	err := s.db.GetContext(
		ctx,
		ret,
		`
		INSERT INTO mx_fundingsources (id, account_id, mx_user_guid, mx_member_guid, mx_account_guid)
		VALUES ($1, $2, $3, $4, $5) RETURNING *;
		`,
		args.ID,
		args.AccountID,
		args.MxUserGuid,
		args.MxMemberGuid,
		args.MxAccountGuid,
	)
	if err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint \"primary\"") {
			return nil, ErrDuplicate
		}
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return ret, nil
}

func (s service) GetMxFundingSource(ctx context.Context, id string) (*MxFundingSource, error) {
	ret := &MxFundingSource{}
	err := s.db.GetContext(ctx, ret, "SELECT * FROM mx_fundingsources WHERE id=$1", id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%w %s", ErrNotFound, fmt.Sprintf("id=%s", id))
	} else {
		if err != nil {
			return nil, fmt.Errorf("%w %s", ErrInternal, err)
		}
	}

	return ret, nil
}

func (s *service) StartIdentityAggregation(ctx context.Context, mxFundingSourceID string) (*Member, error) {
	mxFs, err := s.GetMxFundingSource(ctx, mxFundingSourceID)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/users/%s/members/%s/identify", s.baseUrl, mxFs.MxUserGuid, mxFs.MxMemberGuid)
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
	mxFs, err := s.GetMxFundingSource(ctx, mxFundingSourceID)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/users/%s/members/%s/status", s.baseUrl, mxFs.MxUserGuid, mxFs.MxMemberGuid)
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
	mxFs, err := s.GetMxFundingSource(ctx, mxFundingSourceID)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/users/%s/members/%s/account_owners", s.baseUrl, mxFs.MxUserGuid, mxFs.MxMemberGuid)
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
		if owner.AccountGuid == mxFs.MxAccountGuidID {
			ret = &owner
			break
		}
	}
	if ret == nil {
		return nil, fmt.Errorf(
			"%w No account owner details found for mx account guid=%s",
			ErrNotFound,
			mxFs.MxAccountGuidID,
		)
	}

	return ret, nil
}

func (s service) GetMxAccount(ctx context.Context, mxFundingSourceID string) (*MxAccount, error) {
	mxFs, err := s.GetMxFundingSource(ctx, mxFundingSourceID)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/users/%s/accounts/%s", s.baseUrl, mxFs.MxUserGuid, mxFs.MxAccountGuidID)
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
		"SELECT DISTINCT mx_user_guid FROM mx_fundingsources WHERE account_id=$1;",
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
