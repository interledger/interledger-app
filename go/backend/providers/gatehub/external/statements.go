package external

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	httplog "gitlab.com/fynbos/backend/providers/http"
)

func (c *client) GetAccountStatement(ctx context.Context, userID, walletAddress string) (io.ReadCloser, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "GET"
		meta.Provider = "gatehub"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "GET",
			Provider: "gatehub",
		})
	}

	endpoint, err := url.JoinPath(c.baseURL, "statement", "v1", "statements", "account-confirmation", walletAddress)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Set(managedUserHeader, userID)
	err = c.Sign(ctx, req, time.Now(), nil, endpoint)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		resp.Body.Close()
		return nil, err
	}

	return resp.Body, nil
}
