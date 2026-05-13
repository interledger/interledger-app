package handler

import (
	"embed"
	"encoding/json"
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

type cardWithTransactions struct {
	Card         *models.Card
	Transactions []*models.CardTransaction
}

const uiCardTxCreatedAt = "2024-02-01T00:00:00.000Z"

func defaultUICardTxObject() string {
	return `{
	"transactionAmount": "10.00",
	"transactionCurrency": "EUR",
	"billingAmount": "10.00",
	"billingCurrency": "EUR",
	"type": 0,
	"merchantName": "Coffee Shop"
}`
}

func buildUICardTransaction(rawTx string, seqID int) (*models.CardTransaction, json.RawMessage, string, error) {
	var txMap map[string]interface{}
	if err := json.Unmarshal([]byte(rawTx), &txMap); err != nil {
		return nil, nil, "", err
	}

	txID := utils.GenerateUUID()
	txMap["transactionId"] = txID
	txMap["createdAt"] = uiCardTxCreatedAt
	txMap["id"] = seqID

	normalizedRaw, err := json.Marshal(txMap)
	if err != nil {
		return nil, nil, "", err
	}

	var tx models.CardTransaction
	if err := json.Unmarshal(normalizedRaw, &tx); err != nil {
		return nil, nil, "", err
	}

	return &tx, normalizedRaw, txID, nil
}

func buildUICardTxWebhookPreview(userID, title, body, txID, cardID string) map[string]interface{} {
	return map[string]interface{}{
		"uuid":        utils.GenerateUUID(),
		"timestamp":   strconv.FormatInt(time.Now().UTC().UnixMilli(), 10),
		"event_type":  consts.WebhookEventCardTransaction,
		"user_uuid":   userID,
		"environment": "sandbox",
		"data": map[string]interface{}{
			"title":         title,
			"body":          body,
			"transactionId": txID,
			"cardId":        cardID,
		},
	}
}

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

	cardData := make([]cardWithTransactions, 0)
	if customer, err := h.store.GetCustomerBySourceID(userID); err == nil && customer != nil && customer.ID != nil {
		if cards, err := h.store.GetCardsByCustomer(*customer.ID); err == nil {
			for _, card := range cards {
				txIDs, err := h.store.GetCardTransactionIDs(card.ID)
				if err != nil {
					logger.Error("ui: failed to list card transaction ids", zap.String("card_id", card.ID), zap.Error(err))
					txIDs = nil
				}

				cardTxs := make([]*models.CardTransaction, 0, len(txIDs))
				for _, txID := range txIDs {
					tx, err := h.store.GetCardTransaction(txID)
					if err != nil {
						logger.Error("ui: failed to load card transaction", zap.String("tx_id", txID), zap.Error(err))
						continue
					}
					cardTxs = append(cardTxs, tx)
				}

				cardData = append(cardData, cardWithTransactions{Card: card, Transactions: cardTxs})
			}
		}
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
		"Cards":        cardData,
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
	title := q.Get("title")
	body := q.Get("body")
	cardTxObject := q.Get("cardTxObject")
	if title == "" {
		title = "Card Purchase"
	}
	if body == "" {
		body = "EUR 10.00 at Coffee Shop"
	}
	if cardTxObject == "" {
		cardTxObject = defaultUICardTxObject()
	}

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
		"Title":          title,
		"Body":           body,
		"CardTxObject":   cardTxObject,
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
	title := r.FormValue("title")
	body := r.FormValue("body")
	rawTx := r.FormValue("cardTxObject")

	if cardID == "" {
		http.Error(w, "Card is required", http.StatusBadRequest)
		return
	}
	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}
	if body == "" {
		http.Error(w, "Body is required", http.StatusBadRequest)
		return
	}
	if rawTx == "" {
		http.Error(w, "Card transaction object is required", http.StatusBadRequest)
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

	seqID, err := h.store.NextCardTransactionSeqID()
	if err != nil {
		logger.Error("ui: failed to get next card transaction sequence id", zap.Error(err))
		http.Error(w, "Failed to create transaction", http.StatusInternalServerError)
		return
	}

	tx, normalizedRaw, txID, err := buildUICardTransaction(rawTx, seqID)
	if err != nil {
		http.Error(w, "Invalid card transaction JSON object", http.StatusBadRequest)
		return
	}

	if err := h.store.CreateCardTransaction(tx); err != nil {
		logger.Error("ui: failed to create card transaction", zap.Error(err))
		http.Error(w, "Failed to create transaction", http.StatusInternalServerError)
		return
	}
	if err := h.store.StoreRawCardTransaction(txID, normalizedRaw); err != nil {
		logger.Error("ui: failed to store raw card transaction payload", zap.Error(err))
		http.Error(w, "Failed to create transaction", http.StatusInternalServerError)
		return
	}
	_ = h.store.AddCardTransactionIndex(cardID, txID)

	webhookData := map[string]interface{}{
		"title":         title,
		"body":          body,
		"transactionId": txID,
		"cardId":        cardID,
	}

	h.webhookManager.SendAsync(consts.WebhookEventCardTransaction, userID, webhookData, 0)

	logger.Info("ui: card transaction simulated",
		zap.String("user_id", userID),
		zap.String("card_id", cardID),
		zap.String("tx_id", txID),
		zap.Int("tx_seq_id", seqID),
		zap.String("title", title),
	)

	http.Redirect(w, r, "/ui/users/"+userID, http.StatusSeeOther)
}

func (h *Handler) UICardTxPreview(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.sendJSON(w, http.StatusBadRequest, map[string]interface{}{"error": "Invalid form data"})
		return
	}

	userID := r.FormValue("userID")
	cardID := r.FormValue("cardID")
	title := r.FormValue("title")
	body := r.FormValue("body")
	rawTx := r.FormValue("cardTxObject")

	if cardID == "" || userID == "" || title == "" || body == "" || rawTx == "" {
		h.sendJSON(w, http.StatusBadRequest, map[string]interface{}{"error": "All fields are required"})
		return
	}

	seqID, err := h.store.PeekCardTransactionSeqID()
	if err != nil {
		h.sendJSON(w, http.StatusInternalServerError, map[string]interface{}{"error": "Failed to build preview"})
		return
	}

	_, normalizedRaw, txID, err := buildUICardTransaction(rawTx, seqID)
	if err != nil {
		h.sendJSON(w, http.StatusBadRequest, map[string]interface{}{"error": "Invalid card transaction JSON object"})
		return
	}

	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"webhook":           buildUICardTxWebhookPreview(userID, title, body, txID, cardID),
		"transaction":       json.RawMessage(normalizedRaw),
		"nextTransactionId": seqID,
	})
}
