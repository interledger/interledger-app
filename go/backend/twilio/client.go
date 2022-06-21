package twilio

import (
	"net/http"
	"net/url"

	"github.com/twilio/twilio-go/client"
)

type CustomClient struct {
	client.Client
	BaseURL string
}

func (c *CustomClient) SendRequest(method string, rawURL string, data url.Values, headers map[string]interface{}) (*http.Response, error) {
	if c.BaseURL == "" {
		resp, err := c.Client.SendRequest(method, rawURL, data, headers)
		return resp, err
	}

	requestUrl, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	baseUrl, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	requestUrl.Scheme = baseUrl.Scheme
	requestUrl.Host = baseUrl.Host

	resp, err := c.Client.SendRequest(method, requestUrl.String(), data, headers)
	return resp, err
}
