package mx

//go:generate mockgen -destination=./mock.go -package=mx -source=./service.go

import (
	"bytes"
	"context"
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
	return t.baseTransport.RoundTrip(r)
}

func newBasicAuthTransport(userName string, password string) *basicAuthTransport {
	return &basicAuthTransport{
		baseTransport: http.DefaultTransport,
		userName:      userName,
		password:      password,
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
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/users", s.baseUrl), bytes.NewBuffer([]byte{}))
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

	user := &user{}
	if err = json.Unmarshal(body, user); err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	return user.Guid, nil
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

	user := &user{}
	if err = json.Unmarshal(body, user); err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}
	return user.ConnectWidgetUrl, nil
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
