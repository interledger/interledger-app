package ops

import (
	"context"
	"fmt"
	"time"

	"github.com/interledger/interledger-app/go/backend/country"
	"github.com/interledger/interledger-app/go/backend/currency"
	"github.com/interledger/interledger-app/go/backend/kyc"
	"github.com/interledger/interledger-app/go/backend/notify"
	"github.com/interledger/interledger-app/go/backend/providers/xago"
	"github.com/interledger/interledger-app/go/backend/rafiki"
	"github.com/interledger/interledger-app/go/backend/slack"
	"github.com/interledger/interledger-app/go/backend/wallets"
	"github.com/interledger/interledger-app/go/env"
	"github.com/interledger/interledger-app/go/log"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

type Activity struct {
	b Backends
}

func NewActivity(b Backends) *Activity {
	return &Activity{b: b}
}

type SetKYCStatusWorkflowArgs struct {
	WalletID string
	Status   kyc.Status
}

func SetKYCStatusWorkflow(ctx workflow.Context, args SetKYCStatusWorkflowArgs) error {
	walletID := args.WalletID
	status := args.Status
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)

	var oldStatus kyc.Status
	err := workflow.ExecuteActivity(ctx, a.GetKYCStatus, walletID).Get(ctx, &oldStatus)
	if err != nil {
		logger.Error("failed to get old KYC status", "err", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.UpdateKYCStatus, walletID, status).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to set KYC status", "err", err)
	}

	if oldStatus != kyc.StatusDenied && status == kyc.StatusDenied {
		_ = workflow.ExecuteActivity(ctx, a.SendDeniedEmail, walletID).Get(ctx, nil)
	} else if oldStatus != kyc.StatusInReview && status == kyc.StatusInReview {
		_ = workflow.ExecuteActivity(ctx, a.SendPendingEmail, walletID).Get(ctx, nil)
	} else if oldStatus != kyc.StatusDocumentsRequired && status == kyc.StatusDocumentsRequired {
		_ = workflow.ExecuteActivity(ctx, a.SendKYCDocumentsRequiredEmail, walletID).Get(ctx, nil)
	} else if oldStatus != kyc.StatusLevel1 && oldStatus != kyc.StatusLevel2 && (status == kyc.StatusLevel1 || status == kyc.StatusLevel2) {
		_ = workflow.ExecuteActivity(ctx, a.SendApprovedEmail, walletID).Get(ctx, nil)

		err = workflow.ExecuteActivity(ctx, a.CreateKYCWallets, walletID).Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	err = workflow.ExecuteActivity(ctx, a.UpdateRafikiStatus, walletID, status).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to update rafiki status", "err", err)
	}
	// Reset the KYC over the limit notifications for going to L2
	if status == kyc.StatusLevel2 {
		err = workflow.ExecuteActivity(ctx, a.ResetExceededLimits, walletID).Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	var wallet wallets.Wallet
	err = workflow.ExecuteActivity(ctx, a.KYCGetWallet, walletID).Get(ctx, &wallet)
	if err != nil {
		return err
	}

	if status == kyc.StatusPending && wallet.Country == country.US {
		err = workflow.ExecuteActivity(ctx, a.KYCCreatePTIWallet, walletID).Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Activity) KYCGetWallet(ctx context.Context, walletID string) (*wallets.Wallet, error) {
	return a.b.Wallets().Get(ctx, walletID)
}

func (a *Activity) ResetExceededLimits(ctx context.Context, walletID string) error {
	_, err := a.b.Wallets().SetExceededLimits(ctx, walletID, false)
	return err
}

func (a *Activity) KYCCreatePTIWallet(ctx context.Context, walletID string) error {
	_, err := a.b.PTI().CreateWallet(ctx, walletID, currency.USD)
	return err
}

func (a *Activity) CreateKYCWallets(ctx context.Context, walletID string) error {
	w, err := a.b.Wallets().Get(ctx, walletID)
	if err != nil {
		return err
	}

	// flag(bradu): this business logic might need to be revisited
	if w.Country == country.CA {
		// nothing to do. wallet already created
	} else if w.Country == country.US {
		// nothing to do. wallet already created when state was pending
	} else if country.EUCountries[w.Country] {
		// nothing to do. Gatehub user should already be created
	} else if w.Country == country.ZA {
		// check if its just a kyc update
		subAccount, err := a.b.Xago().LookupSubAccount(ctx, walletID)
		if err != nil {
			log.Error("failed to lookup xago subaccount", zap.Error(err), zap.String("wallet_id", walletID))
		}
		if subAccount != nil && subAccount.AccountID != "" {
			log.Info("ZA wallet already has xago account", zap.String("wallet_id", walletID))
			return a.b.Xago().UpdateInquiryLink(ctx, subAccount.AccountID, walletID)
		}

		c := currency.ZAR
		_, err = a.b.Xago().CreateBalanceAccount(ctx, xago.CreateBalanceAccArgs{
			WalletID: w.ID,
			Nickname: "ZAR Balance",
			Title:    "ZAR Balance",
			Currency: c,
		})
		if err != nil {
			return err
		}
	} else {
		slack.SendToChannel(ctx, slack.ChannelSignupKYC, "wallet-info-bot", fmt.Sprintf("KYC approved for wallet. %s/wallet/%s/profile. Country=%s. Manual creation of balance account required.", env.AdminURL(), walletID, w.Country))
		return nil
	}

	return nil
}

func (a *Activity) SendApprovedEmail(ctx context.Context, walletID string) error {
	a.b.Email().SendApplicationApprovedEmail(ctx, walletID)
	return nil
}

func (a *Activity) SendPendingEmail(ctx context.Context, walletID string) error {
	a.b.Email().SendApplicationPendingEmail(ctx, walletID)

	return nil
}

func (a *Activity) SendDeniedEmail(ctx context.Context, walletID string) error {
	a.b.Email().SendApplicationDeniedEmail(ctx, walletID)

	return nil
}

func (a *Activity) SendKYCDocumentsRequiredEmail(ctx context.Context, walletID string) error {
	a.b.Email().SendKYCDocumentsRequiredEmail(ctx, walletID)

	return nil
}

func (a *Activity) GetKYCStatus(ctx context.Context, walletID string) (kyc.Status, error) {
	return GetKYCStatus(ctx, a.b, walletID)
}

func (a *Activity) UpdateKYCStatus(ctx context.Context, walletID string, status kyc.Status) error {
	_, err := a.b.DB().ExecContext(ctx,
		"INSERT INTO wallet_kyc_status (wallet_id, status) VALUES ($1, $2) ON CONFLICT (wallet_id) DO UPDATE SET status = excluded.status;",
		walletID, status)
	if err != nil {
		return fmt.Errorf("%w %s", kyc.ErrInternal, err)
	}

	err = a.b.Notify().NotifyWallet(ctx, walletID, notify.NotificationTypeKyc)
	if err != nil {
		log.Error("notify error", zap.Error(err), zap.String("type", "kyc"))
	}

	return nil
}

func (a *Activity) UpdateRafikiStatus(ctx context.Context, walletID string, status kyc.Status) error {
	rafikiStatus := false
	if status == kyc.StatusLevel1 || status == kyc.StatusLevel2 || status == kyc.StatusApproved {
		rafikiStatus = true
	}
	var wallet []rafiki.UpdateAddressStatus
	err := a.b.DB().SelectContext(ctx, &wallet, "SELECT rafiki.payment_pointer_id, wallets.name FROM public.rafiki_payment_pointers as rafiki INNER JOIN wallets as wallets ON rafiki.wallet_id = wallets.id where wallets.id =$1", walletID)
	if err != nil {
		return err
	}
	if len(wallet) == 0 {
		log.Info("No rafiki payment pointer found for wallet", zap.String("walletID", walletID))
		return nil
	}
	err = a.b.Rafiki().UpdateWalletAddressStatus(ctx, wallet[0], rafikiStatus)
	if err != nil {
		return err
	}
	return nil

}
