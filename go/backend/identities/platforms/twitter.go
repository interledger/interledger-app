package platforms

import (
	"context"
	"crypto"
	"encoding/json"
	"fmt"
	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/backend/twitter"
	"go.temporal.io/sdk/activity"
	"regexp"
	"time"

	twitterscraper "github.com/n0madic/twitter-scraper"
	"gitlab.com/fynbos/backend/identities"
	"go.temporal.io/sdk/workflow"
)

type twitterPlatform struct {
	platform identities.Platform
	b        Backends
}

func newTwitter(b Backends, platform identities.Platform) *twitterPlatform {
	return &twitterPlatform{
		platform: platform,
		b:        b,
	}
}

func (tp *twitterPlatform) VerifyWorkflow() interface{} {
	return TwitterVerifyWorkflow
}

func (tp *twitterPlatform) GenerateSignedClaim(ctx context.Context, args *SignedClaimArgs) (*GeneratedSignedClaim, error) {
	walletKeys, err := tp.b.Keys().List(ctx, args.WalletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	// Get first custodial key
	var signingKey *keys.Key
	for _, k := range walletKeys {
		if k.Type == keys.Custodial {
			signingKey = &k
			break
		}
	}

	if signingKey == nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, "no custodial key found")
	}

	pp, err := tp.b.OpenPayments().GetWalletPaymentPointer(ctx, args.WalletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	claim := identities.Claim{
		Wallet:     pp.URL,
		Type:       "twitter",
		Identifier: args.Identifier,
		Kid:        signingKey.ID,
		Ctime:      time.Now().Unix(),
	}

	jsonClaim, err := json.Marshal(claim)

	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	signature, err := tp.b.Keys().Sign(ctx, signingKey.ID, args.WalletID, jsonClaim)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	signatureHash := crypto.SHA256.New()
	signatureHash.Write(signature)
	hash := signatureHash.Sum(nil)

	return &GeneratedSignedClaim{
		Claim:         claim,
		Signature:     signature,
		SignatureHash: hash,
	}, nil
}

func (tp *twitterPlatform) VerifyInstructions(ctx context.Context, args *VerifyInstructionsArgs) (string, error) {
	user, err := twitterscraper.New().GetProfile(args.Identifier)
	if err != nil {
		return "", fmt.Errorf("error getting twitter profile: %s", err)
	}

	connections, err := tp.b.Twitter().GetWalletConnections(ctx, args.WalletID)
	if err != nil {
		return "", fmt.Errorf("error getting oauth twitter tokens: %s", err)
	}

	var connection *twitter.Connection
	for _, c := range connections {
		if c.UserID == user.UserID {
			connection = &c
			break
		}
	}

	if connection == nil {
		return "", fmt.Errorf("no connection found for user %s", user.UserID)
	}

	return "Successful", nil
}

type TwitterActivityBackends interface {
	Identities() identities.Client
}

type TwitterActivity struct {
	b TwitterActivityBackends
}

func NewTwitterActivity(b TwitterActivityBackends) *TwitterActivity {
	return &TwitterActivity{b: b}
}

func TwitterVerifyWorkflow(ctx workflow.Context, id, proof string) (string, error) {
	var a *TwitterActivity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("VerifyWorkflow for twitter platform started", "id", id, "proof", proof)

	var tweetProof *twitterscraper.Tweet
	err := workflow.ExecuteActivity(ctx, a.FetchTweetProof, proof).Get(ctx, tweetProof)
	if err != nil {
		return "", fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	var identity *identities.Identity
	err = workflow.ExecuteActivity(ctx, a.GetIdentity, id).Get(ctx, identity)
	if err != nil {
		return "", fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	// What to verify?
	// 1. Check the tweet handles matches the one in the identity
	// 2. When lookup the identity via the sig hash and then do a signature verification
	// 3. Check if the user_id in the tweet matches the one in the identity
	err = workflow.ExecuteActivity(ctx, a.VerifyProof, identity, tweetProof).Get(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	// Set the identity as verified and set the proof url
	err = workflow.ExecuteActivity(ctx, a.VerifyTwitter, id, proof).Get(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return "OK", nil
}

func (a *TwitterActivity) FetchTweetProof(ctx context.Context, proofUrl string) (*twitterscraper.Tweet, error) {
	tweetId := extractTweetID(proofUrl)
	if tweetId == "" {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, "couldn't parse proof tweet id")
	}

	scraper := twitterscraper.New()
	tweet, err := scraper.GetTweet(tweetId)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return tweet, nil
}

func (a *TwitterActivity) GetIdentity(ctx context.Context, id string) (*identities.Identity, error) {
	identity, err := a.b.Identities().Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return identity, nil
}

// TODO: harden verification
func (a *TwitterActivity) VerifyProof(ctx context.Context, identity *identities.Identity, tweet *twitterscraper.Tweet) error {
	if identity.State != identities.StatePending {
		return fmt.Errorf("%w %s", identities.ErrInternal, "identity is not pending")
	}
	activity.GetLogger(ctx).Info("Verifying proof", "identity", identity.ID, "tweet", tweet.ID)

	return nil
}

func (a *TwitterActivity) VerifyTwitter(ctx context.Context, id, proof string) error {
	err := a.b.Identities().UpdateState(ctx, id, identities.StateVerified, proof)
	if err != nil {
		return fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return nil
}

// TODO: better parsing
func extractTweetID(url string) string {
	re := regexp.MustCompile(`\/([0-9]+)$`)
	matches := re.FindStringSubmatch(url)
	if len(matches) == 2 {
		return matches[1]
	}
	return ""
}
