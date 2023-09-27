package ops_test

import (
	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/payments"
	payments_mock "gitlab.com/fynbos/backend/payments/client/mock"
	"gitlab.com/fynbos/discordbot/ops"
)

type backends struct {
	db      *sqlx.DB
	pc      *payments_mock.MockClient
	discord *MockDiscord
}

func (b backends) Discord() ops.Discord {
	return b.discord
}

func (b backends) DB() *sqlx.DB {
	return b.db
}

func (b backends) Payments() payments.Client {
	return b.pc
}

type MockDiscord struct {
	Responses []*discordgo.WebhookEdit
	UserDMs   []*discordgo.MessageSend
}

func (m *MockDiscord) InteractionResponseEdit(interaction *discordgo.Interaction, newresp *discordgo.WebhookEdit, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	m.Responses = append(m.Responses, newresp)
	return nil, nil
}

func (m *MockDiscord) UserChannelCreate(recipientID string, options ...discordgo.RequestOption) (*discordgo.Channel, error) {
	return &discordgo.Channel{
		ID: uuid.NewString(),
	}, nil
}

func (m *MockDiscord) ChannelMessageSendComplex(channelID string, data *discordgo.MessageSend, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	m.UserDMs = append(m.UserDMs, data)
	return nil, nil
}
