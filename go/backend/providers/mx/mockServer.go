package mx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type MockServerState struct {
	AccountOwners []AccountOwner
	MxAccount     MxAccount
}

func NewMockServer(opts ...func(*MockServerState)) *httptest.Server {
	state := &MockServerState{}
	for _, opt := range opts {
		opt(state)
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users" {
			if r.Method != "POST" {
				http.Error(w, "Method not implemented.", 501)
				return
			}
			user := &user{Guid: uuid.NewString()}
			body, err := json.Marshal(user)
			if err != nil {
				http.Error(w, "Failed to marshal response.", 500)
				return
			}

			if _, err = w.Write(body); err != nil {
				http.Error(w, "Failed to write response.", 500)
				return
			}

			return
		}

		if strings.Contains(r.URL.Path, "connect_widget_url") {
			if r.Method != "POST" {
				http.Error(w, "Method not implemented.", 501)
				return
			}

			params := strings.Split(r.URL.Path, "/")

			user := &user{Guid: params[1], ConnectWidgetUrl: "http://localhost/connectwidget"}
			body, err := json.Marshal(user)
			if err != nil {
				http.Error(w, "Failed to marshal response.", 500)
				return
			}

			if _, err = w.Write(body); err != nil {
				http.Error(w, "Failed to write response.", 500)
				return
			}

			return
		}

		memberStatusRegex := regexp.MustCompile("/users/.*/members/.*/status")
		if memberStatusRegex.Match([]byte(r.URL.Path)) {
			if r.Method != "GET" {
				http.Error(w, "Method not implemented.", 501)
				return
			}

			params := strings.Split(r.URL.Path, "/")
			aggregatedAt := time.Now().Format(time.RFC3339)
			member := &Member{
				ConnectionStatus:         "CONNECTED",
				IsBeingAggregated:        false,
				SuccessfullyAggregatedAt: aggregatedAt,
				AggregatedAt:             aggregatedAt,
				UserGuid:                 params[2],
				Guid:                     params[4],
			}

			body, err := json.Marshal(member)
			if err != nil {
				http.Error(w, "Failed to marshal response.", 500)
				return
			}

			if _, err = w.Write(body); err != nil {
				http.Error(w, "Failed to write response.", 500)
				return
			}
			return
		}

		identifyRegex := regexp.MustCompile("/users/.*/members/.*/identify")
		if identifyRegex.Match([]byte(r.URL.Path)) {
			if r.Method != "POST" {
				http.Error(w, "Method not implemented.", 501)
				return
			}

			params := strings.Split(r.URL.Path, "/")
			aggregatedAt := time.Now().Format(time.RFC3339)
			member := &Member{
				ConnectionStatus:         "CONNECTED",
				IsBeingAggregated:        true,
				SuccessfullyAggregatedAt: aggregatedAt,
				AggregatedAt:             aggregatedAt,
				UserGuid:                 params[2],
				Guid:                     params[4],
			}

			body, err := json.Marshal(member)
			if err != nil {
				http.Error(w, "Failed to marshal response.", 500)
				return
			}

			if _, err = w.Write(body); err != nil {
				http.Error(w, "Failed to write response.", 500)
				return
			}
			return
		}

		accountOwnersRegex := regexp.MustCompile("/users/.*/members/.*/account_owners")
		if accountOwnersRegex.Match([]byte(r.URL.Path)) {
			if r.Method != "GET" {
				http.Error(w, "Method not implemented.", 501)
				return
			}

			body, err := json.Marshal(AccountOwnersResponse{state.AccountOwners})
			if err != nil {
				http.Error(w, "Failed to marshal response.", 500)
				return
			}

			if _, err = w.Write(body); err != nil {
				http.Error(w, "Failed to write response.", 500)
				return
			}
			return
		}

		readAccountRegex := regexp.MustCompile("/users/.*/accounts/.*")
		if readAccountRegex.Match([]byte(r.URL.Path)) {
			if r.Method != "GET" {
				http.Error(w, "Method not implemented", 501)
				return
			}

			params := strings.Split(r.URL.Path, "/")
			accountGuid := params[4]
			if accountGuid != state.MxAccount.Guid {
				http.Error(w, "Not found.", 404)
				return
			}

			body, err := json.Marshal(&ReadAccountResponse{state.MxAccount})
			if err != nil {
				http.Error(w, "Failed to marshal response.", 500)
				return
			}

			if _, err = w.Write(body); err != nil {
				http.Error(w, "Failed to write response.", 500)
				return
			}
			return
		}

		http.Error(w, "Not found.", 404)
	}))
}
