package jobs

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/basistheory"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
	tabapay_workflows "gitlab.com/fynbos/backend/providers/tabapay/workflows"
	"go.temporal.io/sdk/workflow"
)

func BackfillLinkedCardCurrencyInfo(ctx workflow.Context) error {
	var a *Activity
	var tabapayActivity *tabapay_workflows.Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)

	var cards []basistheory.Card
	err := workflow.ExecuteActivity(ctx, a.ListBasisTheoryCards).Get(ctx, &cards)
	if err != nil {
		logger.Error("Failed to list basis theory cards")
		return err
	}

	var failedCardIds []string
	for _, card := range cards {
		var cardInfo external.QueryCardResponse
		err := workflow.ExecuteActivity(ctx, tabapayActivity.QueryCard, tabapay_workflows.QueryCard{
			WalletID:       card.WalletID,
			CardNumber:     fmt.Sprintf("{{ %s | json: '$.number' }}", card.TokenID),
			ExpirationDate: fmt.Sprintf("{{ %s | json: '$.expiration_year' | to_string }}{{ %s | json: '$.expiration_month' | pad_left: 2,'0' }}", card.TokenID, card.TokenID),
			AVS:            true,
		}).Get(ctx, &cardInfo)
		if err != nil {
			logger.Error("Failed to query card.")
			failedCardIds = append(failedCardIds, card.ID)
			continue
		}

		pullNetwork := cardInfo.Card.Pull.Network
		if strings.EqualFold(strings.TrimSpace(pullNetwork), "mastercard") {
			pullNetwork = "Mastercard"
		}
		pushNetwork := cardInfo.Card.Push.Network
		if strings.EqualFold(strings.TrimSpace(pushNetwork), "mastercard") {
			pushNetwork = "Mastercard"
		}
		err = workflow.ExecuteActivity(ctx, a.UpdateLinkedAccountCurrencyData, UpdateLinkedAccountCurrencyData{
			Mask:                cardInfo.Card.Last4,
			CanSend:             cardInfo.Card.Pull.Enabled,
			SendCountry:         country.ParseCountry(cardInfo.Card.Pull.Country),
			SendNetwork:         pullNetwork,
			SendCurrency:        currency.ParseCurrency(cardInfo.Card.Pull.Currency),
			SendAvailability:    "Immediate",
			CanReceive:          cardInfo.Card.Push.Enabled,
			ReceiveCountry:      country.ParseCountry(cardInfo.Card.Push.Country),
			ReceiveCurrency:     currency.ParseCurrency(cardInfo.Card.Push.Currency),
			ReceiveNetwork:      pushNetwork,
			ReceiveAvailability: linkedaccounts.FundsAvailability(cardInfo.Card.Push.Availability),
		}).Get(ctx, nil)
		if err != nil {
			logger.Error(err.Error())
			failedCardIds = append(failedCardIds, card.ID)
		}
	}
	if len(failedCardIds) > 0 {
		logger.Error("Failed to backfill basis theory cards", failedCardIds)
		return errors.New("Failed to backfill basis theory cards")
	}

	return nil
}

func (a *Activity) ListBasisThoeryCards(ctx context.Context) ([]basistheory.Card, error) {
	return a.b.BasisTheory().ListCards(ctx)
}

type UpdateLinkedAccountCurrencyData struct {
	ID                  string
	Mask                string
	CanSend             bool
	SendCountry         country.Country
	SendCurrency        currency.Currency
	SendAvailability    linkedaccounts.FundsAvailability
	SendNetwork         string
	CanReceive          bool
	ReceiveCountry      country.Country
	ReceiveCurrency     currency.Currency
	ReceiveAvailability linkedaccounts.FundsAvailability
	ReceiveNetwork      string
}

func (a *Activity) UpdateLinkedAccountCurrencyData(ctx context.Context, args UpdateLinkedAccountCurrencyData) error {
	result, err := a.b.DB().ExecContext(
		ctx,
		"UPDATE linked_accounts SET send_country=$1, send_currency=$2, send_network=$3, send_availability=$4, receive_country=$5, receive_currency=$6, receive_network=$7, receive_availability=$8, can_send=$9, can_receive=$10 WHERE provider=$11 AND id=$12 AND mask=$13;",
		args.SendCountry,
		args.SendCurrency,
		args.SendNetwork,
		args.SendAvailability,
		args.ReceiveCountry,
		args.ReceiveCurrency,
		args.ReceiveNetwork,
		args.ReceiveAvailability,
		args.CanSend,
		args.CanReceive,
		tabapay.ProviderName,
		args.ID,
		args.Mask,
	)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows < 1 {
		return errors.New("Failed to backfill card")
	}

	return nil
}
