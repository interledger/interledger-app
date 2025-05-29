package workflows

import (
	"context"
	"encoding/base64"
	"fmt"

	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/twitter"
	"go.temporal.io/sdk/temporal"
)

type Activity struct {
	b Backends
}

func NewActivity(b Backends) *Activity {
	return &Activity{
		b: b,
	}
}

func (a *Activity) StartVerification(ctx context.Context, identityID, proofURL string) error {
	_, err := a.b.Identities().StartVerification(ctx, identityID, proofURL)

	return err
}

func (a *Activity) PostProofTweet(ctx context.Context, identityID string) (string, error) {
	// Get Identity
	id, err := a.b.Identities().Get(ctx, identityID)
	if err != nil {
		return "", err
	}

	// Get Connections
	connections, err := a.b.Twitter().GetWalletConnections(ctx, id.WalletID)
	if err != nil {
		return "", err
	}

	// find the connection with the matching username from identity
	var connection *twitter.Connection
	for _, c := range connections {
		if c.Username == id.Identifier {
			connection = &c
			break
		}
	}

	// If no connection found, return error
	if connection == nil {
		return "", temporal.NewNonRetryableApplicationError(fmt.Sprintf("no connection found for identity %s", identityID), "ErrNotFound", nil)
	}

	wallet, err := a.b.Wallets().Get(ctx, id.WalletID)
	if err != nil {
		return "", err
	}
	if len(wallet.Addresses) == 0 {
		return "", fmt.Errorf("no wallet address found for identity %s", identityID)
	}

	// Post Tweet
	base64SigHas := base64.URLEncoding.EncodeToString(id.SignatureHash)

	tweetCopy := "Just connected my @Interledgerdev wallet with my @Twitter account! Now it's easier than ever to send and receive payments using @" + connection.Username + ". \n\n#ConnectedWallet " + wallet.AddressString() + "/identities/" + base64SigHas

	tweet, err := a.b.Twitter().PostTweet(ctx, connection.ID, tweetCopy)
	if err != nil {
		return "", err
	}

	proofUrl := fmt.Sprintf("https://twitter.com/%s/status/%s", connection.Username, tweet.ID)

	return proofUrl, nil
}

type UpdateStateArgs struct {
	IdentityID string `json:"identity_id"`
	Proof      string `json:"proof"`
}

func (a *Activity) UpdateIdentityState(ctx context.Context, args UpdateStateArgs) error {
	err := a.b.Identities().UpdateState(ctx, args.IdentityID, identities.StatePending, args.Proof)
	if err != nil {
		return err
	}

	return nil
}
