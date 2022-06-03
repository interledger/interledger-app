package mx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var (
	ErrInvalidArgument = errors.New("mx provider: invalid argument.")
	ErrInternal        = errors.New("mx provider: internal error.")
)

type (
	Service interface {
		CreateUser(ctx context.Context) (string, error)
		GetWidgetUrl(ctx context.Context, mxUserGuid string) (string, error)
	}

	user struct {
		Guid             string
		ConnectWidgetUrl string `json:"connect_widget_url"`
	}

	Account struct {
		ID         string
		MxID       string
		MxUserID   string
		MxMemberID string
	}

	ServiceArgs struct {
		// Db *sqlx.DB
		BaseUrl  string `validate:"required"`
		Username string `validate:"required"`
		Password string `validate:"required"`
	}

	service struct {
		v        *validator.Validate
		mxClient *http.Client
		baseUrl  string
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
