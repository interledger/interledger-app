package twilio

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
)

func NewMockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		servicesRegex := regexp.MustCompile("^/v2/Services/[^/]+$")
		if servicesRegex.MatchString(r.URL.Path) {
			if r.Method != "GET" {
				http.Error(w, "Method not implemented.", 501)
				return
			}

			data := struct {
				Sid string `json:"sid"`
			}{
				Sid: "VA1234567890abcdef1234567890abcdef",
			}

			body, err := json.Marshal(data)
			if err != nil {
				http.Error(w, "Failed to marshal response.", 500)
				return
			}

			if _, err := w.Write(body); err != nil {
				http.Error(w, "Failed to write response.", 500)
				return
			}

			return
		}

		verificationsRegex := regexp.MustCompile("/v2/Services/.*/Verifications")
		if verificationsRegex.MatchString(r.URL.Path) {
			if r.Method != "POST" {
				http.Error(w, "Method not implemented.", 501)
				return
			}

			err := r.ParseForm()
			if err != nil {
				http.Error(w, "Failed to parse payload.", 500)
				return
			}

			phoneNumber := r.PostForm["To"][0]

			data := struct {
				Sid    string `json:"sid"`
				To     string `json:"to"`
				Status string `json:"status"`
			}{
				Sid:    "VEXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
				To:     phoneNumber,
				Status: "pending",
			}

			body, err := json.Marshal(data)
			if err != nil {
				http.Error(w, "Failed to marshal response.", 500)
				return
			}

			if _, err := w.Write(body); err != nil {
				http.Error(w, "Failed to write response.", 500)
				return
			}

			return
		}

		verificationCheckRegex := regexp.MustCompile("/v2/Services/.*/VerificationCheck")
		if verificationCheckRegex.MatchString(r.URL.Path) {
			if r.Method != "POST" {
				http.Error(w, "Method not implemented.", 501)
				return
			}

			err := r.ParseForm()
			if err != nil {
				http.Error(w, "Failed to parse payload.", 500)
				return
			}

			phoneNumber := r.PostForm["To"][0]

			data := struct {
				Sid    string `json:"sid"`
				To     string `json:"to"`
				Status string `json:"status"`
			}{
				Sid:    "VEXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
				To:     phoneNumber,
				Status: "approved",
			}

			body, err := json.Marshal(data)
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
