package handler

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/go-chi/chi/v5"
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/backend/rafiki"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

type Backends interface {
	Wallets() wallets.Client
	Identities() identities.Client
	Keys() keys.Client
	Rafiki() rafiki.Client
}

// this handler handles ilp.link redirects done by openpayments server previously
func WalletRedirectHandler(b Backends) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		// check if the hostname is one of the ilp.link domains
		if req.Host != removeProtocol(env.OpenPaymentsURL()) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		ctx := req.Context()

		ppURL, suffix, err := extractWalletAddress(getFullURL(req))
		if errors.Is(err, wallets.ErrInvalidAddress) {
			http.Error(w, "Invalid wallet address", http.StatusBadRequest)
			return
		}

		wallet, err := b.Wallets().GetFromAddress(ctx, ppURL)
		if err != nil {
			if errors.Is(err, wallets.ErrNoWalletFound) {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			if errors.Is(err, wallets.ErrInvalidAddress) {
				http.Error(w, "Invalid wallet address", http.StatusBadRequest)
				return
			}

			log.Error("error when getting wallet address via HTTP", zap.Error(err), zap.String("url", req.URL.String()))
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

		rafikiWalletAddress, err := b.Rafiki().GetWalletAddress(ctx, wallet.ID)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		jsonResponse := JsonResponse{
			Id:                wallet.AddressString(),
			PublicName:        wallet.Name,
			Identities:        jsonIds,
			AssetCode:         rafikiWalletAddress.AssetCode,
			AssetScale:        uint(rafikiWalletAddress.AssetScale),
			AuthServerURL:     env.AuthURL(),
			ResourceServerURL: env.OpenPaymentsURL(),
		}
		if env.IsDev() {
			jsonResponse.Identities = nil
		}

		// Fallback to get payment pointer
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "'GET,HEAD,PUT,POST,DELETE,PATCH'")
		err = json.NewEncoder(w).Encode(jsonResponse)
		if err != nil {
			log.Error("error writing http response", zap.Error(err), zap.String("url", req.URL.String()))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func GetIdentityHandler(b Backends) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		// check if the hostname is one of the ilp.link domains
		if req.Host != removeProtocol(env.OpenPaymentsURL()) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		identitySigHash := chi.URLParam(req, "identity_sig_hash")
		ppURL, _, err := extractWalletAddress(getFullURL(req))
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

func extractWalletAddress(rawURL string) (string, string, error) {
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

func listKeys(b Backends, walletID string, w http.ResponseWriter, req *http.Request) {
	wKeys, err := b.Keys().List(req.Context(), walletID)
	if err != nil {
		log.Error("error listing client public keys", zap.Error(err), zap.String("walletID", walletID))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	jwks := make([]Jwk, len(wKeys))
	for i, k := range wKeys {
		jwks[i] = Jwk{
			Kty: "OKP",
			Kid: k.KeyID,
			Crv: "Ed25519",
			Alg: "EdDSA",
			X:   convertToBase64Url(k.PublicKey),
		}
		if !env.IsDev() {
			jwks[i].Use = "sig"
		}
	}

	resp := struct {
		Keys []Jwk `json:"keys"`
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

func convertToBase64Url(publicKey string) string {
	bufferKey, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return publicKey
	}

	return base64.RawURLEncoding.EncodeToString(bufferKey)
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

type Jwk struct {
	Kty string `json:"kty,omitempty"`
	E   string `json:"e,omitempty"`
	Kid string `json:"kid,omitempty"`
	Alg string `json:"alg,omitempty"`
	N   string `json:"n,omitempty"`
	Crv string `json:"crv,omitempty"`
	X   string `json:"x,omitempty"`
	Use string `json:"use,omitempty"`
}
type JsonResponse struct {
	Id                string             `json:"id"`
	PublicName        string             `json:"publicName"`
	Identities        []IdentityResponse `json:"identities,omitempty"`
	AssetCode         string             `json:"assetCode"`
	AssetScale        uint               `json:"assetScale"`
	AuthServerURL     string             `json:"authServer"`
	ResourceServerURL string             `json:"resourceServer"`
}
