package platforms

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/backend/twitter"
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

func (tp *twitterPlatform) NewVerifyCode(ctx context.Context, args *NewVerifyCodeArgs) (string, error) {
	walletKeys, err := tp.b.Keys().List(ctx, args.WalletID)
	if err != nil {
		return "", fmt.Errorf("%w %s", identities.ErrInternal, err)
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
		return "", fmt.Errorf("%w %s", identities.ErrInternal, "no custodial key found")
	}

	claim, err := json.Marshal(&identities.IdentityClaim{
		Wallet:       args.WalletID,
		Identifier:   args.Identifier,
		Type:         "twitter",
		KeyID:        signingKey.ID,
		CreationTime: fmt.Sprint(time.Now().Unix()),
	})

	if err != nil {
		return "", fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	signature, err := tp.b.Keys().Sign(ctx, signingKey.ID, args.WalletID, claim)
	if err != nil {
		return "", fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	encodedSig := base64.RawURLEncoding.EncodeToString(signature)

	return encodedSig, nil
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

	_, err = tp.b.Twitter().PostTweet(ctx, connection.ID, args.Identity.VerificationCode)
	if err != nil {
		return "", fmt.Errorf("error posting tweet: %s", err)
	}

	return "Successful", nil
}

type TwitterActivity struct {
	b Backends
}

func NewTwitterActivity(b Backends) *TwitterActivity {
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

	err = workflow.ExecuteActivity(ctx, a.VerifyProof, identity, tweetProof).Get(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

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
	tweet, err := scraper.GetTweet(proofUrl)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return tweet, nil
}

func (a *TwitterActivity) GetIdentity(ctx context.Context, id string) (*identities.Identity, error) {
	var identity *identities.Identity
	err := a.b.DB().GetContext(ctx, identity, "SELECT * FROM identities WHERE id=$1", id)
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

	if identity.VerificationProof != tweet.PermanentURL {
		return fmt.Errorf("%w %s", identities.ErrInternal, "proof tweet url does not match with verification proof")
	}

	if identity.VerificationCode != tweet.Text {
		return fmt.Errorf("%w %s", identities.ErrInternal, "proof code does not match with tweet text")
	}

	return nil
}

func (a *TwitterActivity) VerifyTwitter(ctx context.Context, id, proof string) error {
	_, err := a.b.DB().ExecContext(ctx, "UPDATE identities SET proof=$1, state=$2, updated_at=now() WHERE id=$3",
		proof, identities.StateVerified, id)
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
