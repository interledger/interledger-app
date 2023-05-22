package workflows

import (
	"context"
	"encoding/base64"
	"gitlab.com/fynbos/backend/identities"
)

type Activity struct {
	b Backends
}

func NewActivity(b Backends) *Activity {
	return &Activity{
		b: b,
	}
}

func (a *Activity) GetWalletPaymentPointerURL(ctx context.Context, walletID string) (string, error) {
	pp, err := a.b.OpenPayments().GetWalletPaymentPointer(ctx, walletID)
	if err != nil {
		return "", err
	}

	return pp.URL, nil
}

func (a *Activity) PostProofTweet(ctx context.Context, signatureHash []byte, connectionID, paymentPointerURL string) (string, error) {
	base64SigHas := base64.URLEncoding.EncodeToString(signatureHash)

	tweet, err := a.b.Twitter().PostTweet(ctx, connectionID, "I’ve connected my fynbos wallet, to my Twitter identity so I can send and receive payments using this identity. \n\nSee the proof at "+paymentPointerURL+"/claims/"+string(base64SigHas))
	if err != nil {
		return "", err
	}

	return tweet.ID, nil
}

func (a *Activity) SetIdentityPending(ctx context.Context, identityID string) error {
	err := a.b.Identities().UpdateState(ctx, identityID, identities.StatePending, "")
	if err != nil {
		return err
	}

	return nil
}
