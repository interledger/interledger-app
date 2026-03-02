package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// consumeResponse is a helper function to consume the response body and unmarshal it into the expected type
// it also checks for the expected status code and returns an error if it doesn't match
// this function is used in the controller methods to handle the responses from the API
// YOU are expected to consume the body with this function
// the body is closed in this function, so you don't have to worry about it in the controller methods
// in case of an error, the body is read and included in the error message for easier debugging
// if the body cannot be read, the error message will indicate that as well
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
