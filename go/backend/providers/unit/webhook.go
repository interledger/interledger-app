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

		payload, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Failed to parse payload"))
			return
		}

		body := Body{}
		if err := json.Unmarshal(payload, &body); err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Failed to parse payload"))
			return
		}

		for _, rawEvent := range body.Data {
			var event Event
			if err := json.Unmarshal(rawEvent, &event); err != nil {
				w.WriteHeader(500)
				w.Write([]byte("Failed to parse payload"))
				return
			}

			err := provider.HandleEvent(context.Background(), EventType(event.Type), rawEvent)
			if err != nil {
				w.WriteHeader(500)
				w.Write([]byte("Failed to handle event"))
				return
			}
		}

		w.WriteHeader(200)
	}
}

type (
	Body struct {
		Data []json.RawMessage `json:"data"`
	}
	EventType string
	Event     struct {
		ID   string    `json:"id"`
		Type EventType `json:"type"`
	}
	CustomerCreatedEvent struct {
		ID            string                       `json:"id"`
		Type          string                       `json:"type"`
		Attributes    CustomerCreatedAttributes    `json:"attributes"`
		Relationships CustomerCreatedRelationships `json:"relationships"`
	}

	CustomerCreatedAttributes struct {
		CreatedAt string            `json:"createdAt"`
		Tags      map[string]string `json:"tags"`
	}

	CustomerCreatedRelationships struct {
		Customer    Customer    `json:"customer"`
		Application Application `json:"application"`
	}

	Customer struct {
		Data Data `json:"data"`
	}

	Application struct {
		Data Data `json:"data"`
	}

	Data struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}
)

const (
	CUSTOMER_CREATED = EventType("customer.created")
)
