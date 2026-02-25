package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListCurrencies_NoAuthRequired(t *testing.T) {
	h := setupTestHandler(t)

	r := httptest.NewRequest(http.MethodGet, "/xago/v1/currencies", nil)
	w := httptest.NewRecorder()

	h.ListCurrencies(w, r)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	assert.GreaterOrEqual(t, len(resp), 2)
	
	// Verify nested format
	assert.Equal(t, "ZAR", resp[0]["currencyCode"])
	assert.Equal(t, "USD", resp[1]["currencyCode"])
	
	// Verify nested structure
	assert.NotNil(t, resp[0]["bankingProviders"])
	assert.NotNil(t, resp[1]["bankingProviders"])
	
	// Verify ZAR has banking providers
	zarProviders, ok := resp[0]["bankingProviders"].([]interface{})
	assert.True(t, ok)
	assert.GreaterOrEqual(t, len(zarProviders), 1)
}

func TestListCurrencies_IsStableAcrossCalls(t *testing.T) {
	h := setupTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/xago/v1/currencies", nil)

	w1 := httptest.NewRecorder()
	h.ListCurrencies(w1, req)
	var first []map[string]string
	json.NewDecoder(w1.Body).Decode(&first)

	w2 := httptest.NewRecorder()
	h.ListCurrencies(w2, req)
	var second []map[string]string
	json.NewDecoder(w2.Body).Decode(&second)

	assert.Equal(t, first, second)
}
