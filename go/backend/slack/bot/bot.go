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
	"sync"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

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
			if err != nil && !errors.Is(err, slack.ErrNotFound) {
				log.Error("failed to lookup to connection for slack bot", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var receiverWallet *wallets.Wallet
			if receiverConnection != nil {
				receiverWallet, err = b.Wallets().Get(r.Context(), receiverConnection.WalletID)
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

				_, err = b.LinkedAccounts().GetDefaultReceive(r.Context(), receiverWallet.ID)
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
			}

			p, err := b.Payments().Create(r.Context(), payments.CreateArgs{
				Sender:         getSenderIdentity(r.Context(), b, senderWallet, senderConnection),
				Receiver:       getReceiverIdentity(r.Context(), b, receiverSlackID, receiverWallet, senderConnection, receiverConnection),
				SenderAmount:   amt,
				SenderAccount:  senderAcc.ID,
				ReceiverAmount: amt,
				Note:           note,
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

			// Start a go routine that will poll every minute on the payment and notify the user if they should link an account
			go pollPaymentUpdates(b, p.ID, s)
			return

		default:
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func getReceiverIdentity(ctx context.Context, b Backends, userID string, w *wallets.Wallet, senderCon, con *slack.Connection) payments.Identity {
	if w == nil || con == nil {
		return payments.Identity{
			Type:       payments.IdentityTypeSlack,
			Identifier: fmt.Sprintf("%s / %s", senderCon.TeamName, getUsername(userID)),
		}
	}
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

func getSenderIdentity(ctx context.Context, b Backends, w *wallets.Wallet, con *slack.Connection) payments.Identity {
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

func getUsername(userID string) string {
	if strings.HasPrefix(userID, "<@") {
		userID = userID[strings.Index(userID, "|")+1 : strings.Index(userID, ">")]
	}
	return userID
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

	// Check if the user has unlinked the connection then don't return it.
	_, err = b.Identities().GetByIdentifier(ctx, res.Identifier())
	if errors.Is(err, identities.ErrNotFound) {
		return nil, fmt.Errorf("%w %s", slack.ErrNotFound, err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", slack.ErrInternal, err)
	}

	return &res, nil
}

func pollPaymentUpdates(b Backends, paymentID string, c ext_slack.SlashCommand) {
	// User has 20  minutes to complete the payment.
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*20)
	defer cancel()

	receiverSlackID := strings.Split(c.Text, " ")[0]
	var sentReceiverPrompt bool

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Minute):
			p, err := b.Payments().Lookup(ctx, paymentID)
			if err != nil {
				log.Error("slack bot failed to lookup payment", zap.Error(err), zap.String("payment_id", paymentID))
				continue
			}

			// Nothing more to do
			if p.State == payments.StateFailed || p.State == payments.StateCompleted {
				return
			}
			// Not ready for us to check anything
			if p.State == payments.StateCreated || p.State == payments.StateConfirmed {
				continue
			}

			senderTX, err := b.Transactions().GetTransaction(ctx, p.Sender.WalletID, p.SendTransactionID)
			if err != nil {
				log.Error("slack bot failed to sender transaction", zap.Error(err), zap.String("payment_id", paymentID))
				continue
			}

			// If there is a transfer it's the pull, now we should notify the receiver to link their identity
			shouldPromt := len(senderTX.Transfers) > 1

			// Check that the userID we got from the command matches any slack connections.
			// This is a sanity check to make sure the name we got from the command matches the actual identity linked.
			// i.e. command might return a user with 'fynbos / barnard' but the identity linked is 'fynbos / Barnard du Toit.
			// Log out these anomalies so we can adjust accordingly.
			con, err := lookupAuth(ctx, b, c.TeamID, receiverSlackID)
			if err != nil && !errors.Is(err, slack.ErrNotFound) {
				log.Error("slack bot failed to lookup receiver connection", zap.Error(err), zap.String("payment_id", paymentID))
				continue
			}
			if errors.Is(err, slack.ErrNotFound) && shouldPromt && !sentReceiverPrompt {
				// Send prompt to user to link an
				sendToUser(ctx, b, c.TeamID, receiverSlackID,
					fmt.Sprintf("%s has sent you a payment of %s. Please create an account, link a bank card and this slack profile to receive your payment on %s",
						c.UserID, p.ReceiverAmount.Format(), env.GetUrl()))
				sentReceiverPrompt = true
				continue
			}

			if p.Receiver.Identifier != con.Identifier() {
				log.Warn("linked slack identity does not match", zap.String("expected", p.Receiver.Identifier), zap.String("got", con.Identifier()))
				// TODO: update the payment and signal the workflow
			}

			// Now check if the user has a liked account that can receive
			_, err = b.LinkedAccounts().GetDefaultReceive(ctx, con.WalletID)
			if errors.Is(err, linkedaccounts.ErrNotFound) && !sentReceiverPrompt {
				// Send prompt to user to link an
				sendToUser(ctx, b, c.TeamID, receiverSlackID,
					fmt.Sprintf("%s has sent you a payment of %s. Please create an account, link a bank card to receive your payment on %s",
						c.UserID, p.ReceiverAmount.Format(), env.GetUrl()))
				sentReceiverPrompt = true
			}

			if sentReceiverPrompt {
				return
			}

			return
		}
	}

}

var clients map[string]*ext_slack.Client // TeamID as the key for the slack client used
var lock sync.RWMutex

func init() {
	clients = make(map[string]*ext_slack.Client)
}

func getTeamClient(ctx context.Context, b Backends, teamID string) (*ext_slack.Client, error) {
	lock.RLock()
	api, ok := clients[teamID]
	if ok {
		defer lock.RUnlock()
		return api, nil
	}
	lock.RUnlock()

	// Upgrade to write lock
	lock.Lock()
	defer lock.Unlock()
	// Double check another thread didn't already popuulate
	api, ok = clients[teamID]
	if ok {

		return api, nil
	}

	var accessToken string
	err := b.DB().GetContext(ctx, &accessToken, "SELECT access_token FROM slack_bot_installs WHERE team_id=$1", teamID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", slack.ErrInternal, err)
	}

	cl := ext_slack.New(accessToken, ext_slack.OptionHTTPClient(otelhttp.DefaultClient))
	clients[teamID] = cl

	return cl, nil
}

func sendToUser(ctx context.Context, b Backends, teamID, userID, message string) {
	cl, err := getTeamClient(ctx, b, teamID)
	if err != nil {
		log.Error("slack bot failed to get slack client for team", zap.Error(err), zap.String("team_id", teamID))
		return
	}

	attachment := ext_slack.Attachment{
		Pretext: message,
	}

	_, _, err = cl.PostMessageContext(ctx, normalizeUserID(userID), ext_slack.MsgOptionAttachments(attachment))

	if err != nil {
		log.Error("slack bot failed to send message as itself", zap.Error(err))
	}
}
