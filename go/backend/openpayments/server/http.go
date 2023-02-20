package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"path"
	"strings"
	"time"

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

	router.Post("/incoming_payment", createIncomingPayment(b))
	router.Get("/incoming_payment/{payment_id}", getIncomingPayment(b))

	router.Post("/outgoing_payment", createOutgoingPayment(b))
	router.Get("/outgoing_payment/{payment_id}", getOutgoingPayment(b))

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
	fURL := getFullURL(req)
	ppURL, _, err := ops.ExtractPaymentPointer(fURL)

	if err != nil {
		log.Error("error trying to parse post to payment pointer", zap.Error(err), zap.String("url", req.URL.String()))
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	_, err = ops.GetPaymentPointer(ctx, b, ppURL)
	if err != nil {
		if errors.Is(err, openpayments.ErrPaymentPointerNotFound) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		if errors.Is(err, openpayments.ErrInvalidPointerURL) {
			http.Error(w, "Invalid Payment Pointer", http.StatusBadRequest)
			return
		}
		log.Error("failed to get payment pointer", zap.Error(err), zap.String("url", req.URL.String()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, req, env.AuthURL(), http.StatusSeeOther)
}

type authoriseClientArgs struct {
	Request           *http.Request
	AccessType        string
	Action            string
	Identifier        string
	ResourceCreatedBy string
}

func authoriseClient(b Backends, args authoriseClientArgs) string {
	gnapToken := args.Request.Header.Get("authorization")
	parts := strings.Split(args.Request.Header.Get("authorization"), " ")
	if len(parts) > 1 && parts[0] == "GNAP" {
		gnapToken = parts[1]
	}
	grant, err := b.Authorisation().Introspect(args.Request.Context(), gnapToken)
	if err != nil {
		log.Error("token introspection failed", zap.Error(err))
		return ""
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
		return ""
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

	client := ""
	if hasAccess {
		client = grant.Client
	}

	return client
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

		client := authoriseClient(b, authoriseClientArgs{
			Request:    req,
			AccessType: "outgoing-payment",
			Action:     "write",
			Identifier: httpArgs.FromPP,
		})
		if client == "" {
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

		if !b.Authorisation().VerifyRequestSig(req.Context(), req, client, []string{"content-digest", "authorization"}) {
			log.Error("create outgoing payment request failed signature validation", zap.String("client", client), zap.Error(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		argAmount := currency.FromFloat64(httpArgs.SendAmount.Amount, currency.ParseCurrency(httpArgs.SendAmount.Currency))

		// Create quote transparently
		q, err := ops.CreateQuote(req.Context(), b, openpayments.CreateQuoteArgs{
			SendPaymentPointer:    httpArgs.FromPP,
			ReceivePaymentPointer: httpArgs.ToPP,
			SendAmount:            argAmount,
			Reference:             httpArgs.ExternalRef,
			ExpiresAt:             time.Now().Add(time.Hour), // Default till the API definition changes
			CreatedBy:             client,
		})
		if err != nil {
			log.Error("failed to create quote for outgoing payment", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		args := openpayments.CreateOutgoingPaymentArgs{
			QuoteID:     q.ID,
			Description: httpArgs.ExternalRef,
			ExternalRef: httpArgs.ExternalRef,
			CreatedBy:   client,
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

		client := authoriseClient(b, authoriseClientArgs{
			Request:    req,
			AccessType: "incoming-payment",
			Action:     "write",
			Identifier: httpArgs.ToPP,
		})
		if client == "" {
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

		if !b.Authorisation().VerifyRequestSig(req.Context(), req, client, []string{"content-digest", "authorization"}) {
			log.Error("create incoming payment request failed signature validation", zap.String("client", client), zap.Error(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		var argAmount *currency.Amount
		if httpArgs.IncomingAmount != nil {
			tmp := currency.FromFloat64(httpArgs.IncomingAmount.Amount, currency.ParseCurrency(httpArgs.IncomingAmount.Currency))
			argAmount = &tmp
		}

		args := openpayments.CreateIncomingPaymentArgs{
			PaymentPointer:     ops.StandardisePaymentPointer(httpArgs.ToPP),
			FromPaymentPointer: ops.StandardisePaymentPointer(httpArgs.FromPP),
			IncomingAmount:     argAmount,
			ExternalRef:        httpArgs.ExternalRef,
			CreatedBy:          client,
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

// getHandler handles all incoming GET requests and handles them as required
func getHandler(b Backends, w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	ppURL, suffix, err := ops.ExtractPaymentPointer(getFullURL(req))
	if errors.Is(err, openpayments.ErrInvalidPointerURL) {
		http.Error(w, "Invalid Payment Pointer", http.StatusBadRequest)
		return
	}

	pp, err := ops.GetPaymentPointer(ctx, b, ppURL)
	if err != nil {
		if errors.Is(err, openpayments.ErrPaymentPointerNotFound) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		if errors.Is(err, openpayments.ErrInvalidPointerURL) {
			http.Error(w, "Invalid Payment Pointer", http.StatusBadRequest)
			return
		}

		log.Error("error when getting payment pointer via HTTP", zap.Error(err), zap.String("url", req.URL.String()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	switch suffix {
	case ".well-known/keys":
		listClientKeys(b, pp, w, req)
		return
	}

	// Fallback to get payment pointer
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(pp)
	if err != nil {
		log.Error("error writing http response", zap.Error(err), zap.String("url", req.URL.String()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func listClientKeys(b Backends, pp *openpayments.PaymentPointer, w http.ResponseWriter, req *http.Request) {
	// we assume the client url has been registered as the pp
	clientURL := pp.URL
	keys, err := b.Authorisation().ListKeys(req.Context(), clientURL)
	if errors.Is(err, authorisation.ErrNotFound) {
		log.Error("Failed to list client keys. clientURL not found.", zap.Error(err), zap.String("clientURL", clientURL))
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
	if err != nil {
		log.Error("error listing client public keys", zap.Error(err), zap.String("clientURL", clientURL))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	keySet := struct {
		Keys []authorisation.Jwk `json:"keys"`
	}{
		Keys: keys,
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(keySet)
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

		client := authoriseClient(b, authoriseClientArgs{
			Request:           req,
			AccessType:        "outgoing-payment",
			Action:            "read",
			Identifier:        op.PaymentPointer,
			ResourceCreatedBy: op.CreatedBy,
		})
		if client == "" {
			log.Error("token does not have required access")
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		// Verify client signature.
		if !b.Authorisation().VerifyRequestSig(req.Context(), req, client, []string{"authorization"}) {
			log.Error("get outgoing payment request failed signature validation", zap.String("client", client), zap.Error(err))
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

		client := authoriseClient(b, authoriseClientArgs{
			Request:           req,
			AccessType:        "incoming-payment",
			Action:            "read",
			Identifier:        ip.PaymentPointer,
			ResourceCreatedBy: ip.CreatedBy,
		})
		if client == "" {
			log.Error("token does not have required access")
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		// Verify client signature.
		if !b.Authorisation().VerifyRequestSig(req.Context(), req, client, []string{"authorization"}) {
			log.Error("get incoming payment request failed signature validation", zap.String("client", client), zap.Error(err))
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
