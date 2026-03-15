package aassassetlinks

import (
	"encoding/json"
	"net/http"

	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

type aasa struct {
	AppLinks       appLinks       `json:"applinks"`
	WebCredentials webCredentials `json:"webcredentials"`
}

type appLinks struct {
	Apps    []string        `json:"apps"`
	Details []appLinkDetail `json:"details"`
}

type appLinkDetail struct {
	AppIDs     []string            `json:"appIDs"`
	Components []map[string]string `json:"components"`
}

type webCredentials struct {
	Apps []string `json:"apps"`
}

func AppSiteAssociationHandler(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		association := aasa{
			AppLinks: appLinks{
				Apps: []string{},
				Details: []appLinkDetail{
					{
						AppIDs: []string{cfg.AppleAppID},
						Components: []map[string]string{
							{"/": "/transactions*"},
						},
					},
				},
			},
			WebCredentials: webCredentials{
				Apps: []string{cfg.AppleAppID},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(association); err != nil {
			log.Warn("failed to encode AASA response", zap.Error(err))
		}
	}
}
