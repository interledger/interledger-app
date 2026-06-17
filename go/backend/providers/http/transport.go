package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/interledger/interledger-app/go/log"
	"go.uber.org/zap"
)

var _ http.RoundTripper = &Transport{}

func NewTransport(base http.RoundTripper, b Backends, redactor Redact) *Transport {
	if base == nil {
		base = http.DefaultTransport
	}

	redact := redactor
	if redact == nil {
		redact = DefaultRedactor
	}

	return &Transport{
		b:      b,
		base:   base,
		redact: redact,
	}
}

type Transport struct {
	b      Backends
	base   http.RoundTripper
	redact Redact
}

func DefaultRedactor(ctx context.Context, req []byte) ([]byte, error) {
	return req, nil
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := t.base.RoundTrip(req)
	if err == nil {
		t.Log(t.b, req, res)
	}

	return res, err
}

func (t *Transport) Log(b Backends, req *http.Request, resp *http.Response) {
	ctx := req.Context()
	meta, ok := MetaForContext(ctx)
	if b.DB() == nil || !ok || req == nil || resp == nil {
		return
	}

	var payload []byte
	if req.GetBody != nil { // not defined for GET requests
		reqBody, err := req.GetBody()
		if err != nil {
			log.Error("httplogger: Failed to log external api request.", zap.Error(err))
		}

		payload, err = io.ReadAll(reqBody)
		origPayload := make([]byte, len(payload))
		copy(origPayload, payload)
		if err != nil {
			log.Error("httplogger: Failed to log external api request.", zap.Error(err))

		}
		defer func() {
			if payload != nil {
				req.Body = io.NopCloser(bytes.NewBuffer(origPayload))
			}
		}()
	}

	respBody, err := io.ReadAll(resp.Body)
	origRespBody := make([]byte, len(respBody))
	copy(origRespBody, respBody)
	if err != nil {
		log.Error("httplogger: Failed to log external api request.", zap.Error(err))
	}
	defer func() {
		if respBody != nil {
			resp.Body = io.NopCloser(bytes.NewBuffer(origRespBody))
		}
	}()

	redactedRequest, err := t.redact(ctx, payload)
	if err != nil {
		log.Error("httplogger: Failed to redact request.", zap.Error(err))
		return
	}

	_, err = b.DB().ExecContext(
		ctx,
		fmt.Sprintf("INSERT INTO external_api_logs (%s) VALUES ($1, $2, $3, $4, $5, $6, $7);", insertFields),
		meta.Provider,
		meta.Context,
		string(redactedRequest),
		req.URL.String(),
		string(respBody),
		resp.Status,
		meta.Method,
	)
	if err != nil {
		log.Error("httplogger: Failed to log external api request.", zap.Error(err))
	}
}
