package ops

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/notify"
	"gitlab.com/fynbos/backend/providers/xago"
	"gitlab.com/fynbos/backend/slack"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/log"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

type Activity struct {
	b Backends
}

func NewActivity(b Backends) *Activity {
	return &Activity{b: b}
}

func SetKYCStatusWorkflow(ctx workflow.Context, walletID string, status kyc.Status) error {
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

	err = workflow.ExecuteActivity(ctx, a.UpdateKYCStatus, walletID, status).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to set KYC status", "err", err)
	}

	if oldStatus != kyc.StatusDenied && status == kyc.StatusDenied {
		_ = workflow.ExecuteActivity(ctx, a.SendDeniedEmail, walletID).Get(ctx, nil)
	} else if oldStatus != kyc.StatusInReview && status == kyc.StatusInReview {
		_ = workflow.ExecuteActivity(ctx, a.SendPendingEmail, walletID).Get(ctx, nil)
	} else if oldStatus != kyc.StatusLevel1 && oldStatus != kyc.StatusLevel2 && (status == kyc.StatusLevel1 || status == kyc.StatusLevel2) {
		_ = workflow.ExecuteActivity(ctx, a.SendApprovedEmail, walletID).Get(ctx, nil)

		err = workflow.ExecuteActivity(ctx, a.CreateKYCWallets, walletID).Get(ctx, nil)
		if err != nil {
			return err
		}
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
	if status == kyc.StatusPending && wallet.Country == country.CA {
		err = workflow.ExecuteActivity(ctx, a.KYCWatchForChimoneySuccessfulKYC, walletID).Get(ctx, nil)
		if err != nil {
			return err
		}
	}
	if status == kyc.StatusPending && wallet.Country == country.US {
		err = workflow.ExecuteActivity(ctx, a.CreateKYCWallets, walletID).Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Activity) KYCGetWallet(ctx context.Context, walletID string) (*wallets.Wallet, error) {
	return a.b.Wallets().Get(ctx, walletID)
}

func (a *Activity) KYCWatchForChimoneySuccessfulKYC(ctx context.Context, walletID string) error {
	return a.b.Chimoney().WatchForSuccessfulKYC(ctx, walletID)
}

func (a *Activity) ResetExceededLimits(ctx context.Context, walletID string) error {
	_, err := a.b.Wallets().SetExceededLimits(ctx, walletID, false)
	return err
}

func (a *Activity) CreateKYCWallets(ctx context.Context, walletID string) error {
	w, err := a.b.Wallets().Get(ctx, walletID)
	if err != nil {
		return err
	}

	if w.Country == country.US {
		_, err = a.b.PTI().CreateWallet(ctx, walletID, currency.USD)
		if err != nil {
			return err
		}
	} else if country.EUCountries[w.Country] {
		// nothing to do. Gatehub user should already be created
	} else if w.Country == country.ZA {
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
		slack.SendToChannel(ctx, slack.ChannelNotifyEvents, "fynbot", fmt.Sprintf("KYC approved for wallet. %s/wallet/%s/profile. Country=%s. Manual creation of balance account required.", env.AdminURL(), walletID, w.Country))
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
