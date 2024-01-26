package basistheory

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"

	"github.com/Basis-Theory/basistheory-go/v3"
	"github.com/google/uuid"

	"gitlab.com/fynbos/log"
	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
)

func Register(r chi.Router) {
	r.HandleFunc("/", forwardRequest)
	r.Get("/tokens/{id}", tokenReq)
}

func tokenReq(w http.ResponseWriter, r *http.Request) {
	var bt basistheory.Token
	bt.SetId(uuid.NewString())
	bt.SetType("card")
	bt.SetFingerprint(uuid.NewString())
	bt.SetData(map[string]interface{}{
		"number":           "2135544",
		"expiration_month": 2,
		"expiration_year":  2027,
	})

	resp, _ := json.Marshal(bt)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(resp)
}

var (
	cardRegex = regexp.MustCompile(`\{\{ (.+?) \| json: '\$\.number' \}\}`)
	cvvRegex  = regexp.MustCompile(`\{\{ (.+?) \| json: '\$\.number' \}\}`)
	month     = regexp.MustCompile(`\{\{ (.+?) \| json: '\$\.expiration_month' \| pad_left: 2,'0' \}\}`)
	year      = regexp.MustCompile(`\{\{ (.+?) \| json: '\$\.expiration_year' \| to_string \| slice: -2, 2 \}\}`)
)

func forwardRequest(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("failed to read body to forward", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	body = cardRegex.ReplaceAll(body, []byte("5425233430109903"))
	body = cvvRegex.ReplaceAll(body, []byte("125"))
	body = month.ReplaceAll(body, []byte("02"))
	body = year.ReplaceAll(body, []byte("26"))

	req, err := http.NewRequestWithContext(r.Context(), r.Method, r.Header.Get("BT-PROXY-URL"), bytes.NewReader(body))
	if err != nil {
		log.Error("failed to create request to forward", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	req.Header = r.Header.Clone()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("failed to forward request", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("failed to read forward request response body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(respBody)
}
