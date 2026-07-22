package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/interledger/interledger-app/go/mock/mockxago/internal/models"
)

func doConvertCurrency(t *testing.T, h *Handler, req models.ConvertCurrencyRequest) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/xago/v1/currencyconvert", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ConvertCurrencyHandler(w, httpReq)
	return w
}

func TestConvertCurrencyHandler_Estimate_ZARtoEUR(t *testing.T) {
	h := setupTestHandler(t)
	params := pairParams[models.ZARtoEUR]
	amount := 1000.0

	w := doConvertCurrency(t, h, models.ConvertCurrencyRequest{
		ConvertCurrencyPair: models.ZARtoEUR,
		Amount:              amount,
		EstimateCalculation: true,
	})

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.EstimateConvertCurrencyResponse
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	finalBuyAmount := roundTo6Decimals(amount / params.buyPrice)
	receivedAmount := roundTo6Decimals(amount * params.rate)
	assert.Equal(t, params.buyPrice, resp.BuyAveragePrice)
	assert.Equal(t, params.sellPrice, resp.SellOrders)
	assert.Equal(t, params.rate, resp.EstimatedRate)
	assert.Equal(t, finalBuyAmount, resp.FinalBuyAmount)
	assert.Equal(t, finalBuyAmount, resp.FinalSellAmount)
	assert.Equal(t, finalBuyAmount, resp.BuyOrders)
	assert.Equal(t, amount, resp.QuoteAmount)
	assert.Equal(t, receivedAmount, resp.ReceivedAmount)
}

func TestConvertCurrencyHandler_Estimate_EURtoZAR(t *testing.T) {
	h := setupTestHandler(t)
	params := pairParams[models.EURtoZAR]
	amount := 500.0

	w := doConvertCurrency(t, h, models.ConvertCurrencyRequest{
		ConvertCurrencyPair: models.EURtoZAR,
		Amount:              amount,
		EstimateCalculation: true,
	})

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.EstimateConvertCurrencyResponse
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	finalBuyAmount := roundTo6Decimals(amount / params.buyPrice)
	receivedAmount := roundTo6Decimals(amount * params.rate)
	assert.Equal(t, params.buyPrice, resp.BuyAveragePrice)
	assert.Equal(t, params.sellPrice, resp.SellOrders)
	assert.Equal(t, params.rate, resp.EstimatedRate)
	assert.Equal(t, finalBuyAmount, resp.FinalBuyAmount)
	assert.Equal(t, amount, resp.QuoteAmount)
	assert.Equal(t, receivedAmount, resp.ReceivedAmount)
}

func TestConvertCurrencyHandler_Actual_Success(t *testing.T) {
	h := setupTestHandler(t)
	params := pairParams[models.ZARtoEUR]
	amount := 1000.0

	w := doConvertCurrency(t, h, models.ConvertCurrencyRequest{
		ConvertCurrencyPair: models.ZARtoEUR,
		Amount:              amount,
		EstimateCalculation: false,
	})

	assert.Equal(t, http.StatusOK, w.Code)

	var convertID string
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&convertID))
	assert.NotEmpty(t, convertID)

	conv, err := h.store.GetCurrencyConversion(context.Background(), convertID)
	assert.NoError(t, err)
	sendFee := roundTo6Decimals(amount * params.sendFeeRate)
	finalBuyAmount := roundTo6Decimals(amount / params.buyPrice)
	receivedAmount := roundTo6Decimals(amount * params.rate)
	assert.Equal(t, convertID, conv.ConvertID)
	assert.Equal(t, params.sendCurrency, conv.SendCurrencyCode)
	assert.Equal(t, params.receiveCurrency, conv.ReceiveCurrencyCode)
	assert.Equal(t, "SUCCESS", conv.Status)
	assert.Equal(t, "SUCCESS", conv.BuyStatus)
	assert.Equal(t, "COMPLETED", conv.SellStatus)
	assert.Equal(t, "Conversion", conv.Type)
	assert.Equal(t, "XRP", conv.BridgeCurrency)
	assert.Equal(t, amount, conv.SendAmount)
	assert.Equal(t, sendFee, conv.SendFee)
	assert.Equal(t, finalBuyAmount, conv.BridgeAmount)
	assert.Equal(t, params.buyPrice, conv.BuyPrice)
	assert.Equal(t, params.sellPrice, conv.SellPrice)
	assert.Equal(t, params.rate, conv.Rate)
	assert.Equal(t, receivedAmount, conv.ReceiveAmount)
	assert.NotEmpty(t, conv.ID)
	assert.NotEmpty(t, conv.UUID)
	assert.NotEmpty(t, conv.BuyOrderID)
	assert.NotEmpty(t, conv.SellOrderID)
}

func TestConvertCurrencyHandler_InvalidJSON(t *testing.T) {
	h := setupTestHandler(t)

	httpReq := httptest.NewRequest(http.MethodPost, "/xago/v1/currencyconvert", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()
	h.ConvertCurrencyHandler(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "invalid_request", resp.Error)
}

func TestConvertCurrencyHandler_UnsupportedPair(t *testing.T) {
	h := setupTestHandler(t)

	w := doConvertCurrency(t, h, models.ConvertCurrencyRequest{
		ConvertCurrencyPair: models.ConvertCurrencyPairEnum("USD/GBP"),
		Amount:              100,
		EstimateCalculation: true,
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "invalid_request", resp.Error)
}

func TestConvertCurrencyHandler_ZeroAmount(t *testing.T) {
	h := setupTestHandler(t)

	w := doConvertCurrency(t, h, models.ConvertCurrencyRequest{
		ConvertCurrencyPair: models.ZARtoEUR,
		Amount:              0,
		EstimateCalculation: true,
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "invalid_request", resp.Error)
}

func TestConvertCurrencyHandler_NegativeAmount(t *testing.T) {
	h := setupTestHandler(t)

	w := doConvertCurrency(t, h, models.ConvertCurrencyRequest{
		ConvertCurrencyPair: models.ZARtoEUR,
		Amount:              -50,
		EstimateCalculation: true,
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "invalid_request", resp.Error)
}

func TestGetConvertCurrencyDetails_Success(t *testing.T) {
	h := setupTestHandler(t)
	params := pairParams[models.ZARtoEUR]
	amount := 1000.0

	createW := doConvertCurrency(t, h, models.ConvertCurrencyRequest{
		ConvertCurrencyPair: models.ZARtoEUR,
		Amount:              amount,
		EstimateCalculation: false,
	})
	var convertID string
	json.NewDecoder(createW.Body).Decode(&convertID)

	getReq := httptest.NewRequest(http.MethodGet, "/xago/v1/currencyconvert?convertId="+convertID, nil)
	w := httptest.NewRecorder()
	h.GetConvertCurrencyDetails(w, getReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var conv models.CurrencyConversion
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&conv))
	receivedAmount := roundTo6Decimals(amount * params.rate)
	assert.Equal(t, convertID, conv.ConvertID)
	assert.Equal(t, params.sendCurrency, conv.SendCurrencyCode)
	assert.Equal(t, params.receiveCurrency, conv.ReceiveCurrencyCode)
	assert.Equal(t, amount, conv.SendAmount)
	assert.Equal(t, receivedAmount, conv.ReceiveAmount)
}

func TestGetConvertCurrencyDetails_MissingConvertID(t *testing.T) {
	h := setupTestHandler(t)

	getReq := httptest.NewRequest(http.MethodGet, "/xago/v1/currencyconvert", nil)
	w := httptest.NewRecorder()
	h.GetConvertCurrencyDetails(w, getReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "invalid_request", resp.Error)
}

func TestGetConvertCurrencyDetails_NotFound(t *testing.T) {
	h := setupTestHandler(t)

	getReq := httptest.NewRequest(http.MethodGet, "/xago/v1/currencyconvert?convertId=nonexistent", nil)
	w := httptest.NewRecorder()
	h.GetConvertCurrencyDetails(w, getReq)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "not_found", resp.Error)
}
