package handler

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"gitlab.com/fynbos/mock/mockgatehub/internal/consts"
	"gitlab.com/fynbos/mock/mockgatehub/internal/logger"
	"gitlab.com/fynbos/mock/mockgatehub/internal/models"
	"gitlab.com/fynbos/mock/mockgatehub/internal/utils"
	"go.uber.org/zap"
)

//go:embed web/ui/dashboard.html web/ui/user.html web/ui/kyc_action.html web/ui/card_tx_action.html
var uiTemplateFS embed.FS

func (h *Handler) UIDashboard(w http.ResponseWriter, r *http.Request) {
	users, err := h.store.ListUsers()
	if err != nil {
		logger.Error("ui: failed to list users", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFS(uiTemplateFS, "web/ui/dashboard.html")
	if err != nil {
		logger.Error("ui: failed to parse dashboard template", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, map[string]interface{}{"Users": users}); err != nil {
		logger.Error("ui: failed to render dashboard", zap.Error(err))
	}
}

func (h *Handler) UIUserDetail(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	user, err := h.store.GetUser(userID)
	if err != nil || user == nil {
		http.NotFound(w, r)
		return
	}

	balances, err := h.store.GetAllBalances(userID)
	if err != nil {
		logger.Error("ui: failed to get balances", zap.String("user_id", userID), zap.Error(err))
		balances = map[string]float64{}
	}

	txns, err := h.store.ListTransactionsByUser(userID)
	if err != nil {
		logger.Error("ui: failed to list transactions", zap.String("user_id", userID), zap.Error(err))
		txns = nil
	}

	tmpl, err := template.ParseFS(uiTemplateFS, "web/ui/user.html")
	if err != nil {
		logger.Error("ui: failed to parse user template", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, map[string]interface{}{
		"User":         user,
		"Balances":     balances,
		"Transactions": txns,
	}); err != nil {
		logger.Error("ui: failed to render user detail", zap.Error(err))
	}
}

func (h *Handler) UIKYCForm(w http.ResponseWriter, r *http.Request) {
	users, err := h.store.ListUsers()
	if err != nil {
		logger.Error("ui: failed to list users for kyc form", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	q := r.URL.Query()
	data := map[string]interface{}{
		"Users":          users,
		"SelectedUserID": q.Get("userID"),
		"Gateway":        q.Get("gateway"),
		"Status":         q.Get("status"),
		"Flash":          q.Get("flash"),
		"FlashOK":        q.Get("ok") == "1",
	}
	if data["Gateway"] == "" {
		data["Gateway"] = "paywiser-eu-sandbox"
	}
	if data["Status"] == "" {
		data["Status"] = "accepted"
	}

	tmpl, err := template.ParseFS(uiTemplateFS, "web/ui/kyc_action.html")
	if err != nil {
		logger.Error("ui: failed to parse kyc template", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		logger.Error("ui: failed to render kyc form", zap.Error(err))
	}
}

func (h *Handler) UIKYCAction(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/ui/actions/kyc?flash=Invalid+form+data", http.StatusSeeOther)
		return
	}

	userID := r.FormValue("userID")
	gateway := r.FormValue("gateway")
	status := r.FormValue("status")

	if userID == "" {
		http.Redirect(w, r, "/ui/actions/kyc?flash=User+is+required", http.StatusSeeOther)
		return
	}
	if gateway == "" {
		gateway = "paywiser-eu-sandbox"
	}

	user, err := h.store.GetUser(userID)
	if err != nil || user == nil {
		http.Redirect(w, r, "/ui/actions/kyc?flash=User+not+found&userID="+url.QueryEscape(userID), http.StatusSeeOther)
		return
	}

	var kycState, eventType string
	switch status {
	case "accepted":
		kycState = consts.KYCStateAccepted
		eventType = consts.WebhookEventKYCAccepted
	case "rejected":
		kycState = consts.KYCStateRejected
		eventType = consts.WebhookEventKYCRejected
	default:
		kycState = consts.KYCStateActionRequired
		eventType = consts.WebhookEventKYCActionRequired
	}

	user.KYCState = kycState
	if err := h.store.UpdateUser(user); err != nil {
		logger.Error("ui: failed to update user kyc state", zap.String("user_id", userID), zap.Error(err))
		http.Redirect(w, r, "/ui/actions/kyc?flash=Failed+to+update+user&userID="+url.QueryEscape(userID), http.StatusSeeOther)
		return
	}

	h.webhookManager.SendAsync(eventType, userID, map[string]interface{}{
		"gateway": gateway,
	}, 0)

	logger.Info("ui: kyc event triggered",
		zap.String("user_id", userID),
		zap.String("state", kycState),
		zap.String("gateway", gateway),
	)

	redirectURL := "/ui/actions/kyc?ok=1&flash=KYC+event+sent&userID=" + url.QueryEscape(userID) +
		"&status=" + url.QueryEscape(status) + "&gateway=" + url.QueryEscape(gateway)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (h *Handler) UICardTxCards(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		h.sendJSON(w, http.StatusOK, []*models.Card{})
		return
	}

	customer, err := h.store.GetCustomerBySourceID(userID)
	if err != nil || customer == nil || customer.ID == nil {
		h.sendJSON(w, http.StatusOK, []*models.Card{})
		return
	}

	all, err := h.store.GetCardsByCustomer(*customer.ID)
	if err != nil {
		h.sendJSON(w, http.StatusOK, []*models.Card{})
		return
	}

	active := make([]*models.Card, 0, len(all))
	for _, c := range all {
		if c.Status == consts.CardStatusActive {
			active = append(active, c)
		}
	}
	h.sendJSON(w, http.StatusOK, active)
}

func (h *Handler) UICardTxForm(w http.ResponseWriter, r *http.Request) {
	users, err := h.store.ListUsers()
	if err != nil {
		logger.Error("ui: failed to list users for card tx form", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	q := r.URL.Query()
	selectedUserID := q.Get("userID")

	var cards []*models.Card
	if selectedUserID != "" {
		if customer, err := h.store.GetCustomerBySourceID(selectedUserID); err == nil && customer != nil && customer.ID != nil {
			if all, err := h.store.GetCardsByCustomer(*customer.ID); err == nil {
				for _, c := range all {
					if c.Status == consts.CardStatusActive {
						cards = append(cards, c)
					}
				}
			}
		}
	}

	tmpl, err := template.ParseFS(uiTemplateFS, "web/ui/card_tx_action.html")
	if err != nil {
		logger.Error("ui: failed to parse card tx template", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, map[string]interface{}{
		"Users":          users,
		"Cards":          cards,
		"SelectedUserID": selectedUserID,
		"SelectedCardID": q.Get("cardID"),
		"TxType":         q.Get("txType"),
		"Amount":         q.Get("amount"),
		"Currency":       q.Get("currency"),
		"MerchantName":   q.Get("merchantName"),
		"Flash":          q.Get("flash"),
		"FlashOK":        q.Get("ok") == "1",
	}); err != nil {
		logger.Error("ui: failed to render card tx form", zap.Error(err))
	}
}

func (h *Handler) UICardTxAction(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	userID := r.FormValue("userID")
	cardID := r.FormValue("cardID")
	amountStr := r.FormValue("amount")
	currency := r.FormValue("currency")
	txTypeStr := r.FormValue("txType")
	merchantName := r.FormValue("merchantName")

	if cardID == "" {
		http.Error(w, "Card is required", http.StatusBadRequest)
		return
	}

	card, err := h.store.GetCard(cardID)
	if err != nil || card == nil {
		http.Error(w, "Card not found", http.StatusBadRequest)
		return
	}
	if card.Status != consts.CardStatusActive {
		http.Error(w, "Card is not active", http.StatusBadRequest)
		return
	}

	if userID == "" {
		http.Error(w, "User is required", http.StatusBadRequest)
		return
	}
	if _, err := h.store.GetUser(userID); err != nil {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	amountF, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amountF <= 0 {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	if currency == "" {
		http.Error(w, "Currency is required", http.StatusBadRequest)
		return
	}

	balance, _ := h.store.GetBalance(userID, currency)
	if balance < amountF {
		http.Error(w, fmt.Sprintf("Insufficient funds: have %.2f %s", balance, currency), http.StatusBadRequest)
		return
	}

	txType, _ := strconv.Atoi(txTypeStr)
	txID := fmt.Sprintf("tx-%s", utils.GenerateUUID()[:8])
	now := time.Now().UTC().Format(time.RFC3339)
	completed := "COMPLETED"
	amountFormatted := fmt.Sprintf("%.2f", amountF)

	var mName *string
	if merchantName != "" {
		mName = &merchantName
	}

	tx := &models.CardTransaction{
		TransactionID:         txID,
		GHResponseCode:        "00",
		GHResponseDescription: "Approved",
		TransactionAmount:     &amountFormatted,
		TransactionCurrency:   &currency,
		BillingAmount:         &amountFormatted,
		BillingCurrency:       &currency,
		Type:                  txType,
		CardScheme:            2,
		TerminalID:            "TERM001",
		CreatedAt:             now,
		TxStatus:              &completed,
		Operation:             0,
		MerchantName:          mName,
		TransactionDateTime:   &now,
		ProcessDateTime:       &now,
	}

	if err := h.store.CreateCardTransaction(tx); err != nil {
		logger.Error("ui: failed to create card transaction", zap.Error(err))
		http.Error(w, "Failed to create transaction", http.StatusInternalServerError)
		return
	}
	_ = h.store.AddCardTransactionIndex(cardID, txID)

	if err := h.store.DeductBalance(userID, currency, amountF); err != nil {
		logger.Error("ui: failed to deduct balance", zap.String("user_id", userID), zap.Error(err))
	}

	title := "Card Purchase"
	if txType == 1 {
		title = "ATM Withdrawal"
	} else if txType == 20 {
		title = "Card Refund"
	}
	displayMerchant := merchantName
	if displayMerchant == "" {
		displayMerchant = "Merchant"
	}

	h.webhookManager.SendAsync(consts.WebhookEventCardTransaction, userID, map[string]interface{}{
		"title":         title,
		"body":          fmt.Sprintf("%s %.2f at %s", currency, amountF, displayMerchant),
		"transactionId": txID,
		"cardId":        cardID,
	}, 0)

	logger.Info("ui: card transaction simulated",
		zap.String("user_id", userID),
		zap.String("card_id", cardID),
		zap.String("tx_id", txID),
		zap.Float64("amount", amountF),
		zap.String("currency", currency),
	)

	http.Redirect(w, r, "/ui/users/"+userID, http.StatusSeeOther)
}
