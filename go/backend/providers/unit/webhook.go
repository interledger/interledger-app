package unit

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func NewHttpHandler(provider Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := provider.VerifyWebhook(context.Background(), r); err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Signature didn't match."))
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Failed to parse payload"))
			return
		}

		rawEvents := []Event{}
		if err := json.Unmarshal(body, &rawEvents); err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Failed to parse payload"))
			return
		}

		w.WriteHeader(200)
	}
}

type Event struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type CustomerCreatedEvent struct {
	ID            string                       `json:"id"`
	Type          string                       `json:"type"`
	Attributes    CustomerCreatedAttributes    `json:"attributes"`
	Relationships CustomerCreatedRelationships `json:"relationships"`
}

type CustomerCreatedAttributes struct {
	CreatedAt string            `json:"createdAt"`
	Tags      map[string]string `json:"tags"`
}

type CustomerCreatedRelationships struct {
	Customer    Customer    `json:"customer"`
	Application Application `json:"application"`
}

type Customer struct {
	Data Data `json:"data"`
}

type Application struct {
	Data Data `json:"data"`
}

type Data struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}
