package handler

import (
	"encoding/json"
	"math"
	"net/http"
	"sync"

	"gitlab.com/fynbos/mock/mockgatehub/internal/logger"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// FeeConfig holds configurable fee percentages (thread-safe).
type FeeConfig struct {
	mu                   sync.RWMutex
	depositFeePercent    float64
	withdrawalFeePercent float64
	userOverrides        map[string]userFeeOverride
}

type userFeeOverride struct {
	depositPercent    *float64
	withdrawalPercent *float64
}

// NewFeeConfig creates a FeeConfig with 0% defaults.
func NewFeeConfig() *FeeConfig {
	return &FeeConfig{
		userOverrides: make(map[string]userFeeOverride),
	}
}

// GetDepositFeePercent returns the current deposit fee percentage.
func (fc *FeeConfig) GetDepositFeePercent() float64 {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.depositFeePercent
}

// GetWithdrawalFeePercent returns the current withdrawal fee percentage.
func (fc *FeeConfig) GetWithdrawalFeePercent() float64 {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.withdrawalFeePercent
}

// GetDepositFeeForUser returns the deposit fee percent and whether a user override exists.
func (fc *FeeConfig) GetDepositFeeForUser(userID string) (float64, bool) {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	if override, ok := fc.userOverrides[userID]; ok && override.depositPercent != nil {
		return *override.depositPercent, true
	}
	return fc.depositFeePercent, false
}

// GetWithdrawalFeeForUser returns the withdrawal fee percent and whether a user override exists.
func (fc *FeeConfig) GetWithdrawalFeeForUser(userID string) (float64, bool) {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	if override, ok := fc.userOverrides[userID]; ok && override.withdrawalPercent != nil {
		return *override.withdrawalPercent, true
	}
	return fc.withdrawalFeePercent, false
}

// SetDepositFeePercent sets the deposit fee percentage.
func (fc *FeeConfig) SetDepositFeePercent(pct float64) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.depositFeePercent = pct
}

// SetWithdrawalFeePercent sets the withdrawal fee percentage.
func (fc *FeeConfig) SetWithdrawalFeePercent(pct float64) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.withdrawalFeePercent = pct
}

// SetUserFees sets per-user fee overrides. Nil values leave existing overrides unchanged.
func (fc *FeeConfig) SetUserFees(userID string, depositPct, withdrawalPct *float64) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	override := fc.userOverrides[userID]
	if depositPct != nil {
		override.depositPercent = depositPct
	}
	if withdrawalPct != nil {
		override.withdrawalPercent = withdrawalPct
	}
	fc.userOverrides[userID] = override
}

// ClearUserFees removes any per-user fee overrides.
func (fc *FeeConfig) ClearUserFees(userID string) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	delete(fc.userOverrides, userID)
}

// CalculateFee computes fee = amount * (percent/100), rounded to 2 decimals.
func CalculateFee(amount, percent float64) float64 {
	raw := amount * percent / 100.0
	return math.Round(raw*100) / 100
}

// ── Admin endpoints ──────────────────────────────────────────────────

// feeResponse is the JSON shape for GET/PUT /admin/fees.
type feeResponse struct {
	DepositFeePercentage    float64 `json:"deposit_fee_percentage"`
	WithdrawalFeePercentage float64 `json:"withdrawal_fee_percentage"`
}

type feeUpdateRequest struct {
	DepositFeePercentage    *float64 `json:"deposit_fee_percentage"`
	WithdrawalFeePercentage *float64 `json:"withdrawal_fee_percentage"`
}

type userFeeResponse struct {
	UserID                  string  `json:"user_id"`
	DepositFeePercentage    float64 `json:"deposit_fee_percentage"`
	WithdrawalFeePercentage float64 `json:"withdrawal_fee_percentage"`
	DepositFeeSource        string  `json:"deposit_fee_source"`
	WithdrawalFeeSource     string  `json:"withdrawal_fee_source"`
}

// GetFees returns the current fee configuration.
func (h *Handler) GetFees(w http.ResponseWriter, r *http.Request) {
	resp := feeResponse{
		DepositFeePercentage:    h.feeConfig.GetDepositFeePercent(),
		WithdrawalFeePercentage: h.feeConfig.GetWithdrawalFeePercent(),
	}
	h.sendJSON(w, http.StatusOK, resp)
}

// SetFees updates the fee configuration.
func (h *Handler) SetFees(w http.ResponseWriter, r *http.Request) {
	var req feeResponse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate
	if req.DepositFeePercentage < 0 || req.DepositFeePercentage > 100 {
		h.sendError(w, http.StatusBadRequest, "deposit_fee_percentage must be between 0 and 100")
		return
	}
	if req.WithdrawalFeePercentage < 0 || req.WithdrawalFeePercentage > 100 {
		h.sendError(w, http.StatusBadRequest, "withdrawal_fee_percentage must be between 0 and 100")
		return
	}

	h.feeConfig.SetDepositFeePercent(req.DepositFeePercentage)
	h.feeConfig.SetWithdrawalFeePercent(req.WithdrawalFeePercentage)

	logger.Info("fee configuration updated",
		zap.Float64("deposit_fee_pct", req.DepositFeePercentage),
		zap.Float64("withdrawal_fee_pct", req.WithdrawalFeePercentage),
	)

	resp := feeResponse{
		DepositFeePercentage:    h.feeConfig.GetDepositFeePercent(),
		WithdrawalFeePercentage: h.feeConfig.GetWithdrawalFeePercent(),
	}
	h.sendJSON(w, http.StatusOK, resp)
}

// GetUserFees returns the effective fee configuration for a specific user.
func (h *Handler) GetUserFees(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		h.sendError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	depositPct, depositOverride := h.feeConfig.GetDepositFeeForUser(userID)
	withdrawalPct, withdrawalOverride := h.feeConfig.GetWithdrawalFeeForUser(userID)

	resp := userFeeResponse{
		UserID:                  userID,
		DepositFeePercentage:    depositPct,
		WithdrawalFeePercentage: withdrawalPct,
		DepositFeeSource:        feeSource(depositOverride),
		WithdrawalFeeSource:     feeSource(withdrawalOverride),
	}
	h.sendJSON(w, http.StatusOK, resp)
}

// SetUserFees updates fee overrides for a specific user.
func (h *Handler) SetUserFees(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		h.sendError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	var req feeUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.DepositFeePercentage == nil && req.WithdrawalFeePercentage == nil {
		h.sendError(w, http.StatusBadRequest, "at least one fee percentage must be provided")
		return
	}

	if req.DepositFeePercentage != nil {
		if *req.DepositFeePercentage < 0 || *req.DepositFeePercentage > 100 {
			h.sendError(w, http.StatusBadRequest, "deposit_fee_percentage must be between 0 and 100")
			return
		}
	}
	if req.WithdrawalFeePercentage != nil {
		if *req.WithdrawalFeePercentage < 0 || *req.WithdrawalFeePercentage > 100 {
			h.sendError(w, http.StatusBadRequest, "withdrawal_fee_percentage must be between 0 and 100")
			return
		}
	}

	h.feeConfig.SetUserFees(userID, req.DepositFeePercentage, req.WithdrawalFeePercentage)

	logger.Info("user fee configuration updated",
		zap.String("user_id", userID),
		zap.Any("deposit_fee_pct", req.DepositFeePercentage),
		zap.Any("withdrawal_fee_pct", req.WithdrawalFeePercentage),
	)

	h.GetUserFees(w, r)
}

// ClearUserFees removes any fee overrides for a specific user.
func (h *Handler) ClearUserFees(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		h.sendError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	h.feeConfig.ClearUserFees(userID)

	logger.Info("user fee configuration cleared", zap.String("user_id", userID))

	h.GetUserFees(w, r)
}

func feeSource(userOverride bool) string {
	if userOverride {
		return "user"
	}
	return "global"
}
