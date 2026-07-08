package aassassetlinks

import (
	"encoding/json"
	"net/http"

	"github.com/interledger/interledger-app/go/log"
	"go.uber.org/zap"
)

type assetLink struct {
	Relation []string `json:"relation"`
	Target   target   `json:"target"`
}

type target struct {
	Namespace              string   `json:"namespace"`
	PackageName            string   `json:"package_name"`
	SHA256CertFingerprints []string `json:"sha256_cert_fingerprints"`
}

func AssetLinksHandler(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		links := []assetLink{
			{
				Relation: []string{
					"delegate_permission/common.handle_all_urls",
					"delegate_permission/common.get_login_creds",
				},
				Target: target{
					Namespace:   "android_app",
					PackageName: cfg.AndroidPackageName,
					SHA256CertFingerprints: []string{
						cfg.AndroidSHA256,
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(links); err != nil {
			log.Warn("failed to encode assetlinks.json", zap.Error(err))
		}
	}
}
