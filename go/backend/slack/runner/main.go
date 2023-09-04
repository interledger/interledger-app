package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"
)

func main() {

	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, "https://slack.com") //openid/connect/authorize
	if err != nil {
		panic(err)
	}

	oidcConfig := &oidc.Config{
		ClientID: "2317468772181.5841878200565",
	}
	verifier := provider.Verifier(oidcConfig)

	conf := oauth2.Config{
		ClientID:     "2317468772181.5841878200565",
		ClientSecret: "e0705d863bc2726505cd175b65cc12d9",
		Endpoint:     provider.Endpoint(),
		RedirectURL:  "https://she-homepage-awareness-airlines.trycloudflare.com/slackback",
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	router := chi.NewRouter()
	router.Get("/goto", func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, conf.AuthCodeURL("state", oidc.Nonce("nonce")), http.StatusFound)
	})

	router.Get("/slackback", func(w http.ResponseWriter, r *http.Request) {
		state, err := r.Cookie("state")
		if err != nil {
			log.Println("state not found")
		}
		if state != nil && r.URL.Query().Get("state") != state.Value {
			log.Println("state did not match")
		}

		oauth2Token, err := conf.Exchange(ctx, r.URL.Query().Get("code"))
		if err != nil {
			http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
			return
		}
		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			http.Error(w, "No id_token field in oauth2 token.", http.StatusInternalServerError)
			return
		}
		idToken, err := verifier.Verify(ctx, rawIDToken)
		if err != nil {
			http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		nonce, err := r.Cookie("nonce")
		if err != nil {
			log.Println("nonce not found")
		}
		if nonce != nil && idToken.Nonce != nonce.Value {
			log.Println("nonce did not match")
		}

		oauth2Token.AccessToken = "*REDACTED*"

		resp := struct {
			OAuth2Token   *oauth2.Token
			IDTokenClaims *json.RawMessage // ID Token payload is just JSON.
		}{oauth2Token, new(json.RawMessage)}

		if err := idToken.Claims(&resp.IDTokenClaims); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data, err := json.MarshalIndent(resp, "", "    ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(data)
	})

	log.Printf("listening on http://%s/", "127.0.0.1:8080")
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", router))
}
