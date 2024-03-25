package jobs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/providers/astra"
	"gitlab.com/fynbos/backend/providers/astra/ops"
	"gitlab.com/fynbos/backend/wallets"
	"go.temporal.io/sdk/workflow"
)

func CreateAstraBusinessProfile(ctx workflow.Context) error {
	var a *Activity
	var asrtaActivity *ops.Activity
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

	err = workflow.ExecuteActivity(ctx, asrtaActivity.CreateAccount, wallets.AstraBusinessWalletID).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) SaveIntent(ctx context.Context, externalID string) error {
	_, err := a.b.DB().ExecContext(ctx, "INSERT INTO astra_user_intents (intent_id, wallet_id, user_id, status) VALUES ($1, $2, $1, 'verified') ON CONFLICT DO NOTHING", externalID, wallets.AstraBusinessWalletID)

	return err
}

func (a *Activity) CreateExternalBusinessAccount(ctx context.Context) (string, error) {
	return "350868d5-509b-40d2-a61a-02caa4752259", nil
	/*
		ex := external.New(nil)
		return ex.CreateBusinessProfile(ctx, external.CreateBusinessUserReq{
			BusinessInfo: external.BusinessInfo{
				BusinessType:    "llc",
				BusinessName:    "Fynbos Technologies",
				Ein:             "37-2028338",
				DoingBusinessAs: "Fynbos",
				Phone:           "+17178445997",
				Address1:        "30 North Gould Street",
				Address2:        "Suite R",
				City:            "Sheridan",
				PostalCode:      "82801",
				State:           "WY",
				Website:         "https://fynbos.app/",
			},
			BusinessAdmin: external.BusinessAdmin{
				FirstName: "Adrian",
				LastName:  "Hope-Bailie",
				Email:     "adrian@fynbos.dev",
				Address1:  "30 North Gould Street",
				Address2:  "Suite R",
				City:      "Sheridan",
				State:     "WY",
			},
			BusinessController: external.BusinessController{
				FirstName:   "Adrian",
				LastName:    "Hope-Bailie",
				Email:       "adrian@fynbos",
				Title:       "CEO",
				DateOfBirth: "1982-10-02",
				Address1:    "30 North Gould Street",
				Address2:    "Suite R",
				City:        "Sheridan",
				PostalCode:  "82801",
				State:       "WY",
			},
			KybType:   "verified",
			Phone:     "+17178445997",
			FirstName: "Adrian",
			LastName:  "Hope-Bailie",
			Email:     "adrian@fynbos.dev",
			BeneficialOwners: []external.BeneficialOwners{
				{
					FirstName:   "Adrian",
					LastName:    "Hope-Bailie",
					Email:       "adrian@fynbos",
					DateOfBirth: "1982-10-02",
					Address1:    "30 North Gould Street",
					Address2:    "Suite R",
					City:        "Sheridan",
					PostalCode:  "82801",
					State:       "WY",
				},
			},
		})*/
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
