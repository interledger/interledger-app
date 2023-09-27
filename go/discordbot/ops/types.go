package ops

import (
	"database/sql"
	"time"

	"github.com/bwmarrin/discordgo"
)

type PaymentInteraction struct {
	ID                 string       `db:"id"`
	PaymentID          string       `db:"payment_id"`
	RawInteraction     string       `db:"interaction"`
	NotifiedReceiver   bool         `db:"notified_receiver"`
	NotifiedProcessing bool         `db:"notified_processing"`
	CreatedAt          time.Time    `db:"created_at"`
	ExpiredAt          sql.NullTime `db:"expired_at"`
}

type Discord interface {
	InteractionResponseEdit(interaction *discordgo.Interaction, newresp *discordgo.WebhookEdit, options ...discordgo.RequestOption) (*discordgo.Message, error)
	UserChannelCreate(recipientID string, options ...discordgo.RequestOption) (*discordgo.Channel, error)
	ChannelMessageSendComplex(channelID string, data *discordgo.MessageSend, options ...discordgo.RequestOption) (*discordgo.Message, error)
}
