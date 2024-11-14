package ops

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwe"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/slack"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

var ptiPublicKeyJwk = `
{
  "e": "AQAB",
  "kid": "861debeb-98ad-4f9a-a144-351e18093ea9",
  "kty": "RSA",
  "n": "1SG6gvMXxVLH_tQn7N7C5eWnYge7fX9uDdoVQdwQgMwbdDY9XLvBsjHIUZVksEt_t6GY42TeNMfbip8okl44-I_6z6cUzS-5xbiBfE_LEJf4NJKb7zftEj30xsyaT5GXLqdM1FuBE5gq2YBxnKvPVxTvNSsN5I6H9TzlCSVbrG-MMPlVzsai-tW0amy1BlPHIoaExk9HYZQXU6bMqozzy9LXXKh1alo4TzIZKeAU7ID5Vscyvfe7z4MNVvAA1v5FEofNAZBasG5gw6Fiolm9vdcB6Y6kxRWLpsifielF-xs_TfAjh2Ff7rT_tAq6C-s2ETa0kj1WFNEevHPZ3-QfKApn1vULfztrat9Q0knstGq5mGJrNPwzF66E1mG-Nf7q5-IRqGvbtiqggZyPvG6VwIwVi-ZQUQ7wZ2sgIUE_-tLxlTuTP2iWn3fjOy9Oban_AnDydpA5mMPG4N1jxPcXsA9x19wWCgeHNe-AVDG87qW-qBSeh08e6y_6kjwOG-AKctzbpVvpWl9pfDOkZGlgK5QO5n72cKvQ53Dmo1CZkGxXWcUDsd8cwQNgOkbJKl73uzd9Wz3gN7_kZtZRgepSlDSeWAfbrUeg5pyKrXdydVaLOIckgmYQhnFksJNxV7RYNAd1iMXmBazygGpo9cK27b-EePoL6KXOOR36oEOjP1s"
}`

// var ptiPublicKeySource = "https://raw.githubusercontent.com/provenancetech/pti-docs/master/utils/pti-prod-public.jwk"

func ParsePTIPublicKey() (jwk.Key, error) {
	keyToParse := os.Getenv("PTI_PUBLIC_KEY_JWK")
	if keyToParse == "" {
		keyToParse = ptiPublicKeyJwk
	}
	return jwk.ParseKey([]byte(keyToParse), jwk.WithPEM(false))
}

func Webhook(b Backends) (http.HandlerFunc, error) {
	clientID := os.Getenv("PTI_CLIENT_ID")
	ptiPublicKey, err := ParsePTIPublicKey()
	if err != nil {
		return nil, err
	}

	var ptiPrivateKey jwk.Key
	if os.Getenv("PTI_JWK") != "" {
		ptiPrivateKey, err = jwk.ParseKey([]byte(os.Getenv("PTI_JWK")))
		if err != nil {
			return nil, err
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error("pti webhook: Failed to read body", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// hack: add header field if it's missing - otherwise jwe library will fail to parse the message
		var result map[string]interface{}
		err = json.Unmarshal(body, &result)
		if err != nil {
			log.Error("pti webhook: Failed to unmarshal payload", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if result["headers"] == nil {
			headers := map[string]string{}
			headers["alg"] = "RSA-OAEP-256"
			headers["enc"] = "A256CBC-HS512"
			result["header"] = headers
		}

		payloadWithHeader, err := json.Marshal(result)
		if err != nil {
			log.Error("pti webhook: Failed to marshal payload with header", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		d, err := jwe.Decrypt(payloadWithHeader, jwe.WithKey(jwa.RSA_OAEP_256, ptiPrivateKey))
		if err != nil {
			log.Error("pti webhook: Failed to decrypt", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		v, err := jws.Verify(d, jws.WithKey(jwa.RS512, ptiPublicKey))
		if err != nil {
			log.Error("pti webhook: Failed to verify", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		var data WebhookData
		if err = json.Unmarshal(v, &data); err != nil {
			log.Error("pti webhook: Failed to unmarshal json", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if data.ClientID != clientID {
			log.Error("pti webhook: webhook does not match our clientID", zap.String("webhook clientID", data.ClientID))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		switch data.ResourceType {
		case "USER":
			err = HandleUserUpdate(r.Context(), b, v)
		case "USER_ASSESSMENT", "KYC":
			err = HandleAssessmentUpdate(r.Context(), b, v)
		case "TRANSACTION_STATUS":
			log.Error("Unhandled pti transaction status webhook", zap.String("externalUserId", data.UserId), zap.String("requestId", data.RequestID))
		case "TRANSACTION_ASSESSMENT":
			log.Error("Unhandled pti transaction assessment webhook", zap.String("externalUserId", data.UserId), zap.String("requestId", data.RequestID))
		default:
			log.Error("Unknown pti webhook type", zap.String("externalUserId", data.UserId), zap.String("resourceType", data.ResourceType), zap.String("requestId", data.RequestID))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}, nil
}

func HandleUserUpdate(ctx context.Context, b Backends, data []byte) error {
	var userData UserWebhook
	if err := json.Unmarshal(data, &userData); err != nil {
		return err
	}

	result, err := b.DB().ExecContext(ctx, "UPDATE pti_users SET status=$1, updated_at=now() WHERE external_id=$2;", userData.Status, userData.UserId)
	if err != nil {
		return fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	if rows, _ := result.RowsAffected(); rows < 1 {
		return fmt.Errorf("%w Failed to update pti user", pti.ErrInternal)
	}

	return nil
}

func HandleAssessmentUpdate(ctx context.Context, b Backends, data []byte) error {
	var assessmentData AssessmentWebhook
	if err := json.Unmarshal(data, &assessmentData); err != nil {
		return err
	}

	result, err := b.DB().ExecContext(ctx, "UPDATE pti_users SET assessment_status=$1, updated_at=now() WHERE external_id=$2;", assessmentData.Status, assessmentData.UserId)
	if err != nil {
		return fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	if rows, _ := result.RowsAffected(); rows < 1 {
		return fmt.Errorf("%w Failed to update pti user assessment", pti.ErrInternal)
	}

	ptiUser, err := GetUserFromExternalID(ctx, b, assessmentData.UserId)
	if err != nil {
		return fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	if "ACCEPTED" == assessmentData.Status {
		err = b.KYC().SetKYCStatus(ctx, ptiUser.WalletID, kyc.StatusLevel2)
		if err != nil {
			return fmt.Errorf("%w %s", pti.ErrInternal, err)
		}
	} else if "REFUSED" == assessmentData.Status {
		err = b.KYC().SetKYCStatus(ctx, ptiUser.WalletID, kyc.StatusDenied)
		if err != nil {
			return fmt.Errorf("%w %s", pti.ErrInternal, err)
		}
	} else {
		slack.SendToChannel(ctx, slack.ChannelNotifyEvents, "Fynbot", fmt.Sprintf("fiant webhook: kyc assessment status=%s walletID=%s", assessmentData.Status, ptiUser.WalletID))
	}

	return nil
}

type WebhookData struct {
	ResourceType string `json:"resourceType"`
	ClientID     string `json:"clientId"`
	RequestID    string `json:"requestId"`
	UserId       string `json:"userId"`
}

type UserWebhook struct {
	ResourceType string `json:"resourceType"`
	ClientID     string `json:"clientId"`
	RequestID    string `json:"requestId"`
	UserId       string `json:"userId"`
	Status       string `json:"status"`
	StatusReason string `json:"statusReason"`
}

type AssessmentWebhook struct {
	ResourceType  string `json:"resourceType"`
	ClientID      string `json:"clientId"`
	RequestID     string `json:"requestId"`
	UserId        string `json:"userId"`
	Date          string `json:"date"`
	Assessment    string `json:"assessment"`
	Tier          string `json:"tier"`
	RefusalReason string `json:"refusalReason"`
	Status        string `json:"status"`
}
