package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

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

	router.Get("/{payment_pointer}", getPaymentPointer(b))
	log.Info("payment pointers served on http://localhost:{port}/{paymentpointer}", zap.String("port", port))
	go func() {
		log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
	}()
}

func getPaymentPointer(b Backends) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		pp, err := ops.GetPaymentPointer(ctx, b, req.URL.String())
		if err != nil {
			if errors.Is(err, openpayments.ErrPaymentPointerNotFound) {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}

			log.Error("error when getting payment pointer via HTTP", zap.Error(err), zap.String("url", req.URL.String()))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		ppBytes, err := json.Marshal(pp)
		if err != nil {
			log.Error("error marshalling payment pointer", zap.Error(err), zap.String("url", req.URL.String()))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		_, err = w.Write(ppBytes)
		if err != nil {
			log.Error("error writing http response", zap.Error(err), zap.String("url", req.URL.String()))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}
