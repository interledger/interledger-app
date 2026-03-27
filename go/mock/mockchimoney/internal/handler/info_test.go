package handler

import (
	"net/http"
	"testing"
)

func TestFeeEstimateScenarios(t *testing.T) {
	router, _, cleanup := setupFullRouter(t, "http://localhost:1/webhooks")
	defer cleanup()

	okResp := doJSONRequest(t, router, http.MethodPost, "/v0.2.4/info/fee-estimate", `{"amount":100.0,"currency":"CAD","rail":"interac","direction":"payout"}`)
	if okResp.Code != http.StatusOK {
		t.Fatalf("fee estimate status mismatch: got %d", okResp.Code)
	}

	missingAmount := doJSONRequest(t, router, http.MethodPost, "/v0.2.4/info/fee-estimate", `{"currency":"CAD","rail":"interac"}`)
	if missingAmount.Code != http.StatusBadRequest {
		t.Fatalf("missing amount expected 400 got %d", missingAmount.Code)
	}

	noRailUSD := doJSONRequest(t, router, http.MethodPost, "/v0.2.4/info/fee-estimate", `{"amount":50.0,"currency":"USD"}`)
	if noRailUSD.Code != http.StatusOK {
		t.Fatalf("no rail USD expected 200 got %d", noRailUSD.Code)
	}

	noRailCAD := doJSONRequest(t, router, http.MethodPost, "/v0.2.4/info/fee-estimate", `{"amount":50.0,"currency":"CAD"}`)
	if noRailCAD.Code != http.StatusBadRequest {
		t.Fatalf("no rail CAD expected 400 got %d", noRailCAD.Code)
	}
}

func TestConversionScenarios(t *testing.T) {
	router, _, cleanup := setupFullRouter(t, "http://localhost:1/webhooks")
	defer cleanup()

	okResp := doJSONRequest(t, router, http.MethodGet, "/v0.2.4/info/convert/local-amount-to-usd?originCurrency=CAD&amountInOriginCurrency=100", "")
	if okResp.Code != http.StatusOK {
		t.Fatalf("conversion status mismatch: got %d", okResp.Code)
	}

	missingOrigin := doJSONRequest(t, router, http.MethodGet, "/v0.2.4/info/convert/local-amount-to-usd?amountInOriginCurrency=100", "")
	if missingOrigin.Code != http.StatusBadRequest {
		t.Fatalf("missing origin expected 400 got %d", missingOrigin.Code)
	}

	missingAmount := doJSONRequest(t, router, http.MethodGet, "/v0.2.4/info/convert/local-amount-to-usd?originCurrency=CAD", "")
	if missingAmount.Code != http.StatusBadRequest {
		t.Fatalf("missing amount expected 400 got %d", missingAmount.Code)
	}

	zeroResp := doJSONRequest(t, router, http.MethodGet, "/v0.2.4/info/convert/local-amount-to-usd?originCurrency=CAD&amountInOriginCurrency=0", "")
	if zeroResp.Code != http.StatusOK {
		t.Fatalf("zero conversion expected 200 got %d", zeroResp.Code)
	}
}
