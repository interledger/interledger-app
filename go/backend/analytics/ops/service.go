package ops

import (
	segment "github.com/segmentio/analytics-go/v3"
	"gitlab.com/fynbos/backend/analytics"
	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

func Identify(b Backends, args analytics.IdentifyArgs) {
	traits := segment.NewTraits()

	traits.SetEmail(args.Email)
	traits.SetFirstName(args.FirstName)
	traits.SetLastName(args.LastName)

	err := b.Segment().Enqueue(segment.Identify{
		UserId: args.UserId,
		Traits: traits,
	})
	if err != nil {
		log.Error("analytics: error identify user", zap.Error(err))
	}
}

func GroupWallet(b Backends, walletID string, userID string) {
	traits := segment.NewTraits()
	traits.Set("type", "individual")
	err := b.Segment().Enqueue(segment.Group{
		GroupId: walletID,
		UserId:  userID,
		Traits:  traits,
	})
	if err != nil {
		log.Error("analytics: error GroupWallet", zap.Error(err))
	}
}

func TrackUserSignup(b Backends, userID string) {
	err := b.Segment().Enqueue(segment.Track{
		UserId:     userID,
		Event:      "User Signup",
		Properties: segment.NewProperties(),
		Context:    &segment.Context{},
	})
	if err != nil {
		log.Error("analytics: error TrackUserSignup", zap.Error(err))
	}
}

func TrackUserLogin(b Backends, userID string) {
	err := b.Segment().Enqueue(segment.Track{
		UserId:     userID,
		Event:      "User Login",
		Properties: segment.NewProperties(),
		Context:    &segment.Context{},
	})
	if err != nil {
		log.Error("analytics: error TrackUserLogin", zap.Error(err))
	}
}

func TrackUserLogout(b Backends, userID string) {
	err := b.Segment().Enqueue(segment.Track{
		UserId:     userID,
		Event:      "User Logout",
		Properties: segment.NewProperties(),
		Context:    &segment.Context{},
	})
	if err != nil {
		log.Error("analytics: error TrackUserLogout", zap.Error(err))
	}
}

func TrackWalletCreated(b Backends, walletID string, userID string) {

	GroupWallet(b, walletID, userID)

	err := b.Segment().Enqueue(segment.Track{
		Event:      "Wallet Created",
		UserId:     userID,
		Properties: segment.NewProperties(),
		Context: &segment.Context{
			Extra: map[string]interface{}{
				"groupId": walletID,
			},
		},
	})
	if err != nil {
		log.Error("analytics: error TrackWalletCreated", zap.Error(err))
	}
}

func TrackWalletPaymentPointerCreated(b Backends, walletID string) {
	err := b.Segment().Enqueue(segment.Track{
		Event:      "Payment Pointer Created",
		Properties: segment.NewProperties(),
		Context: &segment.Context{
			Extra: map[string]interface{}{
				"groupId": walletID,
			},
		},
	})
	if err != nil {
		log.Error("analytics: error TrackWalletPaymentPointerCreated", zap.Error(err))
	}
}

func TrackWalletTransactionCreated(b Backends, walletID string, args analytics.WalletTransactionArgs) {

	props := segment.NewProperties()
	props.Set("id", args.ID)
	props.Set("amount", args.Amount.Value)
	props.Set("currency", args.Amount.Currency)
	props.Set("type", args.TrxType)
	props.Set("prover", args.Provider)
	props.Set("state", transactions.StatePending)
	props.Set("walletId", walletID)

	err := b.Segment().Enqueue(segment.Track{
		Event:      "Transaction Created",
		UserId:     args.UserID,
		Properties: props,
		Context: &segment.Context{
			Extra: map[string]interface{}{
				"groupId": walletID,
			},
		},
	})
	if err != nil {
		log.Error("analytics: error TrackWalletTransactionCreated", zap.Error(err))
	}
}

func TrackWalletTransactionFailed(b Backends, walletID string, args analytics.WalletTransactionArgs) {

	props := segment.NewProperties()
	props.Set("id", args.ID)
	props.Set("amount", args.Amount.Value)
	props.Set("currency", args.Amount.Currency)
	props.Set("type", args.TrxType)
	props.Set("prover", args.Provider)
	props.Set("state", transactions.StateFailed)
	props.Set("walletId", walletID)
	props.Set("userId", args.UserID)

	err := b.Segment().Enqueue(segment.Track{
		Event:      "Transaction Failed",
		UserId:     args.UserID,
		Properties: props,
		Context: &segment.Context{
			Extra: map[string]interface{}{
				"groupId": walletID,
			},
		},
	})
	if err != nil {
		log.Error("analytics: error TrackWalletTransactionFailed", zap.Error(err))
	}
}

func TrackWalletTransactionCompleted(b Backends, walletID string, args analytics.WalletTransactionArgs) {

	props := segment.NewProperties()
	props.Set("id", args.ID)
	props.Set("amount", args.Amount.Value)
	props.Set("currency", args.Amount.Currency)
	props.Set("type", args.TrxType)
	props.Set("prover", args.Provider)
	props.Set("state", transactions.StateCompleted)
	props.Set("walletId", walletID)
	props.Set("userId", args.UserID)

	err := b.Segment().Enqueue(segment.Track{
		Event:      "Transaction Completed",
		UserId:     args.UserID,
		Properties: props,
		Context: &segment.Context{
			Extra: map[string]interface{}{
				"groupId": walletID,
			},
		},
	})
	if err != nil {
		log.Error("analytics: error TrackWalletTransactionCompleted", zap.Error(err))
	}
}

func TrackWalletMachnetKYCStatus(b Backends, args analytics.MachnetKYCArgs) {

	event := "Machnet User KYC"
	props := segment.NewProperties()
	switch args.Status {
	case machnet.KYCStatusInProgress:
		event = event + " Started"
		props.Set("status", "in_progress")
	case machnet.KYCStatusSuspended:
		event = event + " Suspended"
		props.Set("status", "suspended")
	case machnet.KYCStatusRetry:
		event = event + " Retry"
		props.Set("status", "retry")
	case machnet.KYCStatusVerified:
		event = event + " Verified"
		props.Set("status", "verified")
	case machnet.KYCStatusReviewPending:
		event = event + " Review Pending"
		props.Set("status", "review_pending")
	}

	err := b.Segment().Enqueue(segment.Track{
		Event:      event,
		UserId:     args.UserID,
		Properties: props,
		Context: &segment.Context{
			Extra: map[string]interface{}{
				"groupId": args.WalletID,
			},
		},
	})
	if err != nil {
		log.Error("analytics: error TrackWalletMachnetKYCStatus", zap.Error(err))
	}
}
