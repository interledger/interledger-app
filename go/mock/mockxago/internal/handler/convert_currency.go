package handler

import (
	"encoding/json"
	"math"
	"net/http"
	"time"

	"gitlab.com/fynbos/mock/mockxago/internal/logger"
	"gitlab.com/fynbos/mock/mockxago/internal/models"
	"gitlab.com/fynbos/mock/mockxago/internal/storage"
	"gitlab.com/fynbos/mock/mockxago/internal/utils"
)

// fxParams holds mock FX parameters for a currency pair.
type fxParams struct {
	rate            float64 // send-to-receive rate
	buyPrice        float64 // send-currency per XRP
	sellPrice       float64 // receive-currency per XRP
	sendFeeRate     float64 // fraction of send amount charged as fee
	sendCurrency    string
	receiveCurrency string
}

var pairParams = map[models.ConvertCurrencyPairEnum]fxParams{
	models.ZARtoEUR: {
		rate:            0.051218,
		buyPrice:        23.907695,
		sellPrice:       1.230025,
		sendFeeRate:     0.00224495,
		sendCurrency:    "ZAR",
		receiveCurrency: "EUR",
	},
	models.EURtoZAR: {
		rate:            19.524,
		buyPrice:        0.8130,
		sellPrice:       15.874,
		sendFeeRate:     0.00225,
		sendCurrency:    "EUR",
		receiveCurrency: "ZAR",
	},
}

func round6(v float64) float64 {
	return math.Round(v*1e6) / 1e6
}

// ConvertCurrencyHandler handles POST /v1/exchange/currencyconvert.
// When estimateCalculation is true it returns an EstimateConvertCurrencyResponse.
// When estimateCalculation is false it records the conversion and returns its UUID.
func (h *Handler) ConvertCurrencyHandler(w http.ResponseWriter, r *http.Request) {
	var req models.ConvertCurrencyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	params, ok := pairParams[req.ConvertCurrencyPair]
	if !ok {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "unsupported currency pair")
		return
	}

	// TODO <= 0
	if req.Amount < 0 {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "amount must be positive")
		return
	}

	sendFee := round6(req.Amount * params.sendFeeRate)
	finalBuyAmount := round6(req.Amount / params.buyPrice)
	receivedAmount := round6(req.Amount * params.rate)

	if req.EstimateCalculation {
		resp := models.EstimateConvertCurrencyResponse{
			BuyAveragePrice: params.buyPrice,
			BuyOrders:       finalBuyAmount,
			EstimatedRate:   params.rate,
			FinalBuyAmount:  finalBuyAmount,
			FinalSellAmount: finalBuyAmount,
			QuoteAmount:     req.Amount,
			ReceivedAmount:  receivedAmount,
			SellOrders:      params.sellPrice,
		}
		h.sendJSON(w, http.StatusOK, resp)
		return
	}

	// Actual conversion — persist and return the convertId.
	now := time.Now().UTC()
	conv := &models.CurrencyConversion{
		ID:                  utils.GenerateUUID(),
		UUID:                utils.GenerateUUID(),
		Timestamp:           now.UnixMilli(),
		SendCurrencyCode:    params.sendCurrency,
		ReceiveCurrencyCode: params.receiveCurrency,
		ConvertID:           utils.GenerateUUID(),
		BuyOrderID:          utils.GenerateUUID(),
		Status:              "SUCCESS",
		SendAmount:          req.Amount,
		Type:                "Conversion",
		BuyStatus:           "SUCCESS",
		CreatedAt:           now,
		UpdatedAt:           now,
		BridgeAmount:        finalBuyAmount,
		BridgeCurrency:      "XRP",
		BuyPrice:            params.buyPrice,
		SendFee:             sendFee,
		SellOrderID:         utils.GenerateUUID(),
		SellStatus:          "COMPLETED",
		Rate:                params.rate,
		ReceiveAmount:       receivedAmount,
		SellPrice:           params.sellPrice,
	}

	if err := h.store.SaveCurrencyConversion(r.Context(), conv); err != nil {
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to save conversion")
		logger.Errorf("failed to save currency conversion: %v", err)
		return
	}

	logger.Infof("currency conversion created: convertId=%s pair=%s amount=%v", conv.ConvertID, req.ConvertCurrencyPair, req.Amount)

	h.sendJSON(w, http.StatusOK, conv.ConvertID)
}

// GetConvertCurrencyDetails handles GET /v1/exchange/currencyconvert?convertId=...
func (h *Handler) GetConvertCurrencyDetails(w http.ResponseWriter, r *http.Request) {
	convertID := r.URL.Query().Get("convertId")
	if convertID == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "convertId query parameter is required")
		return
	}

	conv, err := h.store.GetCurrencyConversion(r.Context(), convertID)
	if err != nil {
		if err == storage.ErrCurrencyConversionNotFound {
			h.sendError(w, http.StatusNotFound, "not_found", "currency conversion not found")
			return
		}
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to retrieve conversion")
		logger.Errorf("failed to get currency conversion %s: %v", convertID, err)
		return
	}

	h.sendJSON(w, http.StatusOK, conv)
}
