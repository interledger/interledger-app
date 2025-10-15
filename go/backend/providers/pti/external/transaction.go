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

	httplog "gitlab.com/fynbos/backend/providers/http"
)

func (c client) StartTransferAssessment(ctx context.Context, args TransferArgs) (*IDResponse, error) {
	requestID := args.RequestID
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "pti"
		meta.Context = strings.Join(
			[]string{fmt.Sprintf("%s=%s", ptiScenarioIDHeader, args.ScenarioID), fmt.Sprintf("%s=%s", ptiRequestIDHeader, requestID), meta.Context},
			",",
		)
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "POST",
			Provider: "pti",
			Context:  fmt.Sprintf("%s=%s,%s=%s", ptiScenarioIDHeader, args.ScenarioID, ptiRequestIDHeader, requestID),
		})
	}

	url, err := url.JoinPath(c.baseURL, "transactions", "assessments")
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	payload, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Add(ptiScenarioIDHeader, args.ScenarioID)
	req.Header.Add(ptiRequestIDHeader, requestID)
	req.Header.Add(ptiClientIDHeader, c.clientID)
	req.Header.Add(ptiSessionIDHeader, args.SessionID)
	req.Header.Add(ptiDisableWebhookHeader, fmt.Sprintf("%t", args.DisableWebhook))
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
	defer resp.Body.Close()

	var idResp IDResponse
	err = json.Unmarshal(body, &idResp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &idResp, nil
}

func (c client) GetTransactionAssessment(ctx context.Context, requestID string) (*TransactionAssessment, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "GET"
		meta.Provider = "pti"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "GET",
			Provider: "pti",
		})
	}

	url, err := url.JoinPath(c.baseURL, "transactions", "assessments", requestID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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

	var assessment TransactionAssessment
	err = json.Unmarshal(body, &assessment)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &assessment, nil
}

func (c client) CreateTransfer(ctx context.Context, args TransferArgs) (*IDResponse, error) {
	requestID := args.RequestID
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "pti"
		meta.Context = strings.Join(
			[]string{fmt.Sprintf("%s=%s", ptiScenarioIDHeader, args.ScenarioID), fmt.Sprintf("%s=%s", ptiRequestIDHeader, requestID), meta.Context},
			",",
		)
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "POST",
			Provider: "pti",
			Context:  fmt.Sprintf("%s=%s,%s=%s", ptiScenarioIDHeader, args.ScenarioID, ptiRequestIDHeader, requestID),
		})
	}

	url, err := url.JoinPath(c.baseURL, "transactions", "transfers")
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	payload, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Add(ptiScenarioIDHeader, args.ScenarioID)
	req.Header.Add(ptiRequestIDHeader, requestID)
	req.Header.Add(ptiClientIDHeader, c.clientID)
	req.Header.Add(ptiSessionIDHeader, args.SessionID)
	req.Header.Add(ptiDisableWebhookHeader, fmt.Sprintf("%t", args.DisableWebhook))
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
	defer resp.Body.Close()

	var idResp IDResponse
	err = json.Unmarshal(body, &idResp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &idResp, nil
}

func (c client) GetTransaction(ctx context.Context, requestID string) (*TransactionStatus, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "GET"
		meta.Provider = "pti"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "GET",
			Provider: "pti",
		})
	}

	url, err := url.JoinPath(c.baseURL, "transactions", requestID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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

	var trx TransactionStatus
	err = json.Unmarshal(body, &trx)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &trx, nil
}

func (c client) WalletDeposit(ctx context.Context, args DepositArgs) (string, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = http.MethodPost
		meta.Provider = ptiProviderName
		meta.Context = strings.Join(
			[]string{fmt.Sprintf("%s=%s", ptiScenarioIDHeader, args.ScenarioID), fmt.Sprintf("%s=%s", ptiRequestIDHeader, args.RequestID), meta.Context},
			",",
		)
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   http.MethodPost,
			Provider: ptiProviderName,
			Context:  fmt.Sprintf("%s=%s,%s=%s", ptiScenarioIDHeader, args.ScenarioID, ptiRequestIDHeader, args.RequestID),
		})
	}

	url, err := url.JoinPath(c.baseURL, "transactions", "deposits")
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	// sourcePaymentMethodType := "CREDIT_CARD"
	// sourcePaymentInformationType := "ENCRYPTED_CREDIT_CARD"
	// if args.ExternalPaymentMethodType == "bank" {
	// 	sourcePaymentMethodType = "ACH"
	// 	sourcePaymentInformationType = "BANK_ACCOUNT"
	// }

	sourcePaymentMethodType := "ACH"
	sourcePaymentInformationType := "BANK_ACCOUNT"

	reqArgs := internalCreateDepositArgs{
		Initiator: User{
			ID:   args.UserID,
			Type: "PERSON",
		},
		SourceMethod: SourceMethod{
			Currency: args.Amount.Currency.String(),
			PaymentInformation: PaymentInformation{
				Type:              sourcePaymentInformationType,
				ID:                args.ExternalPaymentMethodID,
				AccountHolderName: args.AccountHolderName,
			},
			PaymentMethodType: sourcePaymentMethodType,
		},
		DestinationMethod: DestinationMethod{
			PaymentMethodType: "WALLET", // checked
			PaymentInformation: DestinationInformation{
				Type: "WALLET",
				ID:   args.ExternalWalletID,
			},
		},
		Amount:    args.Amount.Float64(),
		USDAmount: args.Amount.Float64(),
		Type:      "DEPOSIT",
		Date:      time.Now().Format(time.RFC3339),
	}

	payload, err := json.Marshal(reqArgs)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Add(ptiScenarioIDHeader, args.ScenarioID)
	req.Header.Add(ptiRequestIDHeader, args.RequestID)
	req.Header.Add(ptiClientIDHeader, c.clientID)
	req.Header.Add(ptiSessionIDHeader, args.SessionID)
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

	var txResp CreateTxResponse
	err = json.Unmarshal(body, &txResp)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	return txResp.ID, nil
}

func (c client) WalletWithdrawal(ctx context.Context, args WithdrawalArgs) (string, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "pti"
		meta.Context = strings.Join(
			[]string{fmt.Sprintf("%s=%s", ptiScenarioIDHeader, args.ScenarioID), fmt.Sprintf("%s=%s", ptiRequestIDHeader, args.RequestID), meta.Context},
			",",
		)
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "POST",
			Provider: "pti",
			Context:  fmt.Sprintf("%s=%s,%s=%s", ptiScenarioIDHeader, args.ScenarioID, ptiRequestIDHeader, args.RequestID),
		})
	}

	url, err := url.JoinPath(c.baseURL, "transactions", "withdrawals")
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	reqArgs := InternalCreateWithdrawalArgs{
		Initiator: Initiator{
			UserID: args.UserID,
			Type:   "PERSON",
		},
		SourceMethod: SourceMethod{
			PaymentInformation: PaymentInformation{
				Type:     "WALLET",
				WalletID: args.ExternalWalletID,
			},
			PaymentMethodType: "WALLET",
		},
		DestinationMethod: WithdrawalDestinationMethod{
			Currency:          args.Amount.Currency.String(),
			PaymentMethodType: "ACH",
			PaymentInformation: PaymentInformation{
				Type: "BANK_ACCOUNT",
				ID:   args.ExternalBankAccountID,
			},
		},
		Amount:    args.Amount.Float64(),
		USDAmount: args.Amount.Float64(),
		Type:      "WITHDRAWAL",
		Date:      time.Now().Format(time.RFC3339),
	}

	payload, err := json.Marshal(reqArgs)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Add(ptiScenarioIDHeader, args.ScenarioID)
	req.Header.Add(ptiRequestIDHeader, args.RequestID)
	req.Header.Add(ptiClientIDHeader, c.clientID)
	req.Header.Add(ptiSessionIDHeader, args.SessionID)
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

	var txResp CreateTxResponse
	err = json.Unmarshal(body, &txResp)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	return txResp.ID, nil
}

func (c client) UpdateTransactionStatus(ctx context.Context, args UpdateTxStatusArgs) (string, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "pti"
		meta.Context = strings.Join(
			[]string{fmt.Sprintf("%s=%s", ptiRequestIDHeader, args.RequestID), meta.Context},
			",",
		)
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "POST",
			Provider: "pti",
			Context:  fmt.Sprintf("%s=%s", ptiRequestIDHeader, args.RequestID),
		})
	}

	url, err := url.JoinPath(c.baseURL, "transactions", args.RequestID, "updates")
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	payload, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Add(ptiRequestIDHeader, args.RequestID)
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

	var txResp CreateTxResponse
	err = json.Unmarshal(body, &txResp)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	return txResp.ID, nil
}
