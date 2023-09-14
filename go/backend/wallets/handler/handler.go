package handler

import (
	"errors"
	"fmt"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type Backends interface {
	Wallets() wallets.Client
}

// this handler handles fynbos.me redirects done by openpayments server previously
func NewWalletRedirectHandler(b Backends) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		// check if the hostname is one of the fynbos.me domains
		if req.Host != removeProtocol(env.OpenPaymentsURL()) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		ctx := req.Context()

		ppURL, _, err := extractPaymentPointer(getFullURL(req))
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
			u, err := url.JoinPath(env.GetUrl(), "/me/", removeProtocol(wallet.AddressString()))
			if err != nil {
				log.Error("error generating url", zap.Error(err))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, req, u, http.StatusFound)
			return
		}
	}
}

func extractPaymentPointer(rawURL string) (string, string, error) {
	var res string
	for _, res = range wallets.ReservedURLParts {
		if strings.Contains(rawURL, res) {
			waRaw := rawURL[:strings.LastIndex(rawURL, res)]

			wa, err := wallets.ParseAddress(waRaw)
			if err != nil {
				return "", "", err
			}

			return wa.String(), res, nil
		}
	}

	// No suffix found, return the original sanitized
	wa, err := wallets.ParseAddress(rawURL)
	if err != nil {
		return "", "", err
	}
	return wa.String(), "", err
}

func getFullURL(req *http.Request) string {
	fullUrl := req.URL.String()
	if req.URL.Scheme == "" && req.URL.Host == "" {
		fullUrl = fmt.Sprintf("https://%s", path.Join(req.Host, req.URL.Path))
	}

	if strings.HasPrefix(fullUrl, "http://") {
		fullUrl = strings.Replace(fullUrl, "http://", "https://", 1)
	}

	return fullUrl
}

func removeProtocol(url string) string {
	url = strings.Replace(url, "http://", "", 1)
	return strings.Replace(url, "https://", "", 1)
}
