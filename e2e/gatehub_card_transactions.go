package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const mockgatehubCardTxBaseURL = "https://mockgatehub.interledger.test"

const (
	cardTxUserName    = "card-tx-user"
	cardTxEmailSuffix = "cardtx@test.example.com"
	cardTxPassword    = "CardTxTest2024!"
)

// loadCardTxUserDetails populates sc.userDetails for the card-tx user using
// the scenario's own testIdentifier, giving each scenario an isolated user.
func (sc *E2EContext) loadCardTxUserDetails() {
	if sc.userDetails == nil {
		sc.userDetails = make(map[string]*UserDetails)
	}
	sc.userDetails[cardTxUserName] = &UserDetails{
		Fields: map[string]string{
			"emailSuffix": cardTxEmailSuffix,
			"email":       sc.testIdentifier + "-" + cardTxEmailSuffix,
			"password":    cardTxPassword,
			"country":     "Germany",
			"firstName":   "Card",
			"lastName":    "Tester",
		},
	}
	sc.currentUser = cardTxUserName
	sc.password = cardTxPassword
}

// iEnsureCardTxUserIsSetUp is the Background step for card-transaction scenarios.
// It creates a fresh user per scenario, funding them and ordering a card.
func (sc *E2EContext) iEnsureCardTxUserIsSetUp() error {
	sc.loadCardTxUserDetails()
	return sc.runCardTxUserSetup()
}

func (sc *E2EContext) runCardTxUserSetup() error {
	if err := sc.iCompleteMinimalKYCFlow(cardTxUserName); err != nil {
		return fmt.Errorf("card-tx setup: KYC: %w", err)
	}
	if err := sc.iDepositViATheDepositIframeAsUser("200", "EUR", cardTxUserName); err != nil {
		return fmt.Errorf("card-tx setup: deposit: %w", err)
	}
	return sc.iOrderAPhysicalCardFor(cardTxUserName)
}

// mockgatehubHTTPClient returns an HTTP client that skips TLS verification
// and does not follow redirects (the MockGatehub UI endpoints return 302 on success).
func mockgatehubHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// triggerCardTransaction POSTs to MockGatehub's /ui/actions/card-transaction endpoint,
// creating a card transaction and firing the signed webhook to the backend.
// cardTxJSON is the raw transaction JSON; MockGatehub will overwrite transactionId with a new UUID.
func (sc *E2EContext) triggerCardTransaction(gatehubUserID, cardID, title, body, cardTxJSON string) error {
	baseURL := mockgatehubCardTxBaseURL
	if sc.mockgatehubBaseURL != "" {
		baseURL = sc.mockgatehubBaseURL
	}

	formData := url.Values{
		"userID":       {gatehubUserID},
		"cardID":       {cardID},
		"title":        {title},
		"body":         {body},
		"cardTxObject": {cardTxJSON},
	}

	reqURL := baseURL + "/ui/actions/card-transaction"
	req, err := http.NewRequestWithContext(context.Background(), "POST", reqURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("triggerCardTransaction: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := mockgatehubHTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("triggerCardTransaction: request: %w", err)
	}
	defer resp.Body.Close()

	// 302 redirect = success; any non-redirect is an error
	if resp.StatusCode != http.StatusSeeOther && resp.StatusCode != http.StatusFound {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("triggerCardTransaction: unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// setCardTransactionStatus updates a MockGatehub card transaction's status to "COMPLETED" or "FAILED".
func (sc *E2EContext) setCardTransactionStatus(txID, status, gatehubUserID string) error {
	baseURL := mockgatehubCardTxBaseURL
	if sc.mockgatehubBaseURL != "" {
		baseURL = sc.mockgatehubBaseURL
	}

	formData := url.Values{
		"txID":   {txID},
		"status": {status},
		"userID": {gatehubUserID},
	}

	reqURL := baseURL + "/ui/actions/card-transaction/status"
	req, err := http.NewRequestWithContext(context.Background(), "POST", reqURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("setCardTransactionStatus: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := mockgatehubHTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("setCardTransactionStatus: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther && resp.StatusCode != http.StatusFound {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("setCardTransactionStatus: unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// getLatestGatehubCardTxID returns the most recently created gatehub_card_transactions.id for the user.
func (sc *E2EContext) getLatestGatehubCardTxID(email string) (string, error) {
	if err := sc.ensureDB(); err != nil {
		return "", fmt.Errorf("getLatestGatehubCardTxID: %w", err)
	}

	gatehubUserID, err := sc.getGatehubUserIDByEmail(email)
	if err != nil {
		return "", fmt.Errorf("getLatestGatehubCardTxID: %w", err)
	}

	var txID string
	err = sc.db.QueryRow(`
		SELECT id FROM gatehub_card_transactions
		WHERE user_id = $1::uuid
		ORDER BY created_at DESC LIMIT 1
	`, gatehubUserID).Scan(&txID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("getLatestGatehubCardTxID: no card transactions found for user %s", email)
		}
		return "", fmt.Errorf("getLatestGatehubCardTxID: query error: %w", err)
	}

	return txID, nil
}

// waitForGatehubCardTx polls the backend DB until at least `atLeast` gatehub_card_transactions rows exist for the user.
func (sc *E2EContext) waitForGatehubCardTx(email string, atLeast int, timeout time.Duration) error {
	if err := sc.ensureDB(); err != nil {
		return fmt.Errorf("waitForGatehubCardTx: %w", err)
	}

	gatehubUserID, err := sc.getGatehubUserIDByEmail(email)
	if err != nil {
		return fmt.Errorf("waitForGatehubCardTx: %w", err)
	}

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		var total int
		if queryErr := sc.db.QueryRow(`
			SELECT COUNT(*) FROM gatehub_card_transactions WHERE user_id = $1::uuid
		`, gatehubUserID).Scan(&total); queryErr != nil {
			debugPrintf("   ⚠️  waitForGatehubCardTx: query error: %v\n", queryErr)
			time.Sleep(time.Second)
			continue
		}

		if total >= atLeast {
			debugPrintf("   ✓ Found %d gatehub card transaction(s) for user %s\n", total, email)
			return nil
		}

		debugPrintf("   ⏳ Waiting for %d gatehub card tx(s), have %d for user %s\n", atLeast, total, email)
		time.Sleep(time.Second)
	}

	return fmt.Errorf("waitForGatehubCardTx: timed out waiting for %d card transaction(s) for user %s", atLeast, email)
}

type backendCardTx struct {
	State               string
	Note                sql.NullString
	ExchangeRateApplied sql.NullString
}

// getLatestBackendCardTx returns the most recently created transactions row for this user with type = 'card_transaction'.
func (sc *E2EContext) getLatestBackendCardTx(email string) (*backendCardTx, error) {
	if err := sc.ensureDB(); err != nil {
		return nil, fmt.Errorf("getLatestBackendCardTx: %w", err)
	}

	walletID, err := sc.getWalletIDByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("getLatestBackendCardTx: %w", err)
	}

	var tx backendCardTx
	err = sc.db.QueryRow(`
		SELECT state, note, exchange_rate_applied
		FROM transactions
		WHERE wallet_id = $1 AND type = 'card_transaction'
		ORDER BY created_at DESC LIMIT 1
	`, walletID).Scan(&tx.State, &tx.Note, &tx.ExchangeRateApplied)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("getLatestBackendCardTx: no card transactions in backend DB for user %s", email)
		}
		return nil, fmt.Errorf("getLatestBackendCardTx: query error: %w", err)
	}

	return &tx, nil
}

// waitForBackendCardTx polls until at least `atLeast` transactions with type='card_transaction' exist for the user.
func (sc *E2EContext) waitForBackendCardTx(email string, atLeast int, timeout time.Duration) error {
	if err := sc.ensureDB(); err != nil {
		return fmt.Errorf("waitForBackendCardTx: %w", err)
	}

	walletID, err := sc.getWalletIDByEmail(email)
	if err != nil {
		return fmt.Errorf("waitForBackendCardTx: %w", err)
	}

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		var total int
		if queryErr := sc.db.QueryRow(`
			SELECT COUNT(*) FROM transactions WHERE wallet_id = $1 AND type = 'card_transaction'
		`, walletID).Scan(&total); queryErr != nil {
			debugPrintf("   ⚠️  waitForBackendCardTx: query error: %v\n", queryErr)
			time.Sleep(time.Second)
			continue
		}

		if total >= atLeast {
			debugPrintf("   ✓ Found %d backend card transaction(s) for user %s\n", total, email)
			return nil
		}

		debugPrintf("   ⏳ Waiting for %d backend card tx(s), have %d for user %s\n", atLeast, total, email)
		time.Sleep(time.Second)
	}

	return fmt.Errorf("waitForBackendCardTx: timed out waiting for %d card transaction(s) for user %s", atLeast, email)
}

// waitForBackendCardTxState polls until the latest card transaction for the user has the expected state.
func (sc *E2EContext) waitForBackendCardTxState(email, expectedState string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	lastState := ""

	for time.Now().Before(deadline) {
		tx, err := sc.getLatestBackendCardTx(email)
		if err != nil {
			debugPrintf("   ⚠️  waitForBackendCardTxState: query error: %v\n", err)
			time.Sleep(time.Second)
			continue
		}
		if tx.State == expectedState {
			return nil
		}
		lastState = tx.State
		debugPrintf("   ⏳ Waiting for state '%s', have '%s' for user %s\n", expectedState, tx.State, email)
		time.Sleep(time.Second)
	}
	return fmt.Errorf("waitForBackendCardTxState: timed out waiting for state '%s', last state '%s' for user %s", expectedState, lastState, email)
}

// buildCardTxSetup is a helper shared by all trigger steps to get gatehubUserID and cardID.
func (sc *E2EContext) buildCardTxSetup(userName string) (email, gatehubUserID, cardID string, err error) {
	email, err = sc.getEmailForUser(userName)
	if err != nil {
		return
	}
	gatehubUserID, err = sc.getGatehubUserIDByEmail(email)
	if err != nil {
		return
	}
	cardID, err = sc.getActiveCardID(gatehubUserID)
	return
}

// --- Mock JSON payloads for card transactions ---

const txPurchaseJSON = `{
  "operation": 0,
  "transactionClassification": "Authorization",
  "type": 0,
  "isTrxAmountConverted": false,
  "txStatus": "PROCESSING",
  "ghResponseCode": "OK",
  "responseCode": "OK",
  "MastercardConversion": null,
  "transactionAmount": "1.89",
  "transactionCurrency": "EUR",
  "billingAmount": "1.89",
  "billingCurrency": "EUR",
  "vaultId": 1,
  "cardId": 277,
  "refTransactionId": null,
  "ghResponseDescription": "Approved",
  "terminalId": "C0A17138",
  "cardScheme": 3,
  "source": "P",
  "mcc": "5399",
  "merchantName": "SPAR IVANCNA GORICA",
  "merchantCountry": "SVN",
  "merchantCity": "IVANCNA GORIC",
  "merchantZip": "1295",
  "merchantStreet": "STANTETOVA ULICA 1A",
  "merchantId": "002295628",
  "transactionDateTime": "2024-12-19T13:15:38.000Z",
  "processDateTime": "2024-12-19T13:15:38.000Z"
}`

const txPurchaseWithFXJSON = `{
  "operation": 0,
  "transactionClassification": "Authorization",
  "type": 0,
  "isTrxAmountConverted": true,
  "txStatus": "PROCESSING",
  "ghResponseCode": "OK",
  "responseCode": "OK",
  "MastercardConversion": {"convRate": "0.1934", "refConRate": null, "refConfRateDiff": "0.66"},
  "transactionAmount": "50",
  "transactionCurrency": "RON",
  "billingAmount": "9.67",
  "billingCurrency": "EUR",
  "vaultId": 1,
  "cardId": 277,
  "refTransactionId": null,
  "ghResponseDescription": "Approved",
  "terminalId": "C0A17138",
  "cardScheme": 3,
  "source": "P",
  "mcc": "5399",
  "merchantName": "SPAR IVANCNA GORICA",
  "merchantCountry": "SVN",
  "merchantCity": "IVANCNA GORIC",
  "merchantZip": "1295",
  "merchantStreet": "STANTETOVA ULICA 1A",
  "merchantId": "002295628",
  "transactionDateTime": "2024-12-19T13:15:38.000Z",
  "processDateTime": "2024-12-19T13:15:38.000Z"
}`

const txATMWithdrawalJSON = `{
  "operation": 0,
  "transactionClassification": "Authorization",
  "type": 1,
  "isTrxAmountConverted": false,
  "txStatus": "PROCESSING",
  "ghResponseCode": "OK",
  "responseCode": "OK",
  "MastercardConversion": null,
  "transactionAmount": "50.00",
  "transactionCurrency": "EUR",
  "billingAmount": "50.00",
  "billingCurrency": "EUR",
  "vaultId": 1,
  "cardId": 277,
  "refTransactionId": null,
  "ghResponseDescription": "Approved",
  "terminalId": "ATM17138",
  "cardScheme": 3,
  "source": "P",
  "mcc": "6011",
  "merchantName": "RAIFFEISEN ATM",
  "merchantCountry": "SVN",
  "merchantCity": "LJUBLJANA",
  "merchantZip": "1000",
  "merchantStreet": "SLOVENSKA 1",
  "merchantId": "ATM0001",
  "transactionDateTime": "2024-12-19T13:15:38.000Z",
  "processDateTime": "2024-12-19T13:15:38.000Z"
}`

const txCashAdvanceJSON = `{
  "operation": 0,
  "transactionClassification": "Authorization",
  "type": 17,
  "isTrxAmountConverted": false,
  "txStatus": "PROCESSING",
  "ghResponseCode": "OK",
  "responseCode": "OK",
  "MastercardConversion": null,
  "transactionAmount": "20.00",
  "transactionCurrency": "EUR",
  "billingAmount": "20.00",
  "billingCurrency": "EUR",
  "vaultId": 1,
  "cardId": 277,
  "refTransactionId": null,
  "ghResponseDescription": "Approved",
  "terminalId": "C0A17138",
  "cardScheme": 3,
  "source": "P",
  "mcc": "6011",
  "merchantName": "RAIFFEISEN BANK",
  "merchantCountry": "SVN",
  "merchantCity": "LJUBLJANA",
  "merchantZip": "1000",
  "merchantStreet": "SLOVENSKA 1",
  "merchantId": "BANK0001",
  "transactionDateTime": "2024-12-19T13:15:38.000Z",
  "processDateTime": "2024-12-19T13:15:38.000Z"
}`

const txPreauthorizationJSON = `{
  "operation": 0,
  "transactionClassification": "Authorization",
  "type": 101,
  "isTrxAmountConverted": false,
  "txStatus": "PROCESSING",
  "ghResponseCode": "OK",
  "responseCode": "OK",
  "MastercardConversion": null,
  "transactionAmount": "5.00",
  "transactionCurrency": "EUR",
  "billingAmount": "5.00",
  "billingCurrency": "EUR",
  "vaultId": 1,
  "cardId": 277,
  "refTransactionId": null,
  "ghResponseDescription": "Approved",
  "terminalId": "C0A17138",
  "cardScheme": 3,
  "source": "P",
  "mcc": "5399",
  "merchantName": "SPAR IVANCNA GORICA",
  "merchantCountry": "SVN",
  "merchantCity": "IVANCNA GORIC",
  "merchantZip": "1295",
  "merchantStreet": "STANTETOVA ULICA 1A",
  "merchantId": "002295628",
  "transactionDateTime": "2024-12-19T13:15:38.000Z",
  "processDateTime": "2024-12-19T13:15:38.000Z"
}`

const txTransferFromAccountJSON = `{
  "operation": 0,
  "transactionClassification": "Authorization",
  "type": 108,
  "isTrxAmountConverted": false,
  "txStatus": "PROCESSING",
  "ghResponseCode": "OK",
  "responseCode": "OK",
  "MastercardConversion": null,
  "transactionAmount": "10.00",
  "transactionCurrency": "EUR",
  "billingAmount": "10.00",
  "billingCurrency": "EUR",
  "vaultId": 1,
  "cardId": 277,
  "refTransactionId": null,
  "ghResponseDescription": "Approved",
  "terminalId": "C0A17138",
  "cardScheme": 3,
  "source": "P",
  "transactionDateTime": "2024-12-19T13:15:38.000Z",
  "processDateTime": "2024-12-19T13:15:38.000Z"
}`

const txTransferToAccountJSON = `{
  "operation": 1,
  "transactionClassification": "Authorization",
  "type": 107,
  "isTrxAmountConverted": false,
  "txStatus": "ACQUIRED",
  "ghResponseCode": "OK",
  "responseCode": "OK",
  "MastercardConversion": null,
  "transactionAmount": "2",
  "transactionCurrency": "EUR",
  "billingAmount": "2",
  "billingCurrency": "EUR",
  "vaultId": 1,
  "cardId": 277,
  "refTransactionId": null,
  "ghResponseDescription": "Approved",
  "terminalId": "C0A17138",
  "cardScheme": 3,
  "source": "P",
  "mcc": "5399",
  "merchantName": "SPAR IVANCNA GORICA",
  "merchantCountry": "SVN",
  "merchantCity": "IVANCNA GORIC",
  "merchantZip": "1295",
  "merchantStreet": "STANTETOVA ULICA 1A",
  "merchantId": "002295628",
  "transactionDateTime": "2024-12-19T13:15:38.000Z",
  "processDateTime": "2024-12-19T13:15:38.000Z"
}`

// PLACEHOLDER_REF_TX_ID must be replaced with the real refTransactionId.
const txPurchaseReversalJSON = `{
  "operation": 1,
  "transactionClassification": "Reversal",
  "type": 0,
  "isTrxAmountConverted": false,
  "txStatus": "PROCESSING",
  "ghResponseCode": "OK",
  "responseCode": "OK",
  "MastercardConversion": null,
  "transactionAmount": "1.89",
  "transactionCurrency": "EUR",
  "billingAmount": "1.89",
  "billingCurrency": "EUR",
  "vaultId": 1,
  "cardId": 277,
  "refTransactionId": "PLACEHOLDER_REF_TX_ID",
  "ghResponseDescription": "Approved",
  "terminalId": "C0A17138",
  "cardScheme": 3,
  "source": "P",
  "mcc": "5399",
  "merchantName": "SPAR IVANCNA GORICA",
  "merchantCountry": "SVN",
  "merchantCity": "IVANCNA GORIC",
  "merchantZip": "1295",
  "merchantStreet": "STANTETOVA ULICA 1A",
  "merchantId": "002295628",
  "transactionDateTime": "2024-12-19T13:15:38.000Z",
  "processDateTime": "2024-12-19T13:15:38.000Z"
}`

const txPreauthIncrementalNoRefJSON = `{
  "operation": 0,
  "transactionClassification": "Authorization",
  "type": 102,
  "isTrxAmountConverted": false,
  "txStatus": "PROCESSING",
  "ghResponseCode": "OK",
  "responseCode": "OK",
  "MastercardConversion": null,
  "transactionAmount": "5.00",
  "transactionCurrency": "EUR",
  "billingAmount": "5.00",
  "billingCurrency": "EUR",
  "vaultId": 1,
  "cardId": 277,
  "refTransactionId": null,
  "ghResponseDescription": "Approved",
  "terminalId": "C0A17138",
  "cardScheme": 3,
  "source": "P",
  "mcc": "5399",
  "merchantName": "SPAR IVANCNA GORICA",
  "merchantCountry": "SVN",
  "merchantCity": "IVANCNA GORIC",
  "merchantZip": "1295",
  "merchantStreet": "STANTETOVA ULICA 1A",
  "merchantId": "002295628",
  "transactionDateTime": "2024-12-19T13:15:38.000Z",
  "processDateTime": "2024-12-19T13:15:38.000Z"
}`

// PLACEHOLDER_REF_TX_ID must be replaced with the real refTransactionId.
const txPreauthIncrementalRefJSON = `{
  "operation": 0,
  "transactionClassification": "Authorization",
  "type": 102,
  "isTrxAmountConverted": false,
  "txStatus": "PROCESSING",
  "ghResponseCode": "OK",
  "responseCode": "OK",
  "MastercardConversion": null,
  "transactionAmount": "5.00",
  "transactionCurrency": "EUR",
  "billingAmount": "5.00",
  "billingCurrency": "EUR",
  "vaultId": 1,
  "cardId": 277,
  "refTransactionId": "PLACEHOLDER_REF_TX_ID",
  "ghResponseDescription": "Approved",
  "terminalId": "C0A17138",
  "cardScheme": 3,
  "source": "P",
  "mcc": "5399",
  "merchantName": "SPAR IVANCNA GORICA",
  "merchantCountry": "SVN",
  "merchantCity": "IVANCNA GORIC",
  "merchantZip": "1295",
  "merchantStreet": "STANTETOVA ULICA 1A",
  "merchantId": "002295628",
  "transactionDateTime": "2024-12-19T13:15:38.000Z",
  "processDateTime": "2024-12-19T13:15:38.000Z"
}`

const txWrongPINWithFXJSON = `{
  "operation": 2,
  "transactionClassification": "Advice",
  "type": 0,
  "isTrxAmountConverted": true,
  "txStatus": null,
  "ghResponseCode": "OK",
  "responseCode": "WPIN",
  "MastercardConversion": {"convRate": "0.1934", "refConRate": null, "refConfRateDiff": "0.66"},
  "transactionAmount": "50",
  "transactionCurrency": "RON",
  "billingAmount": "9.67",
  "billingCurrency": "EUR",
  "vaultId": 1,
  "cardId": 277,
  "refTransactionId": null,
  "ghResponseDescription": "Wrong PIN",
  "terminalId": "C0A17138",
  "cardScheme": 3,
  "source": "P",
  "mcc": "5399",
  "merchantName": "SPAR BUCURESTI",
  "merchantCountry": "ROU",
  "merchantCity": "BUCURESTI",
  "merchantZip": "010101",
  "merchantStreet": "CALEA VICTORIEI 1",
  "merchantId": "002295629",
  "transactionDateTime": "2024-12-19T13:15:38.000Z",
  "processDateTime": "2024-12-19T13:15:38.000Z"
}`

const txInsufficientBalanceJSON = `{
  "operation": 2,
  "transactionClassification": "Authorization",
  "type": 0,
  "isTrxAmountConverted": false,
  "txStatus": null,
  "ghResponseCode": "CRGUI",
  "responseCode": "OK",
  "MastercardConversion": null,
  "transactionAmount": "100.00",
  "transactionCurrency": "EUR",
  "billingAmount": "100.00",
  "billingCurrency": "EUR",
  "vaultId": 1,
  "cardId": 277,
  "refTransactionId": null,
  "ghResponseDescription": "Insufficient Balance",
  "terminalId": "C0A17138",
  "cardScheme": 3,
  "source": "P",
  "mcc": "5399",
  "merchantName": "SPAR IVANCNA GORICA",
  "merchantCountry": "SVN",
  "merchantCity": "IVANCNA GORIC",
  "merchantZip": "1295",
  "merchantStreet": "STANTETOVA ULICA 1A",
  "merchantId": "002295628",
  "transactionDateTime": "2024-12-19T13:15:38.000Z",
  "processDateTime": "2024-12-19T13:15:38.000Z"
}`

const txWrongPINJSON = `{
  "operation": 2,
  "transactionClassification": "Advice",
  "type": 0,
  "isTrxAmountConverted": false,
  "txStatus": null,
  "ghResponseCode": "OK",
  "responseCode": "WPIN",
  "MastercardConversion": null,
  "transactionAmount": "1.89",
  "transactionCurrency": "EUR",
  "billingAmount": "1.89",
  "billingCurrency": "EUR",
  "vaultId": 1,
  "cardId": 277,
  "refTransactionId": null,
  "ghResponseDescription": "Wrong PIN",
  "terminalId": "C0A17138",
  "cardScheme": 3,
  "source": "P",
  "mcc": "5399",
  "merchantName": "SPAR IVANCNA GORICA",
  "merchantCountry": "SVN",
  "merchantCity": "IVANCNA GORIC",
  "merchantZip": "1295",
  "merchantStreet": "STANTETOVA ULICA 1A",
  "merchantId": "002295628",
  "transactionDateTime": "2024-12-19T13:15:38.000Z",
  "processDateTime": "2024-12-19T13:15:38.000Z"
}`

const txInvalidCVVJSON = `{
  "operation": 2,
  "transactionClassification": "Authorization",
  "type": 0,
  "isTrxAmountConverted": false,
  "txStatus": null,
  "ghResponseCode": "OK",
  "responseCode": "WCVV2",
  "MastercardConversion": null,
  "transactionAmount": "1.89",
  "transactionCurrency": "EUR",
  "billingAmount": "1.89",
  "billingCurrency": "EUR",
  "vaultId": 1,
  "cardId": 277,
  "refTransactionId": null,
  "ghResponseDescription": "Wrong CVV",
  "terminalId": "C0A17138",
  "cardScheme": 3,
  "source": "P",
  "mcc": "5399",
  "merchantName": "SPAR IVANCNA GORICA",
  "merchantCountry": "SVN",
  "merchantCity": "IVANCNA GORIC",
  "merchantZip": "1295",
  "merchantStreet": "STANTETOVA ULICA 1A",
  "merchantId": "002295628",
  "transactionDateTime": "2024-12-19T13:15:38.000Z",
  "processDateTime": "2024-12-19T13:15:38.000Z"
}`

const txCardExpiredJSON = `{
  "operation": 2,
  "transactionClassification": "Authorization",
  "type": 0,
  "isTrxAmountConverted": false,
  "txStatus": null,
  "ghResponseCode": "OK",
  "responseCode": "CAEXP",
  "MastercardConversion": null,
  "transactionAmount": "1.89",
  "transactionCurrency": "EUR",
  "billingAmount": "1.89",
  "billingCurrency": "EUR",
  "vaultId": 1,
  "cardId": 277,
  "refTransactionId": null,
  "ghResponseDescription": "Card Expired",
  "terminalId": "C0A17138",
  "cardScheme": 3,
  "source": "P",
  "mcc": "5399",
  "merchantName": "SPAR IVANCNA GORICA",
  "merchantCountry": "SVN",
  "merchantCity": "IVANCNA GORIC",
  "merchantZip": "1295",
  "merchantStreet": "STANTETOVA ULICA 1A",
  "merchantId": "002295628",
  "transactionDateTime": "2024-12-19T13:15:38.000Z",
  "processDateTime": "2024-12-19T13:15:38.000Z"
}`

const txFrozenCardJSON = `{
  "operation": 2,
  "transactionClassification": "Authorization",
  "type": 0,
  "isTrxAmountConverted": false,
  "txStatus": null,
  "ghResponseCode": "OK",
  "responseCode": "CASUS",
  "MastercardConversion": null,
  "transactionAmount": "1.89",
  "transactionCurrency": "EUR",
  "billingAmount": "1.89",
  "billingCurrency": "EUR",
  "vaultId": 1,
  "cardId": 277,
  "refTransactionId": null,
  "ghResponseDescription": "Frozen Card",
  "terminalId": "C0A17138",
  "cardScheme": 3,
  "source": "P",
  "mcc": "5399",
  "merchantName": "SPAR IVANCNA GORICA",
  "merchantCountry": "SVN",
  "merchantCity": "IVANCNA GORIC",
  "merchantZip": "1295",
  "merchantStreet": "STANTETOVA ULICA 1A",
  "merchantId": "002295628",
  "transactionDateTime": "2024-12-19T13:15:38.000Z",
  "processDateTime": "2024-12-19T13:15:38.000Z"
}`

const txCardVerificationInquiryJSON = `{
  "operation": 2,
  "transactionClassification": "Authorization",
  "type": 6,
  "isTrxAmountConverted": false,
  "txStatus": null,
  "ghResponseCode": "OK",
  "responseCode": "OK",
  "MastercardConversion": null,
  "transactionAmount": "0",
  "transactionCurrency": "EUR",
  "billingAmount": "0",
  "billingCurrency": "EUR",
  "vaultId": 1,
  "cardId": 277,
  "refTransactionId": null,
  "ghResponseDescription": "Card Verification Inquiry",
  "terminalId": "C0A17138",
  "cardScheme": 3,
  "source": "P",
  "mcc": "5399",
  "merchantName": "SPAR IVANCNA GORICA",
  "merchantCountry": "SVN",
  "merchantCity": "IVANCNA GORIC",
  "merchantZip": "1295",
  "merchantStreet": "STANTETOVA ULICA 1A",
  "merchantId": "002295628",
  "transactionDateTime": "2024-12-19T13:15:38.000Z",
  "processDateTime": "2024-12-19T13:15:38.000Z"
}`

// --- Step trigger implementations ---

func (sc *E2EContext) iTriggerACardPurchaseTransactionFor(userName string) error {
	email, gatehubUserID, cardID, err := sc.buildCardTxSetup(userName)
	if err != nil {
		return fmt.Errorf("iTriggerACardPurchaseTransactionFor: %w", err)
	}
	if err = sc.triggerCardTransaction(gatehubUserID, cardID, "Card Purchase", "EUR 1.89 at SPAR IVANCNA GORICA", txPurchaseJSON); err != nil {
		return fmt.Errorf("iTriggerACardPurchaseTransactionFor: %w", err)
	}
	return sc.waitForGatehubCardTx(email, 1, 30*time.Second)
}

func (sc *E2EContext) iTriggerACardPurchaseWithFXTransactionFor(userName string) error {
	email, gatehubUserID, cardID, err := sc.buildCardTxSetup(userName)
	if err != nil {
		return fmt.Errorf("iTriggerACardPurchaseWithFXTransactionFor: %w", err)
	}
	if err = sc.triggerCardTransaction(gatehubUserID, cardID, "Card Purchase with FX", "RON 50.00 at SPAR IVANCNA GORICA", txPurchaseWithFXJSON); err != nil {
		return fmt.Errorf("iTriggerACardPurchaseWithFXTransactionFor: %w", err)
	}
	return sc.waitForGatehubCardTx(email, 1, 30*time.Second)
}

func (sc *E2EContext) iTriggerACardATMWithdrawalTransactionFor(userName string) error {
	email, gatehubUserID, cardID, err := sc.buildCardTxSetup(userName)
	if err != nil {
		return fmt.Errorf("iTriggerACardATMWithdrawalTransactionFor: %w", err)
	}
	if err = sc.triggerCardTransaction(gatehubUserID, cardID, "ATM Withdrawal", "EUR 50.00 at RAIFFEISEN ATM", txATMWithdrawalJSON); err != nil {
		return fmt.Errorf("iTriggerACardATMWithdrawalTransactionFor: %w", err)
	}
	return sc.waitForGatehubCardTx(email, 1, 30*time.Second)
}

func (sc *E2EContext) iTriggerACardCashAdvanceTransactionFor(userName string) error {
	email, gatehubUserID, cardID, err := sc.buildCardTxSetup(userName)
	if err != nil {
		return fmt.Errorf("iTriggerACardCashAdvanceTransactionFor: %w", err)
	}
	if err = sc.triggerCardTransaction(gatehubUserID, cardID, "Cash Advance", "EUR 20.00 Cash Advance", txCashAdvanceJSON); err != nil {
		return fmt.Errorf("iTriggerACardCashAdvanceTransactionFor: %w", err)
	}
	return sc.waitForGatehubCardTx(email, 1, 30*time.Second)
}

func (sc *E2EContext) iTriggerACardPreauthorizationTransactionFor(userName string) error {
	email, gatehubUserID, cardID, err := sc.buildCardTxSetup(userName)
	if err != nil {
		return fmt.Errorf("iTriggerACardPreauthorizationTransactionFor: %w", err)
	}
	if err = sc.triggerCardTransaction(gatehubUserID, cardID, "Preauthorization", "EUR 5.00 at SPAR IVANCNA GORICA", txPreauthorizationJSON); err != nil {
		return fmt.Errorf("iTriggerACardPreauthorizationTransactionFor: %w", err)
	}
	return sc.waitForGatehubCardTx(email, 1, 30*time.Second)
}

func (sc *E2EContext) iTriggerACardTransferFromAccountTransactionFor(userName string) error {
	email, gatehubUserID, cardID, err := sc.buildCardTxSetup(userName)
	if err != nil {
		return fmt.Errorf("iTriggerACardTransferFromAccountTransactionFor: %w", err)
	}
	if err = sc.triggerCardTransaction(gatehubUserID, cardID, "Mastercard Send", "EUR 10.00 Mastercard Send", txTransferFromAccountJSON); err != nil {
		return fmt.Errorf("iTriggerACardTransferFromAccountTransactionFor: %w", err)
	}
	return sc.waitForGatehubCardTx(email, 1, 30*time.Second)
}

func (sc *E2EContext) iTriggerACardTransferToAccountTransactionFor(userName string) error {
	email, gatehubUserID, cardID, err := sc.buildCardTxSetup(userName)
	if err != nil {
		return fmt.Errorf("iTriggerACardTransferToAccountTransactionFor: %w", err)
	}
	if err = sc.triggerCardTransaction(gatehubUserID, cardID, "Mastercard Send In", "EUR 2.00 Mastercard Send", txTransferToAccountJSON); err != nil {
		return fmt.Errorf("iTriggerACardTransferToAccountTransactionFor: %w", err)
	}
	return sc.waitForGatehubCardTx(email, 1, 30*time.Second)
}

func (sc *E2EContext) iTriggerAPurchaseReversalForTheLastCardTransactionOf(userName string) error {
	email, err := sc.getEmailForUser(userName)
	if err != nil {
		return fmt.Errorf("iTriggerAPurchaseReversalForTheLastCardTransactionOf: %w", err)
	}

	// Get the most recent card transaction ID (the original purchase's GateHub UUID)
	refTxID, err := sc.getLatestGatehubCardTxID(email)
	if err != nil {
		return fmt.Errorf("iTriggerAPurchaseReversalForTheLastCardTransactionOf: get ref tx id: %w", err)
	}

	gatehubUserID, err := sc.getGatehubUserIDByEmail(email)
	if err != nil {
		return fmt.Errorf("iTriggerAPurchaseReversalForTheLastCardTransactionOf: %w", err)
	}

	cardID, err := sc.getActiveCardID(gatehubUserID)
	if err != nil {
		return fmt.Errorf("iTriggerAPurchaseReversalForTheLastCardTransactionOf: %w", err)
	}

	reversalJSON := strings.ReplaceAll(txPurchaseReversalJSON, "PLACEHOLDER_REF_TX_ID", refTxID)
	if err = sc.triggerCardTransaction(gatehubUserID, cardID, "Purchase Reversal", "EUR 1.89 refund", reversalJSON); err != nil {
		return fmt.Errorf("iTriggerAPurchaseReversalForTheLastCardTransactionOf: %w", err)
	}

	// Wait for the reversal card tx to appear (now 2 total)
	return sc.waitForGatehubCardTx(email, 2, 30*time.Second)
}

func (sc *E2EContext) iTriggerAPreauthorizationIncrementalWithReferenceToTheLastCardTransactionOf(userName string) error {
	email, err := sc.getEmailForUser(userName)
	if err != nil {
		return fmt.Errorf("iTriggerAPreauthorizationIncrementalWithReferenceToTheLastCardTransactionOf: %w", err)
	}

	// Get the most recent card transaction ID (the original preauth's GateHub UUID)
	refTxID, err := sc.getLatestGatehubCardTxID(email)
	if err != nil {
		return fmt.Errorf("iTriggerAPreauthorizationIncrementalWithReferenceToTheLastCardTransactionOf: get ref tx id: %w", err)
	}

	gatehubUserID, err := sc.getGatehubUserIDByEmail(email)
	if err != nil {
		return fmt.Errorf("iTriggerAPreauthorizationIncrementalWithReferenceToTheLastCardTransactionOf: %w", err)
	}

	cardID, err := sc.getActiveCardID(gatehubUserID)
	if err != nil {
		return fmt.Errorf("iTriggerAPreauthorizationIncrementalWithReferenceToTheLastCardTransactionOf: %w", err)
	}

	incrementalJSON := strings.ReplaceAll(txPreauthIncrementalRefJSON, "PLACEHOLDER_REF_TX_ID", refTxID)
	if err = sc.triggerCardTransaction(gatehubUserID, cardID, "Preauthorization Incremental", "EUR 5.00 Preauthorization", incrementalJSON); err != nil {
		return fmt.Errorf("iTriggerAPreauthorizationIncrementalWithReferenceToTheLastCardTransactionOf: %w", err)
	}

	// Wait for both card transactions to appear
	return sc.waitForGatehubCardTx(email, 2, 30*time.Second)
}

func (sc *E2EContext) iTriggerAPreauthorizationIncrementalWithoutReferenceFor(userName string) error {
	email, gatehubUserID, cardID, err := sc.buildCardTxSetup(userName)
	if err != nil {
		return fmt.Errorf("iTriggerAPreauthorizationIncrementalWithoutReferenceFor: %w", err)
	}
	if err = sc.triggerCardTransaction(gatehubUserID, cardID, "Preauthorization Incremental", "EUR 5.00 Preauthorization", txPreauthIncrementalNoRefJSON); err != nil {
		return fmt.Errorf("iTriggerAPreauthorizationIncrementalWithoutReferenceFor: %w", err)
	}
	return sc.waitForGatehubCardTx(email, 1, 30*time.Second)
}

func (sc *E2EContext) iTriggerACardWrongPinWithFXTransactionFor(userName string) error {
	email, gatehubUserID, cardID, err := sc.buildCardTxSetup(userName)
	if err != nil {
		return fmt.Errorf("iTriggerACardWrongPinWithFXTransactionFor: %w", err)
	}
	if err = sc.triggerCardTransaction(gatehubUserID, cardID, "Card Declined", "Wrong PIN with FX", txWrongPINWithFXJSON); err != nil {
		return fmt.Errorf("iTriggerACardWrongPinWithFXTransactionFor: %w", err)
	}
	return sc.waitForGatehubCardTx(email, 1, 30*time.Second)
}

func (sc *E2EContext) iTriggerACardInsufficientBalanceTransactionFor(userName string) error {
	email, gatehubUserID, cardID, err := sc.buildCardTxSetup(userName)
	if err != nil {
		return fmt.Errorf("iTriggerACardInsufficientBalanceTransactionFor: %w", err)
	}
	if err = sc.triggerCardTransaction(gatehubUserID, cardID, "Card Declined", "Insufficient Balance", txInsufficientBalanceJSON); err != nil {
		return fmt.Errorf("iTriggerACardInsufficientBalanceTransactionFor: %w", err)
	}
	return sc.waitForGatehubCardTx(email, 1, 30*time.Second)
}

func (sc *E2EContext) iTriggerACardWrongPinTransactionFor(userName string) error {
	email, gatehubUserID, cardID, err := sc.buildCardTxSetup(userName)
	if err != nil {
		return fmt.Errorf("iTriggerACardWrongPinTransactionFor: %w", err)
	}
	if err = sc.triggerCardTransaction(gatehubUserID, cardID, "Card Declined", "Wrong PIN", txWrongPINJSON); err != nil {
		return fmt.Errorf("iTriggerACardWrongPinTransactionFor: %w", err)
	}
	return sc.waitForGatehubCardTx(email, 1, 30*time.Second)
}

func (sc *E2EContext) iTriggerACardInvalidCVVTransactionFor(userName string) error {
	email, gatehubUserID, cardID, err := sc.buildCardTxSetup(userName)
	if err != nil {
		return fmt.Errorf("iTriggerACardInvalidCVVTransactionFor: %w", err)
	}
	if err = sc.triggerCardTransaction(gatehubUserID, cardID, "Card Declined", "Invalid CVV", txInvalidCVVJSON); err != nil {
		return fmt.Errorf("iTriggerACardInvalidCVVTransactionFor: %w", err)
	}
	return sc.waitForGatehubCardTx(email, 1, 30*time.Second)
}

func (sc *E2EContext) iTriggerACardExpiredTransactionFor(userName string) error {
	email, gatehubUserID, cardID, err := sc.buildCardTxSetup(userName)
	if err != nil {
		return fmt.Errorf("iTriggerACardExpiredTransactionFor: %w", err)
	}
	if err = sc.triggerCardTransaction(gatehubUserID, cardID, "Card Declined", "Card Expired", txCardExpiredJSON); err != nil {
		return fmt.Errorf("iTriggerACardExpiredTransactionFor: %w", err)
	}
	return sc.waitForGatehubCardTx(email, 1, 30*time.Second)
}

func (sc *E2EContext) iTriggerAFrozenCardTransactionFor(userName string) error {
	email, gatehubUserID, cardID, err := sc.buildCardTxSetup(userName)
	if err != nil {
		return fmt.Errorf("iTriggerAFrozenCardTransactionFor: %w", err)
	}
	if err = sc.triggerCardTransaction(gatehubUserID, cardID, "Card Declined", "Frozen Card", txFrozenCardJSON); err != nil {
		return fmt.Errorf("iTriggerAFrozenCardTransactionFor: %w", err)
	}
	return sc.waitForGatehubCardTx(email, 1, 30*time.Second)
}

func (sc *E2EContext) iTriggerACardVerificationInquiryTransactionFor(userName string) error {
	email, gatehubUserID, cardID, err := sc.buildCardTxSetup(userName)
	if err != nil {
		return fmt.Errorf("iTriggerACardVerificationInquiryTransactionFor: %w", err)
	}
	if err = sc.triggerCardTransaction(gatehubUserID, cardID, "Card Verification", "Card Verification Inquiry", txCardVerificationInquiryJSON); err != nil {
		return fmt.Errorf("iTriggerACardVerificationInquiryTransactionFor: %w", err)
	}
	return sc.waitForGatehubCardTx(email, 1, 30*time.Second)
}

// --- Polling/status step implementations ---

func (sc *E2EContext) iSetTheCardTransactionStatusToForTheLastCardTransactionOf(status, userName string) error {
	email, err := sc.getEmailForUser(userName)
	if err != nil {
		return fmt.Errorf("iSetTheCardTransactionStatusToForTheLastCardTransactionOf: %w", err)
	}

	gatehubUserID, err := sc.getGatehubUserIDByEmail(email)
	if err != nil {
		return fmt.Errorf("iSetTheCardTransactionStatusToForTheLastCardTransactionOf: %w", err)
	}

	txID, err := sc.getLatestGatehubCardTxID(email)
	if err != nil {
		return fmt.Errorf("iSetTheCardTransactionStatusToForTheLastCardTransactionOf: %w", err)
	}

	return sc.setCardTransactionStatus(txID, status, gatehubUserID)
}

func (sc *E2EContext) iRunTheGatehubClearingCardTransactionsPollWorkflow() error {
	return sc.runTemporalWorkflowByName("GatehubClearingCardTransactionsPollWorkflow")
}

func (sc *E2EContext) iRunTheGatehubRealtimeCardTransactionsPollWorkflow() error {
	return sc.runTemporalWorkflowByName("GatehubRealtimeCardTransactionsPollWorkflow")
}

// --- Assertion step implementations ---

func (sc *E2EContext) aCardTransactionRecordShouldExistInTheDatabaseFor(userName string) error {
	email, err := sc.getEmailForUser(userName)
	if err != nil {
		return err
	}
	return sc.waitForGatehubCardTx(email, 1, 30*time.Second)
}

func (sc *E2EContext) aPendingWithdrawalTransactionShouldExistWithNoteFor(userName, note string) error {
	email, err := sc.getEmailForUser(userName)
	if err != nil {
		return err
	}
	return sc.assertLatestCardTxHasStateAndNote(email, "Pending", note, 30*time.Second)
}

func (sc *E2EContext) aPendingDepositTransactionShouldExistWithNoteFor(userName, note string) error {
	email, err := sc.getEmailForUser(userName)
	if err != nil {
		return err
	}
	return sc.assertLatestCardTxHasStateAndNote(email, "Pending", note, 30*time.Second)
}

func (sc *E2EContext) aFailedTransactionShouldExistWithNoteFor(userName, note string) error {
	email, err := sc.getEmailForUser(userName)
	if err != nil {
		return err
	}
	return sc.assertLatestCardTxHasStateAndNote(email, "Failed", note, 30*time.Second)
}

func (sc *E2EContext) aCompletedTransactionShouldExistWithNoteFor(userName, note string) error {
	email, err := sc.getEmailForUser(userName)
	if err != nil {
		return err
	}
	return sc.assertLatestCardTxHasStateAndNote(email, "Completed", note, 30*time.Second)
}

// assertLatestCardTxHasStateAndNote polls until the latest card transaction has the expected state and note.
func (sc *E2EContext) assertLatestCardTxHasStateAndNote(email, expectedState, expectedNote string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastTx *backendCardTx

	for time.Now().Before(deadline) {
		tx, err := sc.getLatestBackendCardTx(email)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		lastTx = tx

		actualNote := ""
		if tx.Note.Valid {
			actualNote = tx.Note.String
		}

		if tx.State == expectedState && actualNote == expectedNote {
			debugPrintf("   ✓ Card transaction state=%s note=%q for user %s\n", tx.State, actualNote, email)
			return nil
		}

		debugPrintf("   ⏳ Waiting for state=%s note=%q, got state=%s note=%q\n", expectedState, expectedNote, tx.State, actualNote)
		time.Sleep(time.Second)
	}

	if lastTx != nil {
		actualNote := ""
		if lastTx.Note.Valid {
			actualNote = lastTx.Note.String
		}
		return fmt.Errorf("assertLatestCardTxHasStateAndNote: expected state=%s note=%q but got state=%s note=%q for user %s",
			expectedState, expectedNote, lastTx.State, actualNote, email)
	}
	return fmt.Errorf("assertLatestCardTxHasStateAndNote: no card transaction found for user %s", email)
}

func (sc *E2EContext) nCardTransactionsShouldExistInTheDatabaseFor(n int, userName string) error {
	email, err := sc.getEmailForUser(userName)
	if err != nil {
		return err
	}
	if err = sc.waitForBackendCardTx(email, n, 30*time.Second); err != nil {
		return err
	}

	if err = sc.ensureDB(); err != nil {
		return err
	}
	walletID, err := sc.getWalletIDByEmail(email)
	if err != nil {
		return err
	}

	var total int
	err = sc.db.QueryRow(`
		SELECT COUNT(*) FROM transactions WHERE wallet_id = $1 AND type = 'card_transaction'
	`, walletID).Scan(&total)
	if err != nil {
		return fmt.Errorf("nCardTransactionsShouldExistInTheDatabaseFor: query error: %w", err)
	}
	if total != n {
		return fmt.Errorf("nCardTransactionsShouldExistInTheDatabaseFor: expected %d card transactions but found %d for user %s", n, total, email)
	}
	return nil
}

func (sc *E2EContext) theFirstCardTransactionShouldBeFailedFor(userName string) error {
	email, err := sc.getEmailForUser(userName)
	if err != nil {
		return err
	}
	if err = sc.ensureDB(); err != nil {
		return err
	}
	walletID, err := sc.getWalletIDByEmail(email)
	if err != nil {
		return err
	}

	deadline := time.Now().Add(30 * time.Second)
	lastState := ""

	for time.Now().Before(deadline) {
		var state string
		if queryErr := sc.db.QueryRow(`
			SELECT state FROM transactions
			WHERE wallet_id = $1 AND type = 'card_transaction'
			ORDER BY created_at ASC LIMIT 1
		`, walletID).Scan(&state); queryErr != nil {
			debugPrintf("   ⚠️  theFirstCardTransactionShouldBeFailedFor: query error: %v\n", queryErr)
			time.Sleep(time.Second)
			continue
		}
		lastState = state
		if state == "Failed" {
			debugPrintf("   ✓ First card transaction is failed for user %s\n", email)
			return nil
		}
		time.Sleep(time.Second)
	}

	return fmt.Errorf("theFirstCardTransactionShouldBeFailedFor: expected first card transaction to be failed but got '%s' for user %s", lastState, email)
}

func (sc *E2EContext) theLastCardTransactionShouldBeCompletedFor(userName string) error {
	email, err := sc.getEmailForUser(userName)
	if err != nil {
		return err
	}
	return sc.waitForBackendCardTxState(email, "Completed", 30*time.Second)
}

func (sc *E2EContext) theLastCardTransactionShouldBeFailedFor(userName string) error {
	email, err := sc.getEmailForUser(userName)
	if err != nil {
		return err
	}
	return sc.waitForBackendCardTxState(email, "Failed", 30*time.Second)
}

func (sc *E2EContext) theLastTransactionShouldHaveExchangeRateDataFor(userName string) error {
	email, err := sc.getEmailForUser(userName)
	if err != nil {
		return err
	}
	if err = sc.waitForBackendCardTx(email, 1, 30*time.Second); err != nil {
		return err
	}

	deadline := time.Now().Add(30 * time.Second)

	for time.Now().Before(deadline) {
		tx, err := sc.getLatestBackendCardTx(email)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		if tx.ExchangeRateApplied.Valid && tx.ExchangeRateApplied.String != "" {
			debugPrintf("   ✓ Exchange rate data present: %s for user %s\n", tx.ExchangeRateApplied.String, email)
			return nil
		}
		time.Sleep(time.Second)
	}

	return fmt.Errorf("theLastTransactionShouldHaveExchangeRateDataFor: no exchange rate data found for user %s", email)
}

