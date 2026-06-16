package ops

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/notify"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/gatehub/external"
	"gitlab.com/fynbos/backend/slack"
	"gitlab.com/fynbos/backend/transactions"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

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

func CreateCardTransaction(ctx workflow.Context, wh CardTransactionEventWebhook) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	notifyCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 1},
	})

	var ct external.CardTransaction
	err := workflow.ExecuteActivity(ctx, a.GetCardTransaction, wh.UserID, wh.Data.TransactionID).Get(ctx, &ct)
	if err != nil {
		return err
	}

	var card external.Card
	err = workflow.ExecuteActivity(ctx, a.GetCardDetails, wh.UserID, wh.Data.CardID).Get(ctx, &card)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.SaveGatehubCardTransaction, wh.UserID, card.ID, card.MaskedPan, ct).Get(ctx, nil)
	if err != nil {
		return err
	}

	if ct.GHResponseCode == external.CardTransactionGHResponseCodeTRXNS || ct.GHResponseCode == external.CardTransactionGHResponseCodeSYSEX {
		_ = workflow.ExecuteActivity(notifyCtx, slack.SendToChannelActivity, slack.ChannelError, "wallet-info-bot", fmt.Sprintf("!!! Received card transaction with unsupported GateHub response code:\nCard TX ID: %s\nCard ID: %s\nGateHub User ID: %s\nGH ResponseCode: %s", ct.TransactionID, card.ID, wh.UserID, ct.GHResponseCode)).Get(notifyCtx, nil)
		return temporal.NewNonRetryableApplicationError("Unsupported GateHub response code", "ErrInternal", fmt.Errorf("%w unsupported GHResponseCode", gatehub.ErrInternal))
	}

	var txID string
	err = workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		return uuid.NewString()
	}).Get(&txID)
	if err != nil {
		return nil
	}

	var ctMeta CardTransactionMeta
	err = workflow.ExecuteActivity(ctx, a.ComputeCardTransactionMeta, wh.UserID, ct).Get(ctx, &ctMeta)
	if err != nil {
		return err
	}

	var fx *RecordGatehubCardFXData
	if ct.IsTrxAmountConverted {
		if fx, err = buildFX(ct.MastercardConversion, ct.TransactionCurrency, ct.TransactionAmount); err != nil {
			// send slack notification but continue the flow
			_ = workflow.ExecuteActivity(notifyCtx, slack.SendToChannelActivity, slack.ChannelError, "wallet-info-bot", fmt.Sprintf("!!! Missing or invalid Mastercard conversion data for card transaction with FX:\nCard TX ID: %s\nCard ID: %s\nGateHub User ID: %s\nError: %s", ct.TransactionID, card.ID, wh.UserID, err)).Get(notifyCtx, nil)
		}
	}

	switch ct.Operation {
	case external.CardTransactionOperationWithdraw:
		var classification string
		if ct.TransactionClassification != nil {
			classification = *ct.TransactionClassification
		}

		switch classification {
		case external.CardTransactionClassificationAuthorization:
			recordWithdrawalArgs := RecordGatehubCardWithdrawalArgs{
				RecordGatehubCardTxData: RecordGatehubCardTxData{
					WalletID:        ctMeta.WalletID,
					WalletAddress:   ctMeta.WalletAddress,
					LinkedAccountID: ctMeta.LinkedAccountID,
					MerchantName:    ctMeta.MerchantName,
					Note:            getNoteForWithdrawals(ct.Type),
					BillingAmount:   ctMeta.BillingAmount,
				},
			}
			if fx != nil {
				recordWithdrawalArgs.RecordGatehubCardFXData = *fx
			}
			switch ct.Type {
			case external.CardTransactionTypePurchase,
				external.CardTransactionTypeATMWithdrawal,
				external.CardTransactionTypeCashAdvance,
				external.CardTransactionTypePreauthorization:
				if err = workflow.ExecuteActivity(ctx, a.RecordGatehubCardWithdrawal, txID, ct, recordWithdrawalArgs).Get(ctx, nil); err != nil {
					return err
				}
			case external.CardTransactionTypeTransferFromAccount:
				if err = workflow.ExecuteActivity(ctx, a.RecordGatehubCardWithdrawal, txID, ct, recordWithdrawalArgs).Get(ctx, nil); err != nil {
					return err
				}
			case external.CardTransactionTypePreauthorizationIncremental:
				if ct.RefTransactionID != nil {
					var prevInternalTx *transactions.Transaction
					if err = workflow.ExecuteActivity(ctx, a.GetGateHubTransactionByForeignID, ctMeta.WalletID, *ct.RefTransactionID).Get(ctx, &prevInternalTx); err != nil {
						return err
					}
					if err = workflow.ExecuteActivity(ctx, a.RollbackGatehubCardTransaction, *ct.RefTransactionID, prevInternalTx.ID).Get(ctx, nil); err != nil {
						return err
					}
				}
				if err = workflow.ExecuteActivity(ctx, a.RecordGatehubCardWithdrawal, txID, ct, recordWithdrawalArgs).Get(ctx, nil); err != nil {
					return err
				}
			case external.CardTransactionTypePreauthorizationCompletion:
				// Send slack notification for now, see issue wal-872
				_ = workflow.ExecuteActivity(notifyCtx, slack.SendToChannelActivity, slack.ChannelError, "wallet-info-bot", fmt.Sprintf("!!! Received card transaction with unsupported type:\nCard TX ID: %s\nCard ID: %s\nGateHub User ID: %s\nOperation: %d\nClassification: %s\nType: %d", ct.TransactionID, card.ID, wh.UserID, ct.Operation, classification, ct.Type)).Get(notifyCtx, nil)
				return temporal.NewNonRetryableApplicationError("Unsupported type", "ErrInternal", fmt.Errorf("%w unsupported Type", gatehub.ErrInternal))
			default:
				_ = workflow.ExecuteActivity(notifyCtx, slack.SendToChannelActivity, slack.ChannelError, "wallet-info-bot", fmt.Sprintf("!!! Received card transaction with unsupported type:\nCard TX ID: %s\nCard ID: %s\nGateHub User ID: %s\nOperation: %d\nClassification: %s\nType: %d", ct.TransactionID, card.ID, wh.UserID, ct.Operation, classification, ct.Type)).Get(notifyCtx, nil)
				return temporal.NewNonRetryableApplicationError("Unsupported type", "ErrInternal", fmt.Errorf("%w unsupported Type", gatehub.ErrInternal))
			}
		case external.CardTransactionClassificationReversal:
			_ = workflow.ExecuteActivity(notifyCtx, slack.SendToChannelActivity, slack.ChannelError, "wallet-info-bot", fmt.Sprintf("!!! Received card transaction with unsupported classification:\nCard TX ID: %s\nCard ID: %s\nGateHub User ID: %s\nOperation: %d\nTransaction Classification: %s", ct.TransactionID, card.ID, wh.UserID, ct.Operation, classification)).Get(notifyCtx, nil)
			return temporal.NewNonRetryableApplicationError("Unsupported transaction classification", "ErrInternal", fmt.Errorf("%w unsupported TransactionClassification", gatehub.ErrInternal))
		default:
			_ = workflow.ExecuteActivity(notifyCtx, slack.SendToChannelActivity, slack.ChannelError, "wallet-info-bot", fmt.Sprintf("!!! Received card transaction with unsupported classification:\nCard TX ID: %s\nCard ID: %s\nGateHub User ID: %s\nOperation: %d\nTransaction Classification: %s", ct.TransactionID, card.ID, wh.UserID, ct.Operation, classification)).Get(notifyCtx, nil)
			return temporal.NewNonRetryableApplicationError("Unsupported transaction classification", "ErrInternal", fmt.Errorf("%w unsupported TransactionClassification", gatehub.ErrInternal))
		}

		if fx != nil {
			date := ct.CreatedAt
			if ct.TransactionDateTime != nil {
				date = *ct.TransactionDateTime
			}
			if err = workflow.ExecuteActivity(ctx, a.SendCardTransactionFXEmail, SendCardTransactionFXEmailArgs{
				WalletID:          ctMeta.WalletID,
				MaskedPAN:         card.MaskedPan,
				MerchantName:      ctMeta.MerchantName,
				Date:              date,
				Surcharge:         fx.ExchangeRateSurcharge,
				TransactionAmount: *fx.TargetAmount,
				BillingAmount:     ctMeta.BillingAmount,
			}).Get(ctx, nil); err != nil {
				return err
			}
		}
	case external.CardTransactionOperationDeposit:
		var classification string
		if ct.TransactionClassification != nil {
			classification = *ct.TransactionClassification
		}

		recordDepositArgs := RecordGatehubCardDepositArgs{
			RecordGatehubCardTxData: RecordGatehubCardTxData{
				WalletID:        ctMeta.WalletID,
				WalletAddress:   ctMeta.WalletAddress,
				LinkedAccountID: ctMeta.LinkedAccountID,
				MerchantName:    ctMeta.MerchantName,
				Note:            getNoteForDeposits(ct.Type, classification),
				BillingAmount:   ctMeta.BillingAmount,
			},
		}
		switch classification {
		case external.CardTransactionClassificationAuthorization:
			switch ct.Type {
			case external.CardTransactionTypeTransferToAccount:
				if err = workflow.ExecuteActivity(ctx, a.RecordGatehubCardDeposit, txID, ct, recordDepositArgs).Get(ctx, nil); err != nil {
					return err
				}
			default:
				_ = workflow.ExecuteActivity(notifyCtx, slack.SendToChannelActivity, slack.ChannelError, "wallet-info-bot", fmt.Sprintf("!!! Received card transaction with unsupported type:\nCard TX ID: %s\nCard ID: %s\nGateHub User ID: %s\nOperation: %d\nClassification: %s\nType: %d", ct.TransactionID, card.ID, wh.UserID, ct.Operation, classification, ct.Type)).Get(notifyCtx, nil)
				return temporal.NewNonRetryableApplicationError("Unsupported type", "ErrInternal", fmt.Errorf("%w unsupported Type", gatehub.ErrInternal))
			}
		case external.CardTransactionClassificationReversal:
			switch ct.Type {
			case external.CardTransactionTypePurchase,
				external.CardTransactionTypeATMWithdrawal,
				external.CardTransactionTypeCashAdvance,
				external.CardTransactionTypePreauthorization,
				external.CardTransactionTypePreauthorizationIncremental,
				external.CardTransactionTypePreauthorizationCompletion:
				if ct.RefTransactionID != nil {
					var prevInternalTx *transactions.Transaction
					if err = workflow.ExecuteActivity(ctx, a.GetGateHubTransactionByForeignID, ctMeta.WalletID, *ct.RefTransactionID).Get(ctx, &prevInternalTx); err != nil {
						return err
					}
					if err = workflow.ExecuteActivity(ctx, a.FailGatehubCardTransaction, *ct.RefTransactionID, prevInternalTx.ID).Get(ctx, nil); err != nil {
						return err
					}
				} else {
					_ = workflow.ExecuteActivity(notifyCtx, slack.SendToChannelActivity, slack.ChannelError, "wallet-info-bot", fmt.Sprintf("!!! Received reversal for card transaction without a ref transaction id:\nCard TX ID: %s\nCard ID: %s\nGateHub User ID: %s\nType: %d", ct.TransactionID, card.ID, wh.UserID, ct.Type)).Get(notifyCtx, nil)
				}
				if err = workflow.ExecuteActivity(ctx, a.CreateGatehubCardReversal, txID, ct, recordDepositArgs).Get(ctx, nil); err != nil {
					return err
				}
			default:
				_ = workflow.ExecuteActivity(notifyCtx, slack.SendToChannelActivity, slack.ChannelError, "wallet-info-bot", fmt.Sprintf("!!! Received card transaction with unsupported type:\nCard TX ID: %s\nCard ID: %s\nGateHub User ID: %s\nOperation: %d\nClassification: %s\nType: %d", ct.TransactionID, card.ID, wh.UserID, ct.Operation, classification, ct.Type)).Get(notifyCtx, nil)
				return temporal.NewNonRetryableApplicationError("Unsupported type", "ErrInternal", fmt.Errorf("%w unsupported Type", gatehub.ErrInternal))
			}
		default:
			_ = workflow.ExecuteActivity(notifyCtx, slack.SendToChannelActivity, slack.ChannelError, "wallet-info-bot", fmt.Sprintf("!!! Received card transaction with unsupported classification:\nCard TX ID: %s\nCard ID: %s\nGateHub User ID: %s\nOperation: %d\nTransaction Classification: %s", ct.TransactionID, card.ID, wh.UserID, ct.Operation, classification)).Get(notifyCtx, nil)
			return temporal.NewNonRetryableApplicationError("Unsupported transaction classification", "ErrInternal", fmt.Errorf("%w unsupported TransactionClassification", gatehub.ErrInternal))
		}
	case external.CardTransactionOperationNone:
		var responseCode string
		if ct.ResponseCode != nil {
			responseCode = *ct.ResponseCode
		}

		state := transactions.StateFailed
		if ct.GHResponseCode == external.CardTransactionGHResponseCodeOK && responseCode == external.CardTransactionResponseCodeOK {
			state = transactions.StateCompleted
		}

		recordInformationalArgs := RecordGatehubCardInformationalArgs{
			RecordGatehubCardTxData: RecordGatehubCardTxData{
				WalletID:        ctMeta.WalletID,
				WalletAddress:   ctMeta.WalletAddress,
				LinkedAccountID: ctMeta.LinkedAccountID,
				MerchantName:    ctMeta.MerchantName,
				Note:            getNoteForInformationals(ct.Type, ct.GHResponseCode, responseCode),
				BillingAmount:   ctMeta.BillingAmount,
			},
			State: state,
		}
		if fx != nil {
			recordInformationalArgs.RecordGatehubCardFXData = *fx
		}

		if err = workflow.ExecuteActivity(ctx, a.RecordGatehubCardInformational, txID, ct, recordInformationalArgs).Get(ctx, nil); err != nil {
			return err
		}

		if fx != nil {
			date := ct.CreatedAt
			if ct.TransactionDateTime != nil {
				date = *ct.TransactionDateTime
			}
			if err = workflow.ExecuteActivity(ctx, a.SendCardTransactionFXEmail, SendCardTransactionFXEmailArgs{
				WalletID:          ctMeta.WalletID,
				MaskedPAN:         card.MaskedPan,
				MerchantName:      ctMeta.MerchantName,
				Date:              date,
				Surcharge:         fx.ExchangeRateSurcharge,
				TransactionAmount: *fx.TargetAmount,
				BillingAmount:     ctMeta.BillingAmount,
			}).Get(ctx, nil); err != nil {
				return err
			}
		}
	default:
		_ = workflow.ExecuteActivity(notifyCtx, slack.SendToChannelActivity, slack.ChannelError, "wallet-info-bot", fmt.Sprintf("!!! Received card transaction with unsupported operation:\nCard TX ID: %s\nCard ID: %s\nGateHub User ID: %s\nOperation: %d", ct.TransactionID, card.ID, wh.UserID, ct.Operation)).Get(notifyCtx, nil)
		return temporal.NewNonRetryableApplicationError("Unsupported operation", "ErrInternal", fmt.Errorf("%w unsupported Operation", gatehub.ErrInternal))
	}

	err = workflow.ExecuteActivity(ctx, a.Notify, ctMeta.WalletID, notify.NotificationTypeTransaction).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}
