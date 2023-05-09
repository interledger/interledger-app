package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

const insertFields = "provider, context, request_body, request_path, response_body, response_status"

// const fields = "id, provider, context, request_body, request_path, response_body, response_status, created_at"

func Log(b Backends, req *http.Request, resp *http.Response) {
	ctx := req.Context()
	if b.DB() == nil || ctx.Value(ContextKey) == nil {
		return
	}
	meta := ctx.Value(ContextKey).(*Metadata)

	reqBody, err := req.GetBody()
	if err != nil {
		log.Error("httplogger: Failed to log external api request.", zap.Error(err))
	}

	payload, err := io.ReadAll(reqBody)
	if err != nil {
		log.Error("httplogger: Failed to log external api request.", zap.Error(err))
	}
	defer func() {
		if payload != nil {
			req.Body = io.NopCloser(bytes.NewBuffer(payload))
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("httplogger: Failed to log external api request.", zap.Error(err))
	}
	defer func() {
		if respBody != nil {
			resp.Body = io.NopCloser(bytes.NewBuffer(respBody))
		}
	}()

	_, err = b.DB().ExecContext(
		ctx,
		fmt.Sprintf("INSERT INTO external_api_logs (%s) VALUES ($1, $2, $3, $4, $5, $6);", insertFields),
		meta.Provider,
		meta.Context,
		string(payload),
		req.URL.Path,
		string(respBody),
		resp.Status,
	)
	if err != nil {
		log.Error("httplogger: Failed to log external api request.", zap.Error(err))
	}
}
