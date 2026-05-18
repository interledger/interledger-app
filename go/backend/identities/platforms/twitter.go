package platforms

import (
	"context"
	"crypto"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"time"

	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/backend/twitter"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"

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

	wallet, err := tp.b.Wallets().Get(ctx, args.WalletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	claim := identities.Claim{
		Wallet:     wallet.AddressString(),
		Type:       string(identities.PlatformTwitter),
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

func (tp *twitterPlatform) GenerateImages(ctx context.Context, args *GenerateImagesArgs) error {
	_, err := tp.b.Images().GenerateTwitterIdentity(ctx, args.WalletURL, args.Identifier)
	if err != nil {
		return err
	}

	_, err = tp.b.Images().GenerateTwitterIdentityOG(ctx, args.WalletURL, args.Identifier)
	if err != nil {
		return err
	}

	return nil
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
	Twitter() twitter.Client
}

type TwitterActivity struct {
	b TwitterActivityBackends
}

func NewTwitterActivity(b TwitterActivityBackends) *TwitterActivity {
	return &TwitterActivity{b: b}
}

// proof input doesn't matter because it's being redeclared in the workflow, it's just to make platform interface happy
// this can be changed in the future
func TwitterVerifyWorkflow(ctx workflow.Context, id, proof string) (string, error) {
	var a *TwitterActivity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("VerifyWorkflow for twitter platform started", "id", id, "proof", proof)

	err := workflow.ExecuteActivity(ctx, a.PublishTweetProof, id).Get(ctx, &proof)
	if err != nil {
		return "", fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	var tweetProof TweetProof
	err = workflow.ExecuteActivity(ctx, a.FetchTweetProof, proof).Get(ctx, &tweetProof)
	if err != nil {
		return "", fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	var identity identities.Identity
	err = workflow.ExecuteActivity(ctx, a.GetTwitterIdentity, id).Get(ctx, &identity)
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

type TweetProof struct {
	ClaimURL        string `json:"claim_url"`
	TwitterUsername string `json:"twitter_username"`
	TwitterUserID   string `json:"twitter_user_id"`
}

func (a *TwitterActivity) FetchTweetProof(ctx context.Context, proofUrl string) (TweetProof, error) {
	tweetId := extractTweetID(proofUrl)
	if tweetId == "" {
		return TweetProof{}, fmt.Errorf("%w %s", identities.ErrInternal, "couldn't parse proof tweet id")
	}

	tweet, err := a.b.Twitter().GetTweet(ctx, tweetId)
	if err != nil {
		return TweetProof{}, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	urls := tweet.URLs
	if len(urls) == 0 {
		// Non retryable temporal error
		return TweetProof{}, temporal.NewNonRetryableApplicationError("no urls found in tweet", "NO_URLS_FOUND", nil)
	}

	return TweetProof{
		ClaimURL:        urls[0],
		TwitterUsername: tweet.AuthorUsername,
		TwitterUserID:   tweet.AuthorID,
	}, nil
}

func (a *TwitterActivity) GetTwitterIdentity(ctx context.Context, id string) (identities.Identity, error) {
	identity, err := a.b.Identities().Get(ctx, id)
	if err != nil {
		return identities.Identity{}, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return *identity, nil
}

// TODO: harden verification
// Get against the claim.
// Verify the signature matches the one in the identity
// Verify the tweet username matches the one in the identity
func (a *TwitterActivity) VerifyProof(ctx context.Context, identity identities.Identity, tp TweetProof) error {
	activity.GetLogger(ctx).Info("Verifying proof", "identity", identity.ID, "tweet", tp.ClaimURL)

	parsedUrl, err := url.Parse(tp.ClaimURL)
	if err != nil {
		return fmt.Errorf("%w %s", identities.ErrInternal, err)
	}
	sigHash := path.Base(parsedUrl.Path)
	base64SigHash := base64.URLEncoding.EncodeToString(identity.SignatureHash)

	// TODO: check if the signature key is still exist before matching the signature
	if sigHash != base64SigHash {
		return fmt.Errorf("%w %s", identities.ErrInternal, "proof sighash doesn't match identity sighash")
	}

	// verify the username
	if identity.Identifier != tp.TwitterUsername {
		return fmt.Errorf("%w %s", identities.ErrInternal, "twitter username doesn't match identity username")
	}

	return nil
}

func (a *TwitterActivity) VerifyTwitter(ctx context.Context, id, proof string) error {
	err := a.b.Identities().UpdateState(ctx, id, identities.StateVerified, proof)
	if err != nil {
		return fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return nil
}

func (a *TwitterActivity) PublishTweetProof(ctx context.Context, identityID string) (string, error) {
	tweetUrl, err := a.b.Twitter().PublishTweetProof(ctx, identityID)
	if err != nil {
		return "", fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return tweetUrl, nil
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
