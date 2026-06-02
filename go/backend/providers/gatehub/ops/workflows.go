package ops

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/notify"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/gatehub/external"
	"gitlab.com/fynbos/backend/slack"
	"gitlab.com/fynbos/backend/transactions"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func CreateGatehubUserWorkflow(ctx workflow.Context, walletID string) (string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating gatehub wallet.")

	var externalUserID string
	err := workflow.ExecuteActivity(ctx, a.GetGatehubUser, walletID).Get(ctx, &externalUserID)
	var applicationError *temporal.ApplicationError
	if errors.As(err, &applicationError) && applicationError.Type() == "ErrNotFound" {
		innerErr := workflow.ExecuteActivity(ctx, a.CreateGatehubUser, walletID).Get(ctx, &externalUserID)
		if innerErr != nil {
			return "", innerErr
		}
	} else if err != nil {
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.SaveGatehubUser, walletID, externalUserID).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	var la linkedaccounts.LinkedAccount
	err = workflow.ExecuteActivity(ctx, a.CreateGatehubWalletLinkedAccount, walletID).Get(ctx, &la)
	if err != nil {
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.CreateGatehubBalanceAccount, la.ID).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	return externalUserID, nil
}

func LinkGatehubUserToGatewayWorkflow(ctx workflow.Context, externalUserID string) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Link gatehub user to gateway.")
	err := workflow.ExecuteActivity(ctx, a.LinkGatehubUserToGateway, externalUserID).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func CreateGatehubDeposit(ctx workflow.Context, wh DepositWebhook) (string, error) {
	var a *Activity

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating gatehub deposit.")

	cc := currency.ParseCurrency(wh.Data.Currency)
	if cc != currency.EUR {
		return "", temporal.NewNonRetryableApplicationError("Invalid currency", "ErrInternal", fmt.Errorf("%w invalid currency", gatehub.ErrInternal))
	}

	amountValue, err := StringToScaledUInt(wh.Data.Amount)
	if err != nil {
		return "", temporal.NewNonRetryableApplicationError("Invalid amount", "ErrInternal", fmt.Errorf("%w %s", gatehub.ErrInternal, err))
	}

	//get transaction and fee from it
	var providerFeeValue int64
	args := &FeeFromGhArgs{UserID: wh.UserID, TrxID: wh.Data.TrxID}
	err = workflow.ExecuteActivity(ctx, a.GetFeeFromGatehubTrasaction, args).Get(ctx, &providerFeeValue)
	if err != nil {
		return "", err
	}

	remainder := amountValue - providerFeeValue

	// calculate fee and remaining amount
	providerFee := currency.Amount{
		Value:    providerFeeValue,
		Currency: cc,
	}
	amt := currency.Amount{
		Value:    remainder,
		Currency: cc,
	}
	fullAMount := currency.Amount{
		Value:    amountValue,
		Currency: cc,
	}

	var walletID string
	err = workflow.ExecuteActivity(ctx, a.GetWalletFromGatehubUser, wh.UserID).Get(ctx, &walletID)
	if err != nil {
		return "", err
	}

	var txID string
	err = workflow.ExecuteActivity(ctx, a.CreateGatehubDepositTransaction, wh.Data.TrxID, walletID, fullAMount, providerFee).Get(ctx, &txID)
	if err != nil {
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.FinalizeGatehubDeposit, txID, walletID, amt, providerFee).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	return txID, nil
}

// TODO: remove once old withdrawal workflows are drained from the Temporal queue
func ProcessGatehubWithdrawal(ctx workflow.Context, walletID, transactionID string) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating gatehub withdrawal.")

	err := workflow.ExecuteActivity(ctx, a.ReserveGatehubBalance, transactionID, walletID).Get(ctx, nil)
	if err != nil {
		handleFailedWithdrawal(ctx, a, walletID, transactionID)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.CheckGatehubWithdrawalComplete, walletID, transactionID).Get(ctx, nil)
	if err != nil {
		handleFailedWithdrawal(ctx, a, walletID, transactionID)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.FinalizeGatehubBalance, transactionID, walletID).Get(ctx, nil)
	if err != nil {
		handleFailedWithdrawal(ctx, a, walletID, transactionID)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.UpdateGatehubWithdrawalState, walletID, transactionID, transactions.StateCompleted).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func CompleteGatehubWithdrawalWorkflow(ctx workflow.Context, userID, externalTxID string) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Completing gatehub withdrawal.", "externalTxID", externalTxID)

	var walletID string
	if err := workflow.ExecuteActivity(ctx, a.GetWalletFromGatehubUser, userID).Get(ctx, &walletID); err != nil {
		return err
	}

	var internalTxID string
	err := workflow.ExecuteActivity(ctx, a.GetGatehubWithdrawalIDByForeignID, walletID, externalTxID).Get(ctx, &internalTxID)
	if err != nil {
		return err
	}

	if err = workflow.ExecuteActivity(ctx, a.FinalizeGatehubWithdrawal, internalTxID).Get(ctx, nil); err != nil {
		return err
	}

	return nil
}

func ProcessCardCreationWorkflow(ctx workflow.Context, wh CardCreatedWebhook) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)

	var isCustomerCreated bool
	err := workflow.ExecuteActivity(ctx, a.IsCustomerCreated, wh.UserUUID).Get(ctx, &isCustomerCreated)
	if err != nil {
		logger.Warn("failed to check if gatehub customer is created", "gatehub_user_id", wh.UserUUID)
		return err
	}

	if isCustomerCreated {
		return nil
	}

	logger.Info("Storing GateHub customer and account IDs.")

	var shouldOrderPlastic bool
	err = workflow.ExecuteActivity(ctx, a.StoreCustomerIDs, wh.UserUUID, wh.Data.CustomerID, wh.Data.AccountID).Get(ctx, &shouldOrderPlastic)
	if err != nil {
		logger.Warn("failed to store gatehub user customer and account id", "gatehub_user_id", wh.UserUUID, "customer_id", wh.Data.CustomerID, "account_id", wh.Data.AccountID)
		return err
	}

	if shouldOrderPlastic {
		err = workflow.ExecuteActivity(ctx, a.CreatePlasticForCard, wh.UserUUID, wh.Data.CardID).Get(ctx, nil)
		if err != nil {
			logger.Warn("failed to order plastic for card", "gatehub_user_id", wh.UserUUID, "customer_id", wh.Data.CustomerID, "account_id", wh.Data.AccountID, "card_id", wh.Data.CardID)
			return err
		}
	}

	// Mark that the first card was processed
	err = workflow.ExecuteActivity(ctx, a.MarkFirstCardAsProcessed, wh.UserUUID).Get(ctx, nil)
	if err != nil {
		return err
	}

	var walletID string
	err = workflow.ExecuteActivity(ctx, a.GetWalletFromGatehubUser, wh.UserUUID).Get(ctx, &walletID)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.SendCardReadyEmail, walletID, wh.Data.CardID).Get(ctx, nil)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.Notify, walletID, notify.NotificationTypeCardReady).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

type CreateCardWorkflowArgs struct {
	WalletID           string
	ExternalIDs        gatehub.ExternalIDs
	Currency           string
	NameOnCard         string
	WalletAddress      string
	CardProductCode    string
	DeliveryAddressID  *string
	NewDeliveryAddress *external.CreateCustomerDeliveryAddressArgs
	ShouldOrderPlastic bool
}

func CreateGateHubCardWorkflow(ctx workflow.Context, args CreateCardWorkflowArgs) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)

	var deliveryAddressID string
	if args.ShouldOrderPlastic {
		if args.NewDeliveryAddress != nil {
			// TODO: Ask for idempotency
			err := workflow.ExecuteActivity(ctx, a.CreateNewDeliveryAddress, args.ExternalIDs.UserID, args.ExternalIDs.CustomerID.String, external.CreateCustomerDeliveryAddressArgs{
				Type:        args.NewDeliveryAddress.Type,
				Line1:       args.NewDeliveryAddress.Line1,
				Line2:       args.NewDeliveryAddress.Line2,
				Line3:       args.NewDeliveryAddress.Line3,
				City:        args.NewDeliveryAddress.City,
				CountryCode: args.NewDeliveryAddress.CountryCode,
				ZipCode:     args.NewDeliveryAddress.ZipCode,
				PostOffice:  args.NewDeliveryAddress.PostOffice,
				Reason:      args.NewDeliveryAddress.Reason,
			}).Get(ctx, &deliveryAddressID)

			if err != nil {
				logger.Warn("failed to create new delivery address for user", "gatehub_user_id", args.ExternalIDs.UserID)
				return err
			}
		} else if args.DeliveryAddressID != nil && *args.DeliveryAddressID != gatehub.KycAddressID {
			deliveryAddressID = *args.DeliveryAddressID
		}
	}

	var card external.Card
	err := workflow.ExecuteActivity(ctx, a.CreateCard, args.ExternalIDs.UserID, args.ExternalIDs.AccountID.String, external.OrderCardArgs{
		Currency:          args.Currency,
		DeliveryAddressID: &deliveryAddressID,
		NameOnCard:        args.NameOnCard,
		WalletAddress:     args.WalletAddress,
		Card: external.NewCardArgs{
			ProductCode: args.CardProductCode,
		},
	}).Get(ctx, &card)

	if err != nil {
		return err
	}

	if args.ShouldOrderPlastic {
		err := workflow.ExecuteActivity(ctx, a.CreatePlasticForCard, args.ExternalIDs.UserID, card.ID).Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	err = workflow.ExecuteActivity(ctx, a.SendCardReadyEmail, args.WalletID, card.ID).Get(ctx, nil)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.Notify, args.WalletID, notify.NotificationTypeCardReady).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func handleFailedWithdrawal(ctx workflow.Context, a *Activity, walletID, transactionID string) {
	logger := workflow.GetLogger(ctx)

	err := workflow.ExecuteActivity(ctx, a.UpdateGatehubWithdrawalState, walletID, transactionID, transactions.StateFailed).Get(ctx, nil)
	if err != nil {
		logger.Error("Unable to update transaction state to failed", transactionID)
	}
}

func BackfillAccountWorkflow(ctx workflow.Context, walletID string) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Backfill gatehub account.")

	// The activity will check if the config has SendingUserID set
	var externalID string
	err := workflow.ExecuteActivity(ctx, a.CheckIfBackfillWasDone, walletID).Get(ctx, &externalID)
	if err != nil {
		return err
	}

	if externalID != "" {
		var balance gatehub.Balance
		err = workflow.ExecuteActivity(ctx, a.BackfillPaywiserBalanceAfterKYC, walletID).Get(ctx, &balance)
		if err != nil {
			return err
		}

		err = workflow.ExecuteActivity(ctx, a.MarkBackfillUser, walletID, externalID, balance).Get(ctx, nil)
		if err != nil {
			// if this errors we need to go in manually and update the table will
			return err
		}
	}

	err = workflow.ExecuteActivity(ctx, a.SetKYCApprovedForGatehub, walletID).Get(ctx, nil)
	if err != nil {
		// if this  we need to go in manually and update the table will
		return err
	}
	return nil
}

// CreateCardTransaction - TODO: This should be revisited as soon as possible. Card transactions can have
// different types: purchase, ATM withdrawals, etc... We are trying to only handle
// purchases and ATM withdrawals for now.
func CreateCardTransaction(ctx workflow.Context, wh CardTransactionEventWebhook) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var ct external.CardTransaction
	err := workflow.ExecuteActivity(ctx, a.GetCardTransaction, wh.UserID, wh.Data.TransactionID).Get(ctx, &ct)
	if err != nil {
		return err
	}

	// ONLY FAILED TX
	var isCompletedTx bool

	var card external.Card
	err = workflow.ExecuteActivity(ctx, a.GetCardDetails, wh.UserID, wh.Data.CardID).Get(ctx, &card)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.SaveGatehubCardTransaction, wh.UserID, card.ID, card.MaskedPan, ct).Get(ctx, &card)
	if err != nil {
		return err
	}

	// radu: I do not like this..., but... such is life. We need to handle all TX types in the future
	if ct.Type != external.CardTransactionTypePurchase && ct.Type != external.CardTransactionTypeATMWithdrawal {
		slack.SendToChannel(context.Background(), slack.ChannelNotifyEvents, "wallet-info-bot", fmt.Sprintf("!!! Received card transaction with unsupported type (only handling 0 and 1 at the moment)  :\nCard TX ID: %s\nCard ID: %s\nCard TX Type: %d\nGateHub User ID: %s", ct.TransactionID, card.ID, ct.Type, wh.UserID))
		return nil
	}

	if ct.GHResponseCode == "CRGUI" || ct.GHResponseCode == "SYSEX" || ct.GHResponseCode == "TRXNS" {
		s := "FAILED"
		isCompletedTx = true
		ct.TxStatus = &s
	}

	if ct.TxStatus == nil {
		// Notify GateHub as well if this happens.
		slack.SendToChannel(context.Background(), slack.ChannelNotifyEvents, "wallet-info-bot", fmt.Sprintf("!!! Received card transaction with no tx status:\nCard TX ID: %s\nCard ID: %s\nGateHub User ID: %s", ct.TransactionID, card.ID, wh.UserID))
	}

	var txID string
	err = workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		return uuid.NewString()
	}).Get(&txID)
	if err != nil {
		return nil
	}

	err = workflow.ExecuteActivity(ctx, a.CreateGatehubCardTransaction, wh.UserID, txID, ct).Get(ctx, &txID)
	if err != nil {
		return err
	}

	var walletID string
	err = workflow.ExecuteActivity(ctx, a.GetWalletFromGatehubUser, wh.UserID).Get(ctx, &walletID)
	if err != nil {
		return err
	}

	// This is a hack for now - only for instant failed card transactions
	if isCompletedTx {
		err = workflow.ExecuteActivity(ctx, a.RollbackGatehubCardTransaction, ct.TransactionID, txID).Get(ctx, nil)
		if err != nil {
			return err
		}
	} else {
		err = workflow.ExecuteActivity(ctx, a.ReserveGatehubBalance, txID, walletID).Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	err = workflow.ExecuteActivity(ctx, a.Notify, walletID, notify.NotificationTypeTransaction).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}
