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

func NewMockServer() *httptest.Server {
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
			}

			w.Write(body)
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
			}

			w.Write(body)
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
			}

			w.Write(body)
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
			}

			w.Write(body)
			return
		}

		http.Error(w, "Not found.", 404)
	}))
}
