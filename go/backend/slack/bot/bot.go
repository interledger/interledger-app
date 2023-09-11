package bot

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	ext_slack "github.com/slack-go/slack"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/slack"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

func NewSlackCommandHandler(b Backends) http.HandlerFunc {
	signingSecret := os.Getenv("SLACK_SIGNING_SECRET")

	return func(w http.ResponseWriter, r *http.Request) {
		verifier, err := ext_slack.NewSecretsVerifier(r.Header, signingSecret)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		r.Body = io.NopCloser(io.TeeReader(r.Body, &verifier))
		s, err := ext_slack.SlashCommandParse(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err = verifier.Ensure(); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		switch s.Command {
		case "/stagepay", "/fynpay", "/pay": // TODO: pick a better name or a name per env
			team := s.TeamID
			senderSlackID := s.UserID

			var receiverSlackID, amountStr, note string
			parts := strings.Split(s.Text, " ")
			if len(parts) < 2 {
				data, err := json.Marshal(&ext_slack.Msg{Text: "Invalid command format."})
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(data)
				return
			}
			receiverSlackID = parts[0]
			amountStr = parts[1]
			if len(parts) > 2 {
				note = strings.Join(parts[2:], " ")
			}

			amtFloat, err := strconv.ParseFloat(amountStr, 64)
			if err != nil {
				data, err := json.Marshal(&ext_slack.Msg{Text: "Invalid amount format."})
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(data)
				return
			}

			amt := currency.FromFloat64(amtFloat, currency.USD)

			senderConnection, err := lookupAuth(r.Context(), b, team, senderSlackID)
			if errors.Is(err, slack.ErrNotFound) {
				data, err := json.Marshal(&ext_slack.Msg{Text: fmt.Sprintf("You do not have a fynbos wallet, please create one on %s", env.GetUrl())})
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(data)
				return
			}
			if err != nil {
				log.Error("failed to lookup from connection for slack bot", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			senderWallet, err := b.Wallets().Get(r.Context(), senderConnection.WalletID)
			if err != nil {
				log.Error("failed to lookup from wallet for slack bot", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			senderAcc, err := b.LinkedAccounts().GetDefaultSend(r.Context(), senderWallet.ID)
			if errors.Is(err, linkedaccounts.ErrNotFound) {
				data, err := json.Marshal(&ext_slack.Msg{Text: "Please add a account capable of sending payments before continuing"})
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(data)
				return
			}
			if err != nil {
				log.Error("failed to lookup from account for slack bot", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			receiverConnection, err := lookupAuth(r.Context(), b, team, receiverSlackID)
			if errors.Is(err, slack.ErrNotFound) {
				data, err := json.Marshal(&ext_slack.Msg{Text: fmt.Sprintf("%s do not have a fynbos wallet, please have them create one on %s", receiverSlackID, env.GetUrl())})
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(data)
				return
			}
			if err != nil {
				log.Error("failed to lookup to connection for slack bot", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			receiverWallet, err := b.Wallets().Get(r.Context(), receiverConnection.WalletID)
			if err != nil {
				log.Error("failed to lookup from wallet for slack bot", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if senderWallet.ID == receiverWallet.ID {
				data, err := json.Marshal(&ext_slack.Msg{Text: "You cannot send payments to yourself"})
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(data)
				return
			}

			toAcc, err := b.LinkedAccounts().GetDefaultReceive(r.Context(), receiverWallet.ID)
			if errors.Is(err, linkedaccounts.ErrNotFound) {
				data, err := json.Marshal(&ext_slack.Msg{Text: fmt.Sprintf("%s is not capable of receiveing payments, please have them add an account that can receive payments", receiverSlackID)})
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(data)
				return
			}
			if err != nil {
				log.Error("failed to lookup to account for slack bot", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			p, err := b.Payments().Create(r.Context(), payments.CreateArgs{
				Sender:          getIdentity(r.Context(), b, senderWallet, senderConnection),
				Receiver:        getIdentity(r.Context(), b, receiverWallet, receiverConnection),
				SenderAmount:    amt,
				SenderAccount:   senderAcc.ID,
				ReceiverAmount:  amt,
				ReceiverAccount: toAcc.ID,
				Note:            note,
				IPAddress:       "41.71.7.104", // TODO: take in IP address when confirming payment
			})
			if err != nil {
				log.Error("failed to create payment for slack bot", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			authURL := fmt.Sprintf("%s/pay?paymentId=%s&start=2", env.GetUrl(), p.ID)

			data, err := json.Marshal(&ext_slack.Msg{Text: fmt.Sprintf("Your payment to %s requires your authorization %s", receiverSlackID, authURL)})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(data)
			return

		default:
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func getIdentity(ctx context.Context, b Backends, w *wallets.Wallet, con *slack.Connection) payments.Identity {
	resp := payments.Identity{
		Type:       payments.IdentityTypeWalletURL,
		Identifier: w.AddressString(),
	}

	id, err := b.Identities().GetByIdentifier(ctx, con.Identifier())
	// Possibly because the user changed their slack alias since linking their identity.
	if err != nil {
		log.Warn("slackbot: failed to get slack identity from identities service, falling back to wallet URL",
			zap.Error(err), zap.String("wallet_id", w.ID), zap.String("connection_id", con.ID))
		return resp
	}
	if id.Platform != identities.PlatformSlack || id.WalletID != w.ID {
		log.Warn("slackbot: mismatch from connection to identity, falling back to wallet URL",
			zap.Error(err), zap.String("wallet_id", w.ID), zap.String("connection_id", con.ID))
		return resp
	}

	return payments.Identity{
		Type:       payments.IdentityTypeSlack,
		Identifier: con.Identifier(),
	}
}

func normalizeUserID(userID string) string {
	if strings.HasPrefix(userID, "<@") {
		userID = userID[2:strings.Index(userID, "|")]
	}
	return userID
}

func lookupAuth(ctx context.Context, b Backends, teamID, userID string) (*slack.Connection, error) {
	var res slack.Connection
	err := b.DB().GetContext(ctx, &res, "SELECT * FROM slack_connections WHERE team_id=$1 and user_id=$2", teamID, normalizeUserID(userID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w %s", slack.ErrNotFound, err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", slack.ErrInternal, err)
	}

	return &res, nil
}
