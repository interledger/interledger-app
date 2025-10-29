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

	httplog "gitlab.com/fynbos/backend/providers/http"
)

func (c *client) ListCards(ctx context.Context, userID, customerID string) (*ListCardsResponse, error) {
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

	endpoint, err := url.Parse(fmt.Sprintf("%s/cards/v1/cards/%s", c.baseURL, customerID))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	q := endpoint.Query()
	q.Set("pageSize", fmt.Sprintf("%d", 100))
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req.Header.Set(managedUserHeader, userID)
	req.Header.Set(cardAppIDHeader, c.cardAppID)

	err = c.Sign(ctx, req, time.Now(), nil, endpoint.String())
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var res ListCardsResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &res, nil
}

func (c *client) GetDeliveryAddresses(ctx context.Context, userID, customerID string) ([]CustomerDeliveryAddress, error) {
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

	endpoint, err := url.JoinPath(c.baseURL, "cards", "v1", "customers", customerID, "addresses")
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req.Header.Set(managedUserHeader, userID)
	req.Header.Set(cardAppIDHeader, c.cardAppID)

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
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var addrs []CustomerDeliveryAddress
	err = json.Unmarshal(body, &addrs)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return addrs, nil
}

func (c *client) GetCardApplicationProducts(ctx context.Context) ([]CardApplicationProduct, error) {
	if len(c.cardApplicationProducts) != 0 {
		return c.cardApplicationProducts, nil
	}

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

	endpoint, err := url.JoinPath(c.baseURL, "cards", "v1", "card-applications", c.cardAppID, "card-products")
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req.Header.Set(cardAppIDHeader, c.cardAppID)

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
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var products []CardApplicationProduct
	err = json.Unmarshal(body, &products)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	c.cardApplicationProducts = products

	return products, nil
}

func (c *client) CreateCustomerAndCard(ctx context.Context, userID string, args CreateCustomerAndCardArgs) (*Customer, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "gatehub"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "POST",
			Provider: "gatehub",
		})
	}

	endpoint, err := url.JoinPath(c.baseURL, "cards", "v1", "customers")
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	args.Account.WithAccountProductCode(c.cardAccountProductCode)
	body, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req.Header.Set(cardAppIDHeader, c.cardAppID)
	req.Header.Set(managedUserHeader, userID)
	req.Header.Set("Content-Type", "application/json")
	err = c.Sign(ctx, req, time.Now(), body, endpoint)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var customer CustomerResponse
	err = json.Unmarshal(respBody, &customer)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &customer.Customer, nil
}

func (c *client) CreateCustomerDeliveryAddress(ctx context.Context, userID, customerID string, args CreateCustomerDeliveryAddressArgs) (string, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "gatehub"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "POST",
			Provider: "gatehub",
		})
	}

	endpoint, err := url.JoinPath(c.baseURL, "cards", "v1", "customers", customerID, "addresses")
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	req.Header.Set(cardAppIDHeader, c.cardAppID)
	req.Header.Set(managedUserHeader, userID)
	req.Header.Set("Content-Type", "application/json")
	err = c.Sign(ctx, req, time.Now(), body, endpoint)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var cdas []CustomerDeliveryAddress
	err = json.Unmarshal(respBody, &cdas)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	// The newly created address is the last one in the list
	cda := cdas[len(cdas)-1]

	return cda.ID, nil
}

func (c *client) OrderCard(ctx context.Context, userID, accountID string, args OrderCardArgs) (*Card, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "gatehub"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "POST",
			Provider: "gatehub",
		})
	}

	endpoint, err := url.JoinPath(c.baseURL, "cards", "v1", "cards", accountID, "card")
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	args.WithAccountProductCode(c.cardAccountProductCode)
	body, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req.Header.Set(cardAppIDHeader, c.cardAppID)
	req.Header.Set(managedUserHeader, userID)
	req.Header.Set("Content-Type", "application/json")
	err = c.Sign(ctx, req, time.Now(), body, endpoint)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var card Card
	err = json.Unmarshal(respBody, &card)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &card, nil
}

func (c *client) CreatePlasticForCard(ctx context.Context, userID, cardID string) error {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "gatehub"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "POST",
			Provider: "gatehub",
		})
	}

	endpoint, err := url.JoinPath(c.baseURL, "cards", "v1", "cards", cardID, "plastic")
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(nil))
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(managedUserHeader, userID)
	req.Header.Set(cardAppIDHeader, c.cardAppID)
	err = c.Sign(ctx, req, time.Now(), nil, endpoint)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	return nil
}

func (c *client) GetCardToken(ctx context.Context, userID string, tokenType string, args GetCardTokenArgs) (*TokenResponse, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "gatehub"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "POST",
			Provider: "gatehub",
		})
	}

	endpoint, err := url.JoinPath(c.baseURL, "cards", "v1", "token", tokenType)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(managedUserHeader, userID)
	req.Header.Set(cardAppIDHeader, c.cardAppID)
	err = c.Sign(ctx, req, time.Now(), body, endpoint)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var tr TokenResponse
	err = json.Unmarshal(respBody, &tr)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &tr, nil
}

func (c *client) FreezeCard(ctx context.Context, userID, cardID string, args FreezeCardArgs) error {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "PUT"
		meta.Provider = "gatehub"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "PUT",
			Provider: "gatehub",
		})
	}

	endpoint, err := url.Parse(fmt.Sprintf("%s/cards/v1/cards/%s/lock", c.baseURL, cardID))
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	q := endpoint.Query()
	q.Set("reasonCode", args.ReasonCode)
	endpoint.RawQuery = q.Encode()

	body, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint.String(), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(managedUserHeader, userID)
	req.Header.Set(cardAppIDHeader, c.cardAppID)
	err = c.Sign(ctx, req, time.Now(), body, endpoint.String())
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	return nil
}

func (c *client) UnfreezeCard(ctx context.Context, userID, cardID string, args UnfreezeCardArgs) error {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "PUT"
		meta.Provider = "gatehub"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "PUT",
			Provider: "gatehub",
		})
	}

	endpoint, err := url.JoinPath(c.baseURL, "cards", "v1", "cards", cardID, "unlock")
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(managedUserHeader, userID)
	req.Header.Set(cardAppIDHeader, c.cardAppID)
	err = c.Sign(ctx, req, time.Now(), body, endpoint)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	return nil
}

func (c *client) BlockCard(ctx context.Context, userID, cardID string, args BlockCardArgs) error {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "PUT"
		meta.Provider = "gatehub"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "PUT",
			Provider: "gatehub",
		})
	}

	endpoint, err := url.Parse(fmt.Sprintf("%s/cards/v1/cards/%s/block", c.baseURL, cardID))
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	q := endpoint.Query()
	q.Set("reasonCode", args.ReasonCode)
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint.String(), bytes.NewBuffer(nil))
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(managedUserHeader, userID)
	req.Header.Set(cardAppIDHeader, c.cardAppID)
	err = c.Sign(ctx, req, time.Now(), nil, endpoint.String())
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	return nil
}

func (c *client) CloseCard(ctx context.Context, userID, cardID string, args CloseCardArgs) error {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "DELETE"
		meta.Provider = "gatehub"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "DELETE",
			Provider: "gatehub",
		})
	}

	endpoint, err := url.Parse(fmt.Sprintf("%s/cards/v1/cards/%s/card", c.baseURL, cardID))
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	q := endpoint.Query()
	q.Set("reasonCode", args.ReasonCode)
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint.String(), bytes.NewBuffer(nil))
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(managedUserHeader, userID)
	req.Header.Set(cardAppIDHeader, c.cardAppID)
	err = c.Sign(ctx, req, time.Now(), nil, endpoint.String())
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	return nil
}

func (c *client) GetCardDetails(ctx context.Context, userID, cardID string) (*Card, error) {
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

	endpoint, err := url.JoinPath(c.baseURL, "cards", "v1", "cards", cardID, "card")
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req.Header.Set(managedUserHeader, userID)
	req.Header.Set(cardAppIDHeader, c.cardAppID)

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
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var res Card
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &res, nil
}

func (c *client) GetCardTransaction(ctx context.Context, userID, txID string) (*CardTransaction, error) {
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

	endpoint, err := url.JoinPath(c.baseURL, "cards", "v1", "transactions", txID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(managedUserHeader, userID)
	req.Header.Set(cardAppIDHeader, c.cardAppID)
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
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var ct CardTransaction
	err = json.Unmarshal(respBody, &ct)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &ct, nil
}
