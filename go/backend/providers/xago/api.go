package xago

import "net/http"

type Client interface {
	WebhookHandler() http.HandlerFunc
}
