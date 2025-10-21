package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	httplog "gitlab.com/fynbos/backend/providers/http"
)

// https://developers.platform.fiant.io/reference/addauser
func (c client) CreateUser(ctx context.Context, args CreateUserArgs) (string, error) {
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

	url, err := url.JoinPath(c.baseURL, "users")
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	payload, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Add(ptiClientIDHeader, c.clientID)
	req.Header.Add("Content-Type", "application/json")
	date := time.Now()
	req.Header.Add("Date", date.Format(http.TimeFormat))
	if err := sign(req, date, payload, c.privateKey, c.publicKeyThumbprint); err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var userResp CreateUserResponse
	err = json.Unmarshal(body, &userResp)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	return userResp.ID, nil
}

// https://developers.platform.fiant.io/reference/getuser
func (c client) GetUser(ctx context.Context, id string) (*User, error) {
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

	url, err := url.JoinPath(c.baseURL, "users", id)
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

	var user User
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &user, nil
}

// https://developers.platform.fiant.io/reference/mergeuserinfo
func (c client) PatchUser(ctx context.Context, args PatchUserArgs) (string, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = http.MethodPatch
		meta.Provider = ptiProviderName
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   http.MethodPatch,
			Provider: ptiProviderName,
		})
	}

	url, err := url.JoinPath(c.baseURL, "users")
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	payload, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Add(ptiClientIDHeader, c.clientID)
	req.Header.Add("Content-Type", "application/json")
	date := time.Now()
	req.Header.Add("Date", date.Format(http.TimeFormat))
	if err := sign(req, date, payload, c.privateKey, c.publicKeyThumbprint); err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var userResp CreateUserResponse
	err = json.Unmarshal(body, &userResp)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	return userResp.ID, nil
}

// https://developers.platform.fiant.io/reference/updateuser
func (c client) PutUser(ctx context.Context, args PutUserArgs) (string, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = http.MethodPut
		meta.Provider = ptiProviderName
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   http.MethodPut,
			Provider: ptiProviderName,
		})
	}

	url, err := url.JoinPath(c.baseURL, "users")
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	payload, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Add(ptiClientIDHeader, c.clientID)
	req.Header.Add("Content-Type", "application/json")
	date := time.Now()
	req.Header.Add("Date", date.Format(http.TimeFormat))
	if err := sign(req, date, payload, c.privateKey, c.publicKeyThumbprint); err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var userResp CreateUserResponse
	err = json.Unmarshal(body, &userResp)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	return userResp.ID, nil
}

// https://developers.platform.fiant.io/reference/getlastkyc
func (c client) GetUserAssessment(ctx context.Context, userID string) (*Assessment, error) {
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

	url, err := url.JoinPath(c.baseURL, "users", userID, "assessments")
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

	var assessment Assessment
	err = json.Unmarshal(body, &assessment)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &assessment, nil
}

// https://developers.platform.fiant.io/reference/startuserassessment
func (c client) StartUserAssessment(ctx context.Context, args StartUserAssessmentArgs) (string, error) {
	requestID := uuid.NewString()
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = http.MethodPost
		meta.Provider = ptiProviderName
		meta.Context = strings.Join(
			[]string{fmt.Sprintf("%s=%s", ptiScenarioIDHeader, args.ScenarioID), fmt.Sprintf("%s=%s", ptiRequestIDHeader, requestID), meta.Context},
			",",
		)
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   http.MethodPost,
			Provider: ptiProviderName,
			Context:  fmt.Sprintf("%s=%s,%s=%s", ptiScenarioIDHeader, args.ScenarioID, ptiRequestIDHeader, requestID),
		})
	}

	if args.ID == "" {
		return "", fmt.Errorf("%w UserID is required", ErrBadRequest)
	}

	url, err := url.JoinPath(c.baseURL, "users", "assessments")
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	payload, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Add(ptiScenarioIDHeader, args.ScenarioID)
	req.Header.Add(ptiRequestIDHeader, requestID)
	req.Header.Add(ptiClientIDHeader, c.clientID)
	req.Header.Add("Content-Type", "application/json")
	date := time.Now()
	req.Header.Add("Date", date.Format(http.TimeFormat))
	if err := sign(req, date, payload, c.privateKey, c.publicKeyThumbprint); err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var userResp CreateUserResponse
	err = json.Unmarshal(body, &userResp)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	return userResp.ID, nil
}

// https://developers.platform.fiant.io/reference/getuserpaymentinformations
func (c client) GetUsersPaymentInformation(ctx context.Context, userID, id string) (json.RawMessage, error) {
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

	url, err := url.JoinPath(c.baseURL, "users", userID, "payment-information", id)
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

	return body, nil
}

// https://developers.platform.fiant.io/reference/getuserpaymentinformations
func (c client) CreateBankAccount(ctx context.Context, userID string, args BankAccountPaymentInformation) (*BankAccountPaymentInformation, error) {
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

	url, err := url.JoinPath(c.baseURL, "users", userID, "payment-information")
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

	var storedBankAccount BankAccountPaymentInformation
	err = json.Unmarshal(body, &storedBankAccount)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &storedBankAccount, nil
}
