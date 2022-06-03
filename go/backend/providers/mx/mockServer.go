package mx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

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

		http.Error(w, "Not found.", 404)
	}))
}
