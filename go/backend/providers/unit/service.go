package unit

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
	ErrInternal = errors.New("unit: internal error")
)

type Service interface {
	GetApplicationForm(ctx context.Context, userID string) (*ApplicationForm, error)
	CreateApplicationForm(ctx context.Context, args *CreateApplicationFormArgs) (*ApplicationForm, error)
}

type service struct {
	validator *validator.Validate
	baseURL   string
	token     string
}

type ServiceArgs struct {
	BaseURL string `validate:"required"`
	Token   string `validate:"required"`
}

func NewService(args ServiceArgs) (Service, error) {
	validator := validator.New()
	err := validator.Struct(args)
	if err != nil {
		return nil, err
	}

	return &service{
		validator: validator,
		baseURL:   args.BaseURL,
		token:     args.Token,
	}, nil
}

type ApplicationForm struct {
	ID  string
	URL string
}

func (self *service) GetApplicationForm(ctx context.Context, userID string) (*ApplicationForm, error) {

	url := fmt.Sprintf(`%s/application-forms?page[limit]=1&filter[tags]={"userId":"%s"}`, self.baseURL, userID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/vnd.api+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", self.token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var data struct {
		Data []struct {
			ID   string `json:"id"`
			Attr struct {
				Url string `json:"url"`
			} `json:"attributes"`
		} `json:"data"`
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	if len(data.Data) == 0 {
		return nil, nil
	}

	return &ApplicationForm{
		ID:  data.Data[0].ID,
		URL: data.Data[0].Attr.Url,
	}, nil
}

type CreateApplicationFormArgs struct {
	ID      string `validate:"required"`
	Email   string `validate:"required"`
	Country string `validate:"required"`
}

func (self *service) CreateApplicationForm(ctx context.Context, args *CreateApplicationFormArgs) (*ApplicationForm, error) {

	url := fmt.Sprintf(`%s/application-forms`, self.baseURL)

	var jsonStr = []byte(fmt.Sprintf(`{
		"data": {
			"type": "applicationForm",
			"attributes": {
				"tags": {
					"userId": "%s"
				},
				"allowedApplicationTypes": ["Individual"],
				"applicantDetails": {
					"applicationType": "Individual",
					"nationality": "%s",
					"email": "%s"
				}
			}
		}
	}`, args.ID, args.Country, args.Email))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/vnd.api+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", self.token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var data struct {
		Data struct {
			ID   string `json:"id"`
			Attr struct {
				Url string `json:"url"`
			} `json:"attributes"`
		} `json:"data"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	return &ApplicationForm{
		ID:  data.Data.ID,
		URL: data.Data.Attr.Url,
	}, nil
}
