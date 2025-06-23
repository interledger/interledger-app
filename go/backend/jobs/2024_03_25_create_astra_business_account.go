// Todo: Can be removed. We are not going to use Astra and Basis Theory anymore.
package jobs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"gitlab.com/fynbos/env"

	"gitlab.com/fynbos/backend/slack"

	"gitlab.com/fynbos/backend/providers/astra"
	"gitlab.com/fynbos/backend/providers/astra/external"
	"gitlab.com/fynbos/backend/providers/astra/ops"
	"gitlab.com/fynbos/backend/wallets"
	"go.temporal.io/sdk/workflow"
)

func CreateAstraBusinessProfile(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("CreateAstraBusinessProfile workflow started")

	var exists bool
	err := workflow.ExecuteActivity(ctx, a.IntentExists).Get(ctx, &exists)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	var externalID string
	err = workflow.ExecuteActivity(ctx, a.CreateExternalBusinessAccount).Get(ctx, &externalID)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.SaveIntent, externalID).Get(ctx, nil)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.NotifySlack, externalID).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func ExchangeAstraBusinessProfileCode(ctx workflow.Context) error {
	var a *Activity
	var asrtaActivity *ops.Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("CreateAstraBusinessProfile workflow started")

	var code string
	signalChan := workflow.GetSignalChannel(ctx, "temp-astra")
	for {
		signalChan.Receive(ctx, &code)
		if code != "" {
			break
		}
	}

	err := workflow.ExecuteActivity(ctx, a.ExchangeCode, code).Get(ctx, nil)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, asrtaActivity.CreateAccount, wallets.AstraBusinessWalletID).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) ExchangeCode(ctx context.Context, code string) error {
	ex := external.New(nil)
	accessToken, err := ex.CodeExchange(ctx, code)
	if err != nil {
		return err
	}
	_, err = a.b.DB().ExecContext(ctx, "INSERT INTO astra_access_tokens (wallet_id, token, expires_at, refresh_token, refresh_expires_at) VALUES ($1, $2, $3, $4, $5)"+
		" ON CONFLICT (wallet_id) DO UPDATE SET token = excluded.token, expires_at = excluded.expires_at, refresh_token = excluded.refresh_token, refresh_expires_at = excluded.refresh_expires_at",
		wallets.AstraBusinessWalletID, accessToken.AccessToken, time.Now().Add(time.Minute*110), accessToken.RefreshToken, time.Now().Add(time.Hour*24*9))

	return err
}

func (a *Activity) NotifySlack(ctx context.Context, externalID string) error {
	sandbox := "-sandbox"
	if env.IsProd() {
		sandbox = ""
	}

	slack.SendToChannel(ctx, slack.ChannelNotifyEvents, "Astra Business", fmt.Sprintf("Go To {https://app%s.astra.finance/login/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&business=true&business_profile_id=%s} then signal the workflow with the code", sandbox, os.Getenv("ASTRA_CLIENT_ID"), url.QueryEscape(os.Getenv("ASTRA_CODE_EXCHANGE_REDIRECT")), externalID))
	return nil
}

func (a *Activity) SaveIntent(ctx context.Context, externalID string) error {
	_, err := a.b.DB().ExecContext(ctx, "INSERT INTO astra_user_intents (intent_id, wallet_id, user_id, status) VALUES ($1, $2, $1, 'verified') ON CONFLICT DO NOTHING", externalID, wallets.AstraBusinessWalletID)

	return err
}

func (a *Activity) CreateExternalBusinessAccount(ctx context.Context) (string, error) {
	ex := external.New(nil)
	return ex.CreateBusinessProfile(ctx, external.CreateBusinessUserReq{
		BusinessInfo: external.BusinessInfo{
			BusinessType:    "llc",
			BusinessName:    "Fynbos Inc.",
			Ein:             "37-2102250",
			DoingBusinessAs: "Fynbos",
			Phone:           "+13475834006",
			Address1:        "447 Broadway",
			Address2:        "2nd Floor Suite #2233",
			City:            "New York",
			PostalCode:      "10013",
			State:           "NY",
			Website:         "https://fynbos.app/",
		},
		BusinessAdmin: external.BusinessAdmin{
			FirstName:   "Adrian",
			LastName:    "Hope-Bailie",
			Email:       "adrian@fynbos.dev",
			DateOfBirth: "1982-10-02",
			Address1:    "447 Broadway",
			Address2:    "2nd Floor Suite #2233",
			City:        "New York",
			State:       "NY",
		},
		BusinessController: external.BusinessController{
			FirstName:   "Adrian",
			LastName:    "Hope-Bailie",
			Email:       "adrian@fynbos",
			Title:       "CEO",
			DateOfBirth: "1982-10-02",
			Address1:    "447 Broadway",
			Address2:    "2nd Floor Suite #2233",
			City:        "New York",
			PostalCode:  "10013",
			State:       "NY",
		},
		KybType:   "verified",
		Phone:     "+13475834006",
		FirstName: "Adrian",
		LastName:  "Hope-Bailie",
		Email:     "adrian+astra@fynbos.dev",
		BeneficialOwners: []external.BeneficialOwners{
			{
				FirstName:   "Adrian",
				LastName:    "Hope-Bailie",
				Email:       "adrian@fynbos.dev",
				DateOfBirth: "1982-10-02",
				Address1:    "447 Broadway",
				Address2:    "2nd Floor Suite #2233",
				City:        "New York",
				PostalCode:  "10013",
				State:       "NY",
			},
			{
				FirstName:   "Donovan",
				LastName:    "Changfoot",
				Email:       "don@fynbos.dev",
				DateOfBirth: "1992-03-25",
				Address1:    "447 Broadway",
				Address2:    "2nd Floor Suite #2233",
				City:        "New York",
				PostalCode:  "10013",
				State:       "NY",
			},
			{
				FirstName:   "Matthew",
				LastName:    "De Haast",
				Email:       "matt@fynbos.dev",
				DateOfBirth: "1991-08-21",
				Address1:    "447 Broadway",
				Address2:    "2nd Floor Suite #2233",
				City:        "New York",
				PostalCode:  "10013",
				State:       "NY",
			},
			{
				FirstName:   "Cairin",
				LastName:    "Michie",
				Email:       "cairin@fynbos.dev",
				DateOfBirth: "1994-02-18",
				Address1:    "447 Broadway",
				Address2:    "2nd Floor Suite #2233",
				City:        "New York",
				PostalCode:  "10013",
				State:       "NY",
			},
		},
	})
}

func (a *Activity) IntentExists(ctx context.Context) (bool, error) {
	var intentID string
	err := a.b.DB().GetContext(ctx, &intentID, "SELECT intent_id FROM astra_user_intents WHERE wallet_id=$1", wallets.AstraBusinessWalletID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("%w %s", astra.ErrInternal, err)
	}

	return true, nil
}
