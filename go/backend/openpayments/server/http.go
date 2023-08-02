package server

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/wallets"

	"gitlab.com/fynbos/backend/identities"

	"gitlab.com/fynbos/httpmessagesignatures"

	"gitlab.com/fynbos/env"

	"github.com/go-chi/chi/v5"
	"github.com/riandyrn/otelchi"
	"gitlab.com/fynbos/backend/authorisation"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/openpayments/ops"
	"gitlab.com/fynbos/backend/openpayments/workflows"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

func OpenPaymentsHTTPHandler(b Backends) http.Handler {
	router := chi.NewRouter()
	router.Use(otelchi.Middleware("open_payments", otelchi.WithChiRoutes(router)))

	router.Post("/incoming", createIncomingPayment(b))
	router.Get("/incoming/{payment_id}", getIncomingPayment(b))

	router.Post("/outgoing", createOutgoingPayment(b))
	router.Get("/outgoing/{payment_id}", getOutgoingPayment(b))

	router.Get("/{wallet_id}/identities/{identity_sig_hash}", getIdentity(b))

	router.NotFound(catchAllHandler(b))
	return router
}

func catchAllHandler(b Backends) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
			postHandler(b, w, req)
		} else if req.Method == http.MethodGet {
			getHandler(b, w, req)
		} else {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		}
	}
}

func getFullURL(req *http.Request) string {
	url := req.URL.String()
	if req.URL.Scheme == "" && req.URL.Host == "" {
		url = fmt.Sprintf("https://%s", path.Join(req.Host, req.URL.Path))
	}

	if strings.HasPrefix(url, "http://") {
		url = strings.Replace(url, "http://", "https://", 1)
	}

	return url
}

// postHandler handles all incoming POST requests that do not match predefined routes.
// All posts are considered to be GNAP Auth requests to an existing payment pointer.
func postHandler(b Backends, w http.ResponseWriter, req *http.Request) {

	ctx := req.Context()
	ppURL := getFullURL(req)

	_, err := b.Wallets().GetFromAddress(ctx, ppURL)
	if err != nil {
		if errors.Is(err, wallets.ErrNoWalletFound) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		if errors.Is(err, wallets.ErrInvalidAddress) {
			http.Error(w, "Invalid Payment Pointer", http.StatusBadRequest)
			return
		}
		log.Error("failed to get payment pointer", zap.Error(err), zap.String("url", req.URL.String()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, req, env.AuthURL(), http.StatusSeeOther)
}

type authoriseGrantArgs struct {
	Request           *http.Request
	AccessType        string
	Action            string
	Identifier        string
	ResourceCreatedBy string
}

func authoriseGrant(b Backends, args authoriseGrantArgs) *authorisation.Grant {
	gnapToken := args.Request.Header.Get("authorization")
	parts := strings.Split(args.Request.Header.Get("authorization"), " ")
	if len(parts) > 1 && parts[0] == "GNAP" {
		gnapToken = parts[1]
	}
	grant, err := b.Authorisation().Introspect(args.Request.Context(), gnapToken)
	if err != nil {
		log.Error("token introspection failed", zap.Error(err))
		return nil
	}

	var accessToken *authorisation.AccessToken
	for _, t := range grant.Tokens {
		if gnapToken == t.Value {
			accessToken = &t
			break
		}
	}
	if accessToken == nil {
		log.Error("grant does not have token", zap.String("grantID", grant.ID), zap.String("token", gnapToken))
		return nil
	}

	hasAccess := false
	for _, access := range accessToken.Access {
		if access.Type != args.AccessType {
			continue
		}

		// TODO: define policies better
		for _, act := range access.Actions {
			if act == "write" && access.Identifier == args.Identifier {
				hasAccess = true
				break
			}

			if act == "read" && access.Identifier == args.Identifier && grant.Client == args.ResourceCreatedBy {
				hasAccess = true
				break
			}
		}
	}

	if !hasAccess {
		return nil
	}

	return grant
}

type OutgoingPaymentArgs struct {
	FromPP      string `json:"wallet"`
	Type        string `json:"type"`
	ToPP        string `json:"id"`
	ExternalRef string `json:"external_ref"`
	SendAmount  struct {
		Amount   float64 `json:"amount,string"`
		Currency string  `json:"currency"`
	} `json:"send_amount"`
}

func createOutgoingPayment(b Backends) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		bodyData, err := io.ReadAll(req.Body)
		if err != nil {
			log.Error("failed to decode create outgoing payment body", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		defer req.Body.Close()

		var httpArgs OutgoingPaymentArgs
		err = json.Unmarshal(bodyData, &httpArgs)
		if err != nil {
			log.Error("failed to unmarshal create outgoing payment body", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		grant := authoriseGrant(b, authoriseGrantArgs{
			Request:    req,
			AccessType: "outgoing-payment",
			Action:     "write",
			Identifier: httpArgs.FromPP,
		})
		if grant == nil {
			log.Error("token does not have required access")
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		// Verify content-digest and client signature
		if err = httpmessagesignatures.VerifyContentDigest(req.Context(), req.Header.Get("Content-Digest"), bodyData); err != nil {
			log.Error("create outgoing payment request does not match Content-Digest header.", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if !b.Authorisation().VerifyRequestSig(req.Context(), req, grant.Client, []string{"content-digest", "authorization"}) {
			log.Error("create outgoing payment request failed signature validation", zap.String("client", grant.Client), zap.Error(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		argAmount := currency.FromFloat64(httpArgs.SendAmount.Amount, currency.ParseCurrency(httpArgs.SendAmount.Currency))

		fromWallet, err := b.Wallets().GetFromAddress(req.Context(), httpArgs.FromPP)
		if err != nil {
			log.Error("failed to lookup from payment pointer", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Check the limits service to see if the grant client has not exceeded its limits.
		exceeds, err := b.Limits().Exceeds(req.Context(), fromWallet.ID, grant.Client, argAmount)
		if err != nil {
			log.Error("failed to check limits of the grant client", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if exceeds {
			http.Error(w, "usage limits exceeded", http.StatusForbidden)
			return
		}

		exceedsKyc, _, err := b.Limits().ExceedsKYCLimits(req.Context(), fromWallet.ID, argAmount)
		if err != nil {
			log.Error("failed to check limits of the grant client", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if exceedsKyc {
			http.Error(w, "kyc limits exceeded", http.StatusForbidden)
			return
		}

		// Create quote transparently
		q, err := ops.CreateQuote(req.Context(), b, openpayments.CreateQuoteArgs{
			SendPaymentPointer:    httpArgs.FromPP,
			ReceivePaymentPointer: httpArgs.ToPP,
			SendAmount:            argAmount,
			Reference:             httpArgs.ExternalRef,
			ExpiresAt:             time.Now().Add(time.Hour), // Default till the API definition changes
			CreatedBy:             grant.Client,
		})
		if err != nil {
			log.Error("failed to create quote for outgoing payment", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		args := openpayments.CreateOutgoingPaymentArgs{
			QuoteID:     q.ID,
			Description: httpArgs.ExternalRef,
			CreatedBy:   grant.Client,
			GrantID:     grant.ID,
		}

		// Extract IP address from the req
		ip, _, err := net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			log.Error("failed to get ip address from http request, falling back to headers", zap.Error(err), zap.String("remoteAddress", req.RemoteAddr))
		}
		fips := strings.Split(req.Header.Get("X-Forwarded-For"), ",")
		if len(fips) > 0 {
			ip = strings.TrimSpace(fips[0])
		}
		args.IPAddress = ip

		op, err := workflows.StartOutgoingPayment(req.Context(), b, args)
		if errors.Is(err, openpayments.ErrNotFound) {
			log.Error("failed to start outgoing payment, values not found", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		if errors.Is(err, openpayments.ErrPaymentPointerCannotSend) {
			log.Info("payment pointer cannot send", zap.String("fromLinkedAccount", q.FromLinkedAccount))
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		}
		if errors.Is(err, openpayments.ErrPaymentPointerCannotRecv) {
			log.Info("payment pointer cannot receive", zap.String("payment pointer", q.PaymentPointer))
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		}
		if err != nil {
			log.Error("failed to start outgoing payment", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		respBytes, err := json.Marshal(op)
		if err != nil {
			log.Error("failed to marshall create outgoing payment response", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(respBytes)
		if err != nil {
			log.Error("failed to write outgoing payment response", zap.Error(err))
		}
	}
}

type IncomingPaymentArgs struct {
	FromPP         string `json:"id"`
	Type           string `json:"type"`
	ToPP           string `json:"wallet"`
	ExternalRef    string `json:"external_ref"`
	IncomingAmount *struct {
		Amount   float64 `json:"amount,string"`
		Currency string  `json:"currency"`
	} `json:"incoming_amount"`
}

func createIncomingPayment(b Backends) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		bodyData, err := io.ReadAll(req.Body)
		if err != nil {
			log.Error("failed to decode create incoming payment body", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		defer req.Body.Close()

		var httpArgs IncomingPaymentArgs
		err = json.Unmarshal(bodyData, &httpArgs)
		if err != nil {
			log.Error("failed to unmarshal create incoming payment body", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		grant := authoriseGrant(b, authoriseGrantArgs{
			Request:    req,
			AccessType: "incoming-payment",
			Action:     "write",
			Identifier: httpArgs.ToPP,
		})
		if grant == nil {
			log.Error("token does not have required access")
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		// Verify content-digest and client signature
		if err = httpmessagesignatures.VerifyContentDigest(req.Context(), req.Header.Get("Content-Digest"), bodyData); err != nil {
			log.Error("create incoming payment request does not match Content-Digest header.", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if !b.Authorisation().VerifyRequestSig(req.Context(), req, grant.Client, []string{"content-digest", "authorization"}) {
			log.Error("create incoming payment request failed signature validation", zap.String("client", grant.Client), zap.Error(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		var argAmount *currency.Amount
		if httpArgs.IncomingAmount != nil {
			tmp := currency.FromFloat64(httpArgs.IncomingAmount.Amount, currency.ParseCurrency(httpArgs.IncomingAmount.Currency))
			argAmount = &tmp
		}

		senderWa, err := wallets.ParseAddress(httpArgs.FromPP)
		if err != nil {
			http.Error(w, "Invalid Payment Pointer", http.StatusBadRequest)
			return
		}

		receiverWa, err := wallets.ParseAddress(httpArgs.ToPP)
		if err != nil {
			http.Error(w, "Invalid Payment Pointer", http.StatusBadRequest)
			return
		}

		args := openpayments.CreateIncomingPaymentArgs{
			PaymentPointer:     receiverWa.String(),
			FromPaymentPointer: senderWa.String(),
			IncomingAmount:     argAmount,
			ExternalRef:        httpArgs.ExternalRef,
			CreatedBy:          grant.Client,
		}

		q, err := ops.CreateIncomingPayment(req.Context(), b, args)
		if errors.Is(err, openpayments.ErrInvalidArgument) {
			log.Error("invalid arguments to create quote", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if err != nil {
			log.Error("failed to create incoming payment", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		respBytes, err := json.Marshal(q)
		if err != nil {
			log.Error("failed to marshall create quote response", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(respBytes)
		if err != nil {
			log.Error("failed to write create incoming response", zap.Error(err))
		}
	}
}

type JsonResponse struct {
	Id         string             `json:"id"`
	PublicName string             `json:"publicName"`
	Identities []IdentityResponse `json:"identities"`
}

type IdentityResponse struct {
	Identifier    string `json:"identifier"`
	Kid           string `json:"kid"`
	Ctime         int64  `json:"ctime"`
	Signature     string `json:"signature"`
	SignatureHash string `json:"signature_hash"`
	PublicProof   string `json:"public_proof"`
	Type          string `json:"type"`
}

// getHandler handles all incoming GET requests and handles them as required
func getHandler(b Backends, w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	ppURL, suffix, err := ops.ExtractPaymentPointer(getFullURL(req))
	if errors.Is(err, wallets.ErrInvalidAddress) {
		http.Error(w, "Invalid Payment Pointer", http.StatusBadRequest)
		return
	}

	wallet, err := b.Wallets().GetFromAddress(ctx, ppURL)
	if err != nil {
		if errors.Is(err, wallets.ErrNoWalletFound) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		if errors.Is(err, wallets.ErrInvalidAddress) {
			http.Error(w, "Invalid Payment Pointer", http.StatusBadRequest)
			return
		}

		log.Error("error when getting payment pointer via HTTP", zap.Error(err), zap.String("url", req.URL.String()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Check if the content type is from browser and redirect
	if strings.Contains(req.Header.Get("Accept"), "text/html") {
		u, err := url.JoinPath(env.GetUrl(), "/me/", removeProtocol(wallet.Addresses[0].String()))
		if err != nil {
			log.Error("error generating url", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, req, u, http.StatusFound)
		return
	}

	switch suffix {
	case "jwks.json":
		listKeys(b, wallet.ID, w, req)
		return
	}

	ids, err := b.Identities().ListPublic(ctx, wallet.ID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	jsonIds := make([]IdentityResponse, 0)

	for _, id := range ids {
		if id.State == identities.StateVerified {
			sigBase64 := base64.URLEncoding.EncodeToString(id.Signature)
			sigHashBase64 := base64.URLEncoding.EncodeToString(id.SignatureHash)

			jsonIds = append(jsonIds, IdentityResponse{
				Identifier:    id.Identifier,
				Kid:           id.KeyID,
				Ctime:         id.CreatedAt.Unix(),
				Type:          string(id.Platform),
				Signature:     sigBase64,
				SignatureHash: sigHashBase64,
				PublicProof:   id.VerificationProof,
			})
		}
	}

	jsonResponse := JsonResponse{
		Id:         wallet.Addresses[0].String(),
		PublicName: wallet.Name,
		Identities: jsonIds,
	}

	// Fallback to get payment pointer
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(jsonResponse)
	if err != nil {
		log.Error("error writing http response", zap.Error(err), zap.String("url", req.URL.String()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func listKeys(b Backends, walletID string, w http.ResponseWriter, req *http.Request) {
	keys, err := b.Keys().List(req.Context(), walletID)
	if err != nil {
		log.Error("error listing client public keys", zap.Error(err), zap.String("walletID", walletID))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	jwks := make([]openpayments.Jwk, len(keys))
	for i, k := range keys {
		jwks[i] = openpayments.Jwk{
			Kty: "OKP",
			Kid: k.ID,
			Crv: "Ed25519",
			Alg: "edDSA",
			Use: "sign",
			X:   k.PublicKey,
		}
	}

	resp := struct {
		Keys []openpayments.Jwk `json:"keys"`
	}{
		Keys: jwks,
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Error("error writing http response", zap.Error(err), zap.String("url", req.URL.String()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func getOutgoingPayment(b Backends) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		id := chi.URLParam(req, "payment_id")

		op, err := ops.GetOutgoingPayment(req.Context(), b, id)
		if errors.Is(err, openpayments.ErrNotFound) {
			log.Error("outgoing payment not found", zap.Error(err), zap.String("url", getFullURL(req)))
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		if err != nil {
			log.Error("failed to get outgoing payment", zap.Error(err), zap.String("url", getFullURL(req)))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		grant := authoriseGrant(b, authoriseGrantArgs{
			Request:           req,
			AccessType:        "outgoing-payment",
			Action:            "read",
			Identifier:        op.PaymentPointer,
			ResourceCreatedBy: op.CreatedBy,
		})
		if grant == nil {
			log.Error("token does not have required access")
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		// Verify client signature.
		if !b.Authorisation().VerifyRequestSig(req.Context(), req, grant.Client, []string{"authorization"}) {
			log.Error("get outgoing payment request failed signature validation", zap.String("client", grant.Client), zap.Error(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(op)
		if err != nil {
			log.Error("error writing get outgoing payment http response", zap.Error(err), zap.String("url", getFullURL(req)))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}

}

func getIncomingPayment(b Backends) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		id := chi.URLParam(req, "payment_id")
		ip, err := ops.GetIncomingPayment(req.Context(), b, id)
		if errors.Is(err, openpayments.ErrNotFound) {
			log.Error("incoming payment not found", zap.Error(err), zap.String("url", getFullURL(req)))
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		if err != nil {
			log.Error("failed to get incoming payment", zap.Error(err), zap.String("url", getFullURL(req)))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		grant := authoriseGrant(b, authoriseGrantArgs{
			Request:           req,
			AccessType:        "incoming-payment",
			Action:            "read",
			Identifier:        ip.PaymentPointer,
			ResourceCreatedBy: ip.CreatedBy,
		})
		if grant == nil {
			log.Error("token does not have required access")
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		// Verify client signature.
		if !b.Authorisation().VerifyRequestSig(req.Context(), req, grant.Client, []string{"authorization"}) {
			log.Error("get incoming payment request failed signature validation", zap.String("client", grant.Client), zap.Error(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(ip)
		if err != nil {
			log.Error("error writing get incoming payment http response", zap.Error(err), zap.String("url", getFullURL(req)))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
}

func getIdentity(b Backends) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		identitySigHash := chi.URLParam(req, "identity_sig_hash")
		ppURL, _, err := ops.ExtractPaymentPointer(getFullURL(req))
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		wallet, err := b.Wallets().GetFromAddress(req.Context(), ppURL)
		if err != nil {
			if errors.Is(err, wallets.ErrNoWalletFound) {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		sigHash, err := base64.URLEncoding.DecodeString(identitySigHash)
		if err != nil {
			// Leave as not found as decoding errors will give 500's
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		identity, err := b.Identities().GetBySignatureHash(req.Context(), sigHash)
		if err != nil {
			if errors.Is(err, identities.ErrNotFound) {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if identity.WalletID != wallet.ID {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		// Don't allow non verified and non public ones to be shown
		if identity.State != identities.StateVerified || !identity.Public {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		if isSocialMediaScraper(req) {
			u, err := url.JoinPath(env.GetUrl(), "me/identities", identitySigHash)
			if err != nil {
				log.Error("error generate url", zap.Error(err))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			// get the html body from the above url
			resp, err := http.Get(u)
			if err != nil {
				log.Error("error getting url", zap.Error(err))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			b, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Error("error reading response body", zap.Error(err))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()

			// return bytes to the caller as html response
			w.Header().Set("Content-Type", "text/html")
			_, err = w.Write(b)
			if err != nil {
				log.Error("error writing response body", zap.Error(err))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			return
		}

		// if text html redirect
		if strings.Contains(req.Header.Get("Accept"), "text/html") {
			u, err := url.JoinPath(env.GetUrl(), "me/identities", identitySigHash)
			if err != nil {
				log.Error("error generate url", zap.Error(err))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, req, u, http.StatusSeeOther)
			return
		}

		sigBase64 := base64.URLEncoding.EncodeToString(identity.Signature)
		sigHashBase64 := base64.URLEncoding.EncodeToString(identity.SignatureHash)
		jsonResp := IdentityResponse{
			Identifier:    identity.Identifier,
			Kid:           identity.KeyID,
			Ctime:         identity.CreatedAt.Unix(),
			Type:          string(identity.Platform),
			Signature:     sigBase64,
			SignatureHash: sigHashBase64,
			PublicProof:   identity.VerificationProof,
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(jsonResp)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func isSocialMediaScraper(req *http.Request) bool {
	ua := strings.ToLower(req.UserAgent())

	// Check if the User-Agent contains any of the desired strings
	if strings.Contains(ua, "linkedinbot") ||
		strings.Contains(ua, "facebookexternalhit") ||
		strings.Contains(ua, "facebookcatalog") ||
		strings.Contains(ua, "twitterbot") {
		return true
	}

	return false
}

func removeProtocol(url string) string {
	url = strings.Replace(url, "http://", "", 1)
	return strings.Replace(url, "https://", "", 1)
}
