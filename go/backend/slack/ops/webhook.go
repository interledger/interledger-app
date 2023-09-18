package ops

import (
	"net/http"
	"strings"

	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

func BotInstallWebhook(b Backends) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		code := r.URL.Query()["code"]
		if len(code) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		oauthToken, err := b.External().CreateBotToken(r.Context(), code[0])
		if err != nil {
			log.Error("slack bot install failed on token exchange", zap.Error(err), zap.String("code", code[0]))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		scopes := strings.Split(oauthToken.Extra("scope").(string), ",")
		botID := oauthToken.Extra("bot_user_id").(string)
		appID := oauthToken.Extra("app_id").(string)
		teamID := oauthToken.Extra("team").(map[string]interface{})["id"].(string)

		query := `
		INSERT INTO slack_bot_installs (access_token, scopes, team_id, bot_id, app_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (team_id) 
		DO UPDATE SET 
			access_token = EXCLUDED.access_token,
			scopes = EXCLUDED.scopes,
		  	team_id = EXCLUDED.team_id,
			bot_id = EXCLUDED.bot_id,
		  	app_id = EXCLUDED.app_id,
			updated_at = NOW();`

		_, err = b.DB().ExecContext(r.Context(), query, oauthToken.AccessToken, scopes, teamID, botID, appID)
		if err != nil {
			log.Error("slack bot install failed on token insert", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, _ = w.Write([]byte("bot successfully installed"))
		w.WriteHeader(http.StatusOK)
	}
}
