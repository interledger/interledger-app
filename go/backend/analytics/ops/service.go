package ops

import (
	segment "github.com/segmentio/analytics-go/v3"
	"gitlab.com/fynbos/backend/analytics"
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
