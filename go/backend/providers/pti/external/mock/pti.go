package mock

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/interledger/interledger-app/go/backend/providers/pti/external"
)

type PTI struct {
	Users        map[string]external.CreateUserArgs `json:"users"`
	Wallets      map[string]external.Wallet         `json:"wallets"`
	WalletToUser map[string]string                  `json:"walletToUser"`
}

func (p *PTI) Routes() chi.Router {
	r := chi.NewRouter()

	r.HandleFunc("/users", CreateUserHandler(p))
	r.HandleFunc("/users/{userID}/wallets", CreateUserWalletHandler(p))
	r.HandleFunc("/users/assessments", CreateStartAssessmentHandler(p))
	r.HandleFunc("/webhooks/pti", CreateProxyWebhooks(p))
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

		var args external.CreateUserArgs
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

		resp := external.CreateUserResponse{
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

		var args external.CreateWalletArgs
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

		wallet := external.Wallet{
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

		var args external.CreateUserArgs
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

		resp := external.CreateUserResponse{
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

func CreateProxyWebhooks(p *PTI) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Host = os.Getenv("ILW_BACKEND_HOST")
		r.URL.Scheme = "http"
		resp, err := http.DefaultClient.Do(r)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(resp.StatusCode)
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
	}
}

func NewPTI() *PTI {
	dataFilePath := os.Getenv("DATA_FILE_PATH")
	if dataFilePath == "" {
		dataFilePath = "pti_mock.json"
	}

	if _, err := os.Stat(dataFilePath); errors.Is(err, os.ErrNotExist) {
		file, fileErr := os.Create(dataFilePath)
		if fileErr != nil {
			log.Fatalf("Failed to create data file %s. %s", dataFilePath, err)
		}
		_ = file.Close()
	}

	rawData, err := os.ReadFile(dataFilePath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		log.Fatalf("Failed to read PTI data file %s. %s", dataFilePath, err)
	}

	mock := PTI{
		Users:        map[string]external.CreateUserArgs{},
		Wallets:      map[string]external.Wallet{},
		WalletToUser: map[string]string{},
	}
	if len(rawData) > 0 {
		err = json.Unmarshal(rawData, &mock)
		if err != nil {
			log.Fatalf("Failed to unmarshal PTI data file %s. %s", dataFilePath, err)
		}
	}

	return &mock
}

func (p *PTI) Save() error {
	dataFilePath := os.Getenv("DATA_FILE_PATH")
	if dataFilePath == "" {
		dataFilePath = "pti_mock.json"
	}

	file, err := os.Create(dataFilePath)
	if err != nil {
		return err
	}

	rawData, err := json.Marshal(p)
	if err != nil {
		return err
	}

	_, err = file.Write(rawData)

	return err
}
