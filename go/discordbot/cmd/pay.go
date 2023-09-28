package cmd

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/discordbot/ops"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"

	fynbos_env "gitlab.com/fynbos/env"
)

var PaySlashCommandSchema = discordgo.ApplicationCommand{
	Name:        "pay",
	Description: "Pay another discord user",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "Receiver of payment",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionNumber,
			Name:        "amount",
			Description: "Amount in USD",
			MaxValue:    1000,
			Required:    true,
		},
	},
}

func PaySlashCommandHandler(ctx context.Context, b Backends, s *discordgo.Session, i *discordgo.InteractionCreate) {
	// respond to interaction so Discord doesn't timeout
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Error("Failed to respond to /pay", zap.Error(err))
		return
	}

	var senderUsername string
	if i.User != nil {
		senderUsername = i.User.Username
	} else if i.Member != nil {
		senderUsername = i.Member.User.Username
	}

	data := i.ApplicationCommandData()
	w, err := b.Wallets().ForContext(ctx)
	if err != nil {
		return
	}

	var amount float64
	var receiverUsername, receiverUserID string
	for _, opt := range data.Options {
		switch opt.Name {
		case "amount":
			amount = opt.FloatValue()
		case "user":
			usr := opt.UserValue(s)
			receiverUsername = usr.Username
			receiverUserID = usr.ID
		}
	}

	sendLA, err := b.LinkedAccounts().GetDefaultSend(ctx, w.ID)
	if err != nil {
		newPaymentActionRequired(s, i, fmt.Sprintf("Add a send enabled linked account to pay %s", receiverUsername), "Add", fmt.Sprintf("%s/connect/card", fynbos_env.GetUrl()))
		return
	}

	receiver, err := b.Identities().GetByIdentifier(ctx, receiverUsername)
	if err != nil {
		if receiver != nil && receiver.WalletID == w.ID {
			newPaymentFailedMessage(s, i, "You cannot create a payment to yourself.")
			return
		}
	}
	p, err := b.Payments().Create(ctx, payments.CreateArgs{
		Sender: payments.Identity{
			Type:       payments.IdentityTypeDiscord,
			Identifier: senderUsername,
		},
		Receiver: payments.Identity{
			Type:       payments.IdentityTypeDiscord,
			Identifier: receiverUsername,
		},
		ReceiverAmount: currency.FromFloat64(amount, currency.USD),
		SenderAmount:   currency.FromFloat64(amount, currency.USD),
		SenderAccount:  sendLA.ID,
		IPAddress:      "41.71.7.104", // TODO: take in IP address when confirming payment
	})
	if err != nil {
		log.Error("Failed to create payment", zap.Error(err))
		newPaymentFailedMessage(s, i, "We failed to process the payment.")
		return
	}

	content := "Your payment requires your authorization"
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Authorize",
					Style:    discordgo.LinkButton,
					Disabled: false,
					URL:      fmt.Sprintf("%s/pay/%s", fynbos_env.GetUrl(), p.ID),
				},
			},
		},
	}
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content:    &content,
		Components: &components,
	})
	if err != nil {
		log.Error("Failed to send authorize follow up message for /pay command", zap.Error(err))
		return
	}

	_, err = ops.CreatePaymentInteraction(ctx, b, i.Interaction, p.ID, receiverUserID, senderUsername)
	if err != nil {
		log.Error("Failed to create payment interaction", zap.String("paymentID", p.ID), zap.Error(err))
		return
	}
}

func newPaymentFailedMessage(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: message,
		Flags:   discordgo.MessageFlagsEphemeral,
	})
	if err != nil {
		log.Error("Failed to respond to with payment failed message", zap.Error(err))
	}
}

func newPaymentActionRequired(s *discordgo.Session, i *discordgo.InteractionCreate, message, label, actionURL string) {
	_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: message,
		Flags:   discordgo.MessageFlagsEphemeral,
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    label,
						Style:    discordgo.LinkButton,
						Disabled: false,
						URL:      actionURL,
					},
				},
			},
		},
	})
	if err != nil {
		log.Error("Failed to respond to with payment action required", zap.Error(err))
	}
}
