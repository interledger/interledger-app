package ops

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

const fields = "id, payment_id, notified_receiver, notified_processing, interaction, created_at, expired_at"

func CreatePaymentInteraction(ctx context.Context, b Backends, i *discordgo.Interaction, paymentID string) (*PaymentInteraction, error) {
	rawInteraction, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	var pi PaymentInteraction
	err = b.DB().GetContext(ctx, &pi, fmt.Sprintf("INSERT INTO discord_payment_interactions (payment_id, interaction) VALUES ($1, $2) RETURNING %s;", fields), paymentID, rawInteraction)
	if err != nil {
		return nil, err
	}

	return &pi, nil
}

func ExpirePaymentInteraction(ctx context.Context, b Backends, id string) error {
	i, err := GetPaymentInteraction(ctx, b, id)
	if err != nil {
		return err
	}
	if i.ExpiredAt.Valid {
		return nil
	}

	result, err := b.DB().ExecContext(ctx, "UPDATE discord_payment_interactions SET expired_at=now() WHERE id=$1;", i.ID)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows < 1 {
		return errors.New("Failed to insert interaction")
	}

	return nil
}

func GetPaymentInteraction(ctx context.Context, b Backends, id string) (*PaymentInteraction, error) {
	var i PaymentInteraction
	err := b.DB().GetContext(ctx, &i, fmt.Sprintf("SELECT %s FROM discord_payment_interactions WHERE id=$1;", fields), id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("not found")
	}
	if err != nil {
		return nil, err
	}

	return &i, nil
}

func ListPaymentInteractions(ctx context.Context, b Backends, limit uint32) ([]PaymentInteraction, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	var list []PaymentInteraction
	err := b.DB().SelectContext(ctx, &list, fmt.Sprintf("SELECT %s FROM discord_payment_interactions WHERE expired_at IS NULL ORDER BY created_at ASC LIMIT $1;", fields), limit)
	if err != nil {
		return nil, err
	}

	return list, nil
}

var (
	SuccessContent    = "Success"
	SuccessComponents = []discordgo.MessageComponent{}

	FailureContent    = "Your payment failed"
	FailureComponents = []discordgo.MessageComponent{}

	ProcessingContent    = "Your payment is underway"
	ProcessingComponents = []discordgo.MessageComponent{}

	SignupContentTemplate = "%s is trying to pay you"
	SignupComponents      = []discordgo.MessageComponent{
		// 1) Sign up with Fynbos
		// 2) Connect your card
		// 3) Connect your Discord profile
		// 4) Get paid
	}

	ConnectCardContentTemplate = "%s is trying to pay you"
	ConnectCardComponents      = []discordgo.MessageComponent{
		// Connect your card
	}
)

func ProcessInteractions(ctx context.Context, b Backends, is []PaymentInteraction) {
	for _, i := range is {
		// check if we need to respond to sender's interaction
		if i.ExpiredAt.Valid {
			log.Info("Skipping expired interaction")
			continue
		}
		if time.Since(i.CreatedAt) >= 15*time.Minute {
			log.Info("Expiring payment interaction")
			_ = ExpirePaymentInteraction(ctx, b, i.PaymentID)
			continue
		}

		p, err := b.Payments().Lookup(ctx, i.PaymentID)
		if err != nil {
			log.Error("Failed to lookup payment", zap.Error(err))
			continue
		}

		// sanity checks
		if p.Receiver.Type != payments.IdentityTypeDiscord {
			log.Error("Receiver is not a discord identity", zap.String("paymentID", p.ID))
			continue
		}
		if p.Sender.Type != payments.IdentityTypeDiscord {
			log.Error("Sender is not a discord identity", zap.String("paymentID", p.ID))
			continue
		}

		var interaction discordgo.Interaction
		err = interaction.UnmarshalJSON([]byte(i.RawInteraction))
		if err != nil {
			log.Error("Failed to hydrate discord interaction", zap.Error(err))
			continue
		}

		switch p.State {
		case payments.StateCompleted:
			_, err = b.Discord().InteractionResponseEdit(&interaction, &discordgo.WebhookEdit{
				Content:    &SuccessContent,
				Components: &SuccessComponents,
			})
			_ = ExpirePaymentInteraction(ctx, b, i.ID)
		case payments.StateFailed:
			_, err = b.Discord().InteractionResponseEdit(&interaction, &discordgo.WebhookEdit{
				Content:    &FailureContent,
				Components: &FailureComponents,
			})
			_ = ExpirePaymentInteraction(ctx, b, i.ID)
		case payments.StateProcessing:
			if !i.NotifiedProcessing {
				_, err = b.Discord().InteractionResponseEdit(&interaction, &discordgo.WebhookEdit{
					Content:    &ProcessingContent,
					Components: &ProcessingComponents,
				})

				// remember that we've responded
				if err == nil {
					_, err = b.DB().ExecContext(ctx, "UPDATE discord_payment_interactions SET notified_processing=true WHERE id=$1;", i.ID)
				}
			}
		}
		if err != nil {
			log.Error("Failed to process payment interaction", zap.Error(err))
			continue
		}

		// check if we need to DM receiver
		if i.NotifiedReceiver || p.State != payments.StateProcessing {
			continue
		}

		senderTX, err := b.Transactions().GetTransaction(ctx, p.Sender.WalletID, p.SendTransactionID)
		if err != nil {
			log.Error("discord bot failed to sender transaction", zap.Error(err), zap.String("paymentID", p.ID))
			continue
		}

		// If there is a transfer it's the pull, now we should notify the receiver to link their identity,
		// otherwise wait.
		if len(senderTX.Transfers) < 1 {
			continue
		}

		receiverId, err := b.Identities().GetByIdentifier(ctx, p.Receiver.Identifier)
		if errors.Is(err, identities.ErrNotFound) {
			err = SendDM(ctx, b, p.Receiver, i, fmt.Sprintf(SignupContentTemplate, p.Sender.Identifier), SignupComponents)
			if err != nil {
				log.Error("Failed to DM receiver to sign up", zap.String("paymentID", p.ID))
			}
			continue
		}

		_, err = b.LinkedAccounts().GetDefaultReceive(ctx, receiverId.WalletID)
		if errors.Is(err, identities.ErrNotFound) {
			err = SendDM(ctx, b, p.Receiver, i, fmt.Sprintf(ConnectCardContentTemplate, p.Sender.Identifier), ConnectCardComponents)
			if err != nil {
				log.Error("Failed to DM receiver to connect card", zap.String("paymentID", p.ID))
			}
			continue
		}
	}
}

func SendDM(ctx context.Context, b Backends, user payments.Identity, interaction PaymentInteraction, content string, components []discordgo.MessageComponent) error {
	if user.Type != payments.IdentityTypeDiscord {
		return errors.New("Not a discord identity")
	}

	channel, err := b.Discord().UserChannelCreate(user.Identifier)
	if err != nil {
		return err
	}

	_, err = b.Discord().ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
		Content:    content,
		Components: components,
	})
	if err != nil {
		return err
	}

	_, err = b.DB().ExecContext(ctx, "UPDATE discord_payment_interactions SET notified_receiver=$1 WHERE id=$2;", true, interaction.ID)

	return err
}
