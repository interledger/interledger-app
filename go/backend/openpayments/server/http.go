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

	"github.com/go-chi/chi/v5"
	"github.com/riandyrn/otelchi"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/openpayments/ops"
	"gitlab.com/fynbos/backend/openpayments/workflows"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

func OpenPaymentsHTTPHandler(b Backends) http.Handler {
	router := chi.NewRouter()
	router.Use(otelchi.Middleware("open_payments", otelchi.WithChiRoutes(router)))

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

// postHandler handles all incoming POST requests and routes them according to the URL contents and suffixing.
func postHandler(b Backends, w http.ResponseWriter, req *http.Request) {

	ctx := req.Context()
	fURL := getFullURL(req)
	ppURL, suffix, err := ops.ExtractPaymentPointer(fURL)

	if err != nil {
		log.Error("error trying to parse post to payment pointer", zap.Error(err), zap.String("url", req.URL.String()))
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
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
		log.Error("failed to get payment pointer", zap.Error(err), zap.String("url", req.URL.String()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	switch suffix {
	case "outgoing-payments":
		createOutgoingPayment(b, w, req, pp)
		return
	case "incoming-payments":
		createIncomingPayment(b, w, req, pp)
		return
	case "quotes":
		createQuote(b, w, req, pp)
		return
	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

func createOutgoingPayment(b Backends, w http.ResponseWriter, req *http.Request, pp *openpayments.PaymentPointer) {
	bodyData, err := io.ReadAll(req.Body)
	if err != nil {
		log.Error("failed to decode create outgoing payment body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	var args openpayments.CreateOutgoingPaymentArgs
	err = json.Unmarshal(bodyData, &args)
	if err != nil {
		log.Error("failed to unmarshal create outgoing payment body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	_, err = ops.GetPaymentPointerQuote(req.Context(), b, pp.ID, args.QuoteID)
	if errors.Is(err, openpayments.ErrNotFound) {
		log.Error("failed to start outgoing payment, quote not found", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err != nil {
		log.Error("failed to start outgoing payment, could not look up quote", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
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
		log.Error("failed to write create quote response", zap.Error(err))
	}
}

func createQuote(b Backends, w http.ResponseWriter, req *http.Request, pp *openpayments.PaymentPointer) {
	bodyData, err := io.ReadAll(req.Body)
	if err != nil {
		log.Error("failed to decode create quote body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	var args openpayments.CreateQuoteArgs
	err = json.Unmarshal(bodyData, &args)
	if err != nil {
		log.Error("failed to unmarshal create quote body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	args.ReceivePaymentPointer = ops.StandardisePaymentPointer(args.ReceivePaymentPointer)
	args.SendPaymentPointer = pp.URL

	if args.ExpiresAt.IsZero() {
		// Default to a 10 minute time to create a quote
		args.ExpiresAt = time.Now().Add(time.Minute * 10)
	}

	q, err := ops.CreateQuote(req.Context(), b, args)
	if errors.Is(err, openpayments.ErrInvalidArgument) {
		log.Error("invalid arguments to create quote", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if errors.Is(err, openpayments.ErrPaymentPointerNotFound) || errors.Is(err, openpayments.ErrNotFound) {
		log.Error("not found arguments to create quote", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err != nil {
		log.Error("failed to create quote", zap.Error(err))
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
		log.Error("failed to write create quote response", zap.Error(err))
	}
}

func createIncomingPayment(b Backends, w http.ResponseWriter, req *http.Request, pp *openpayments.PaymentPointer) {
	bodyData, err := io.ReadAll(req.Body)
	if err != nil {
		log.Error("failed to decode create incoming payment body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	var args openpayments.CreateIncomingPaymentArgs
	err = json.Unmarshal(bodyData, &args)
	if err != nil {
		log.Error("failed to unmarshal create incoming payment body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	args.FromPaymentPointer = ops.StandardisePaymentPointer(args.FromPaymentPointer)
	args.PaymentPointer = pp.URL

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
	case "quotes":
		getQuote(b, pp, w, req)
		return
	case "incoming-payments":
		getIncomingPayment(b, w, req)
		return
	case "outgoing-payments":
		getOutgoingPayment(b, w, req)
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

func getQuote(b Backends, pp *openpayments.PaymentPointer, w http.ResponseWriter, req *http.Request) {
	q, err := ops.GetPaymentPointerQuote(req.Context(), b, pp.ID, getFullURL(req))
	if errors.Is(err, openpayments.ErrNotFound) {
		log.Error("quote not found", zap.Error(err), zap.String("url", getFullURL(req)))
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err != nil {
		log.Error("failed to get quote", zap.Error(err), zap.String("url", getFullURL(req)))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(q)
	if err != nil {
		log.Error("error writing get quote http response", zap.Error(err), zap.String("url", getFullURL(req)))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func getOutgoingPayment(b Backends, w http.ResponseWriter, req *http.Request) {
	op, err := ops.GetOutgoingPayment(req.Context(), b, getFullURL(req))
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(op)
	if err != nil {
		log.Error("error writing get quote http response", zap.Error(err), zap.String("url", getFullURL(req)))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func getIncomingPayment(b Backends, w http.ResponseWriter, req *http.Request) {
	ip, err := ops.GetIncomingPayment(req.Context(), b, getFullURL(req))
	if errors.Is(err, openpayments.ErrNotFound) {
		log.Error("incoming payment not found", zap.Error(err), zap.String("url", getFullURL(req)))
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err != nil {
		log.Error("failed to get quote", zap.Error(err), zap.String("url", getFullURL(req)))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ip)
	if err != nil {
		log.Error("error writing get quote http response", zap.Error(err), zap.String("url", getFullURL(req)))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
