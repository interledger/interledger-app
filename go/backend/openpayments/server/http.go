package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"

	"github.com/go-chi/chi/v5"
	"github.com/riandyrn/otelchi"
	"go.uber.org/zap"

	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/openpayments/ops"
	"gitlab.com/fynbos/log"
)

func StartOpenPaymentsHTTP(b Backends, port string) {
	router := chi.NewRouter()
	router.Use(otelchi.Middleware("open_payments", otelchi.WithChiRoutes(router)))

	router.NotFound(catchAllHandler(b))
	log.Info("payment pointers served on http://localhost:{port}/{paymentpointer}", zap.String("port", port))
	go func() {
		log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
	}()
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
	url := path.Join(req.Host, req.URL.String())
	if req.TLS == nil {
		return fmt.Sprintf("http://%s", url)
	}
	return fmt.Sprintf("https://%s")
}

// postHandler handles all incoming POST requests and routes them according to the URL contents and suffixing.
func postHandler(b Backends, w http.ResponseWriter, req *http.Request) {

	ctx := req.Context()
	ppURL, suffix, err := ops.ExtractPaymentPointer(getFullURL(req))

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

	switch suffix {
	case "outgoing-payments":
		// TODO: handle
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		return
	case "incoming-payments":
		// TODO: handle
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		return
	case "quote":
		// TODO: handle
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		return
	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

// getHandler handles all incoming GET requests and handles them as required
func getHandler(b Backends, w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	pp, err := ops.GetPaymentPointer(ctx, b, getFullURL(req))
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

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(pp)
	if err != nil {
		log.Error("error writing http response", zap.Error(err), zap.String("url", req.URL.String()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
