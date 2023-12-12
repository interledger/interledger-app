package external

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type PTI struct {
	Users        map[string]CreateUserArgs
	Wallets      map[string]Wallet
	WalletToUser map[string]string
}

func (p *PTI) Routes() chi.Router {
	r := chi.NewRouter()

	r.HandleFunc("/users", CreateUserHandler(p))
	r.HandleFunc("/users/{userID}/wallets", CreateUserWalletHandler(p))
	r.HandleFunc("/users/assessments", CreateStartAssessmentHandler(p))

	return r
}

func CreateUserHandler(p *PTI) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		var args CreateUserArgs
		err = json.Unmarshal(body, &args)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		_, ok := p.Users[args.ID]
		if ok {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}

		p.Users[args.ID] = args

		resp := CreateUserResponse{
			ID:   args.ID,
			Link: fmt.Sprintf("https://pti.com/users/%s", args.ID),
		}

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func CreateUserWalletHandler(p *PTI) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		var args CreateWalletArgs
		err = json.Unmarshal(body, &args)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		_, ok := p.Wallets[args.WalletID]
		if ok {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}

		wallet := Wallet{
			WalletID:       args.WalletID,
			Currency:       args.Currency,
			Reference:      args.Reference,
			CreateDateTime: time.Now().Format(time.RFC3339),
			Balance:        0,
		}
		p.Wallets[args.WalletID] = wallet
		p.WalletToUser[args.WalletID] = args.UserID

		err = json.NewEncoder(w).Encode(wallet)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func CreateStartAssessmentHandler(p *PTI) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		var args CreateUserArgs
		err = json.Unmarshal(body, &args)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		_, ok := p.Users[args.ID]
		if !ok {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		resp := CreateUserResponse{
			ID:   uuid.NewString(),
			Link: fmt.Sprintf("https://pti.com/users/%s", args.ID),
		}

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func NewPTI() *PTI {
	return &PTI{
		Users:        map[string]CreateUserArgs{},
		Wallets:      map[string]Wallet{},
		WalletToUser: map[string]string{},
	}
}
