package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type requestOption func(*http.Request)

func withHeader(key, value string) requestOption {
	return func(r *http.Request) { r.Header.Set(key, value) }
}

func withQueryParam(key, value string) requestOption {
	return func(r *http.Request) {
		q := r.URL.Query()
		q.Set(key, value)
		r.URL.RawQuery = q.Encode()
	}
}

func (client *client) get(ctx context.Context, base, path string, opts ...requestOption) (*http.Response, error) {
	endpoint, err := url.JoinPath(base, path)
	if err != nil {
		return nil, err
	}

	fmt.Println("endpoint", endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(req)
	}

	resp, err := client.http.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (client *client) put(ctx context.Context, base, path string, payload any, opts ...requestOption) (*http.Response, error) {
	endpoint, err := url.JoinPath(base, path)
	if err != nil {
		return nil, err
	}

	fmt.Println("endpoint", endpoint)

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(req)
	}

	resp, err := client.http.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (client *client) post(ctx context.Context, base, path string, payload any, opts ...requestOption) (*http.Response, error) {
	endpoint, err := url.JoinPath(base, path)
	if err != nil {
		return nil, err
	}

	fmt.Println("endpoint", endpoint)

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(req)
	}

	resp, err := client.http.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// consumeResponse is a helper function to consume the response body and unmarshal it into the expected type
// it also checks for the expected status code and returns an error if it doesn't match
// this function is used in the controller methods to handle the responses from the API
// YOU are expected to consume the body with this function
// the body is closed in this function, so you don't have to worry about it in the controller methods
// in case of an error, the body is read and included in the error message for easier debugging
// if the body cannot be read, the error message will indicate that as well
// the payload will be unmarshaled into the expected type if it is a struct, otherwise as is
func consumeResponse[T any](resp *http.Response, expectedStatus int) (T, error) {
	defer resp.Body.Close()
	var zero T
	if resp.StatusCode != expectedStatus {
		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return zero, fmt.Errorf("%w: %d, response body could not be read: %s", ErrUnexpectedStatusCode, resp.StatusCode, err)
		}
		return zero, fmt.Errorf("%w: %d, response %s", ErrUnexpectedStatusCode, resp.StatusCode, string(bytes))
	}
	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return zero, fmt.Errorf("%w: %s", ErrDecodingResponse, err)
	}
	return result, nil
}
