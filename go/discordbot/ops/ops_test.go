package ops_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/payments/client/mock"
	"gitlab.com/fynbos/discordbot/ops"
)

func TestProcessInteractions(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	b := &backends{
		db:      db.MigrateTestDB(t, ctx),
		pc:      mock.NewMockClient(ctrl),
		discord: &MockDiscord{},
	}

	cases := []struct {
		Name                       string
		PaymentID                  string
		PaymentState               payments.State
		PaymentRequiredActions     []payments.RequiredActionType
		AlreadyNotifiedProcessing  bool
		ExpectResponse             bool
		ExpectedResponseContent    string
		ExpectedResponseComponents []discordgo.MessageComponent
		ExpectDM                   bool
		ExpectedDMContent          string
		ExpectedDMComponents       []discordgo.MessageComponent
	}{
		{
			Name:                      "Notifies receiver when payment requires receiver identity",
			PaymentID:                 uuid.NewString(),
			PaymentState:              payments.StateProcessing,
			PaymentRequiredActions:    []payments.RequiredActionType{payments.RequiredActionTypeReceiverIdentifier},
			AlreadyNotifiedProcessing: true,
			ExpectDM:                  true,
			ExpectedDMContent:         "fynbos is trying to pay you",
			ExpectedDMComponents:      ops.SignupComponents,
		},
		{
			Name:                      "Notifies receiver when payment requires receiver account",
			PaymentID:                 uuid.NewString(),
			PaymentState:              payments.StateProcessing,
			PaymentRequiredActions:    []payments.RequiredActionType{payments.RequiredActionTypeReceiverAccount},
			AlreadyNotifiedProcessing: true,
			ExpectDM:                  true,
			ExpectedDMContent:         "fynbos is trying to pay you",
			ExpectedDMComponents:      ops.ConnectCardComponents,
		},
		{
			Name:                       "Responds success when payment is completed",
			PaymentID:                  uuid.NewString(),
			PaymentState:               payments.StateCompleted,
			ExpectResponse:             true,
			ExpectedResponseContent:    ops.SuccessContent,
			ExpectedResponseComponents: ops.SuccessComponents,
		},
		{
			Name:                       "Responds failed when payment has failed",
			PaymentID:                  uuid.NewString(),
			PaymentState:               payments.StateFailed,
			ExpectResponse:             true,
			ExpectedResponseContent:    ops.FailureContent,
			ExpectedResponseComponents: ops.FailureComponents,
		},
		{
			Name:                       "Responds processing when payment has started processing",
			PaymentID:                  uuid.NewString(),
			PaymentState:               payments.StateProcessing,
			ExpectResponse:             true,
			ExpectedResponseContent:    ops.ProcessingContent,
			ExpectedResponseComponents: ops.ProcessingComponents,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(st *testing.T) {
			b.discord.Responses = []*discordgo.WebhookEdit{}
			b.discord.UserDMs = []*discordgo.MessageSend{}
			b.pc.EXPECT().Lookup(gomock.Any(), tc.PaymentID).Return(&payments.Payment{
				ID:    tc.PaymentID,
				State: tc.PaymentState,
				Sender: payments.Identity{
					Type:       payments.IdentityTypeDiscord,
					Identifier: "fynbos",
				},
				Receiver: payments.Identity{
					Type:       payments.IdentityTypeDiscord,
					Identifier: "jimmy",
				},
				RequiredActions: tc.PaymentRequiredActions,
			}, nil).Times(1)
			applicationInteraction := &discordgo.Interaction{
				ID:   uuid.NewString(),
				Type: discordgo.InteractionApplicationCommand,
			}
			pi, err := CreatePaymentInteraction(st, b, applicationInteraction, tc.PaymentID, tc.AlreadyNotifiedProcessing)
			require.NoError(st, err)
			assert.Equal(st, tc.PaymentID, pi.PaymentID)

			ops.ProcessInteractions(ctx, b, []ops.PaymentInteraction{*pi})

			pi, err = ops.GetPaymentInteraction(ctx, b, pi.ID)
			require.NoError(st, err)

			if tc.ExpectResponse {
				require.Len(st, b.discord.Responses, 1)
				assert.Equal(st, tc.ExpectedResponseContent, *b.discord.Responses[0].Content)
				assert.Equal(st, *b.discord.Responses[0].Components, tc.ExpectedResponseComponents)
			} else {
				assert.Empty(st, b.discord.Responses)
			}

			if tc.ExpectDM {
				require.Len(st, b.discord.UserDMs, 1)
				assert.Equal(st, tc.ExpectedDMContent, b.discord.UserDMs[0].Content)
				assert.Equal(st, tc.ExpectedDMComponents, b.discord.UserDMs[0].Components)
			} else {
				assert.Empty(st, b.discord.UserDMs)
			}

			if tc.PaymentState == payments.StateCompleted || tc.PaymentState == payments.StateFailed {
				assert.True(st, pi.ExpiredAt.Valid)
			}
		})
	}
}

func CreatePaymentInteraction(t *testing.T, b *backends, i *discordgo.Interaction, paymentID string, alreadyNotifiedProcessing bool) (*ops.PaymentInteraction, error) {
	rawInteraction, err := json.Marshal(i)
	if err != nil {
		t.Fatal(err)
	}

	id := uuid.NewString()
	b.DB().MustExec("INSERT INTO discord_payment_interactions (id, payment_id, notified_processing, interaction) VALUES ($1, $2, $3, $4) RETURNING *;", id, paymentID, alreadyNotifiedProcessing, rawInteraction)

	return ops.GetPaymentInteraction(context.Background(), b, id)
}
