package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	httplog "github.com/interledger/interledger-app/go/backend/providers/http"
)

func (c client) CreateWallet(ctx context.Context, args CreateWalletArgs) (*Wallet, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = http.MethodPost
		meta.Provider = ptiProviderName

	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   http.MethodPost,
			Provider: ptiProviderName,
		})
	}

	url, err := url.JoinPath(c.baseURL, "users", args.UserID, "wallets")
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	payload, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Add(ptiClientIDHeader, c.clientID)
	req.Header.Add("Content-Type", "application/json")
	date := time.Now()
	req.Header.Add("Date", date.Format(http.TimeFormat))
	if err := sign(req, date, payload, c.privateKey, c.publicKeyThumbprint); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	var wallet Wallet
	err = json.Unmarshal(body, &wallet)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &wallet, nil
}

func (c client) GetWallet(ctx context.Context, userID, id string) (*Wallet, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = http.MethodGet
		meta.Provider = ptiProviderName
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   http.MethodGet,
			Provider: ptiProviderName,
		})
	}

	url, err := url.JoinPath(c.baseURL, "users", userID, "wallets", id)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Add(ptiClientIDHeader, c.clientID)
	date := time.Now()
	req.Header.Add("Date", date.Format(http.TimeFormat))
	if err := sign(req, date, nil, c.privateKey, c.publicKeyThumbprint); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	var wallet Wallet
	err = json.Unmarshal(body, &wallet)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &wallet, nil
}

func (c client) ListWallets(ctx context.Context, userID string) ([]Wallet, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = http.MethodGet
		meta.Provider = ptiProviderName
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   http.MethodGet,
			Provider: ptiProviderName,
		})
	}

	url, err := url.JoinPath(c.baseURL, "users", userID, "wallets")
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Add(ptiClientIDHeader, c.clientID)
	date := time.Now()
	req.Header.Add("Date", date.Format(http.TimeFormat))
	if err := sign(req, date, nil, c.privateKey, c.publicKeyThumbprint); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	var wallets []Wallet
	err = json.Unmarshal(body, &wallets)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return wallets, nil
}
