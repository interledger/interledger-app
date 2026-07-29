package provision

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/interledger/interledger-app/go/performance/client"
	pb "github.com/interledger/interledger-app/go/proto/backend/v1"
)

func fund(ctx context.Context, w *client.Wallet, spec countrySpec, targetMajor int64) (int64, []string, error) {
	switch spec.provider {
	case "xago":
		return fundXago(ctx, w, spec, targetMajor)
	case "gatehub":
		return fundGatehub(ctx, w, spec, targetMajor)
	case "pti":
		return fundPTI(ctx, w, spec, targetMajor)
	default:
		return 0, nil, fmt.Errorf("unsupported country spec %s", spec.code)
	}
}

func fundXago(ctx context.Context, w *client.Wallet, spec countrySpec, targetMajor int64) (int64, []string, error) {
	if _, err := w.AddXagoBalanceAccount(ctx, &pb.AddXagoBalanceAccountRequest{
		CurrencyCode: spec.currencyCode,
		Nickname:     spec.asset + " Balance",
		Title:        spec.asset + " Balance",
	}); err != nil {
		return 0, nil, fmt.Errorf("AddXagoBalanceAccount: %w", client.Classify("signup", err))
	}

	var notes []string
	for i := 0; i < 10; i++ {
		if err := w.DepositTestXago(ctx); err != nil {
			return 0, nil, fmt.Errorf("DepositTestXago: %w", client.Classify("signup", err))
		}
		balance, err := waitForBalance(ctx, w, spec.asset+" Balance")
		if err != nil {
			return 0, nil, err
		}
		if balance >= targetMajor {
			return balance, notes, nil
		}
		notes = append(notes, fmt.Sprintf("xago balance %d/%d", balance, targetMajor))
		time.Sleep(500 * time.Millisecond)
	}

	balance, err := w.GetBalances(ctx)
	if err != nil {
		return 0, nil, fmt.Errorf("GetBalances: %w", err)
	}
	for _, b := range balance {
		if b.GetBalance() != nil && b.GetBalance().GetAsset() == spec.asset {
			return b.GetBalance().GetAmount(), notes, nil
		}
	}
	return 0, notes, nil
}

func fundGatehub(ctx context.Context, w *client.Wallet, spec countrySpec, targetMajor int64) (int64, []string, error) {
	if _, err := w.GetGatehubOnboardingWidget(ctx, &pb.Empty{}); err != nil {
		return 0, nil, fmt.Errorf("GetGatehubOnboardingWidget: %w", client.Classify("signup", err))
	}

	var notes []string
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		balances, err := w.GetBalances(ctx)
		if err != nil {
			return 0, nil, fmt.Errorf("GetBalances: %w", err)
		}
		for _, b := range balances {
			if b.GetBalance() != nil && strings.EqualFold(b.GetBalance().GetAsset(), spec.asset) {
				if b.GetBalance().GetAmount() >= targetMajor {
					return b.GetBalance().GetAmount(), notes, nil
				}
				break
			}
		}
		notes = append(notes, fmt.Sprintf("gatehub balance pending (%d/%d)", targetMajor, targetMajor))
		time.Sleep(1 * time.Second)
	}
	return 0, notes, fmt.Errorf("gatehub balance never reached target")
}

func fundPTI(ctx context.Context, w *client.Wallet, spec countrySpec, targetMajor int64) (int64, []string, error) {
	widget, err := w.GetKYCProviderWidget(ctx, &pb.GetKYCProviderWidgetRequest{IdempotencyKey: uuid.NewString()})
	if err != nil {
		return 0, nil, fmt.Errorf("GetKYCProviderWidget: %w", client.Classify("signup", err))
	}
	if _, err := w.SetKYCStatusPending(ctx, &pb.Empty{}); err != nil {
		return 0, nil, fmt.Errorf("SetKYCStatusPending: %w", client.Classify("signup", err))
	}
	if widget.GetPtiWidget() == nil {
		return 0, nil, fmt.Errorf("PTI widget missing pti details")
	}

	if err := mockPTIAssessment(widget.GetPtiWidget().GetUserId()); err != nil {
		return 0, nil, fmt.Errorf("mock PTI assessment: %w", err)
	}

	deadline := time.Now().Add(30 * time.Second)
	var balanceLA string
	for time.Now().Before(deadline) {
		balances, err := w.GetPtiBalances(ctx)
		if err != nil {
			return 0, nil, fmt.Errorf("GetPtiBalances: %w", err)
		}
		for _, b := range balances.GetBalances() {
			if b.GetLinkedAccount() != "" {
				balanceLA = b.GetLinkedAccount()
				break
			}
		}
		if balanceLA != "" {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if balanceLA == "" {
		return 0, nil, fmt.Errorf("PTI balance linked account did not appear")
	}

	bankLA, err := w.CreatePtiBankAccount(ctx, &pb.CreatePtiBankAccountRequest{
		BankName:      "Perf Bank",
		AccountNumber: "123456789",
		RoutingNumber: "987654321",
		AccountType:   "CHECKING",
	})
	if err != nil {
		return 0, nil, fmt.Errorf("CreatePtiBankAccount: %w", client.Classify("signup", err))
	}
	amount := &pb.Amount{Amount: targetMajor * 100, Asset: spec.asset, AssetScale: 2, Country: spec.country}
	payment, err := w.DepositBalance(ctx, &pb.TransferBalanceRequest{
		FromLinkedAccount: bankLA.GetId(),
		ToLinkedAccount:   balanceLA,
		Amount:            amount,
		Note:              "perf provision",
	})
	if err != nil {
		return 0, nil, fmt.Errorf("DepositBalance: %w", client.Classify("signup", err))
	}
	if _, err := w.PtiCreateDeposit(ctx, &pb.PtiCreateDepositRequest{Id: payment.GetId()}); err != nil {
		return 0, nil, fmt.Errorf("PtiCreateDeposit: %w", client.Classify("signup", err))
	}

	deadline = time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := w.GetPtiBalances(ctx)
		if err != nil {
			return 0, nil, fmt.Errorf("GetPtiBalances: %w", err)
		}
		for _, b := range resp.GetBalances() {
			if b.GetBalance() != nil && b.GetBalance().GetAmount() >= targetMajor*100 {
				return b.GetBalance().GetAmount(), nil, nil
			}
		}
		time.Sleep(1 * time.Second)
	}
	return 0, nil, fmt.Errorf("PTI balance never reached target")
}

func mockPTIAssessment(externalUserID string) error {
	body := map[string]string{"id": externalUserID, "type": "PERSON"}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, "https://mockpti.interledger.test/users/assessments", strings.NewReader(string(payload)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-pti-client-id", "04d3e1b5-96d4-47e4-9eaa-13e9b4b0f219")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}

func generateGatehubSignature(ts, method, url, body, secret string) string {
	msg := fmt.Sprintf("%s|%s|%s|%s", ts, method, url, body)
	hmacHash := hmac.New(sha256.New, []byte(secret))
	_, _ = hmacHash.Write([]byte(msg))
	return hex.EncodeToString(hmacHash.Sum(nil))
}
