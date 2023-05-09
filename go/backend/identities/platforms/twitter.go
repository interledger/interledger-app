package platforms

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	twitterscraper "github.com/n0madic/twitter-scraper"
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/twitter"
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

	signingKey := walletKeys[0]

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

	tokens, err := tp.b.Twitter().GetTokensByUserID(ctx, &twitter.GetTokensByUserIdArgs{
		UserID:   user.UserID,
		WalletID: args.WalletID,
		Scopes:   []string{"offline.access", "tweet.read", "tweet.write", "users.read", "follows.read"},
	})
	if err != nil {
		return "", fmt.Errorf("error getting oauth twitter tokens: %s", err)
	}
	if len(tokens) == 0 {
		auth, err := tp.b.Twitter().CreateAuthURL(ctx, &twitter.CreateAuthURLArgs{
			WalletID: args.WalletID,
			Scopes:   []string{"offline.access", "tweet.read", "tweet.write", "users.read", "follows.read"},
		})
		if err != nil {
			return "", fmt.Errorf("error creating auth url: %s", err)
		}

		return fmt.Sprintf("Please visit %s to authorize Fynbos for posting the public proof on behalf of you", auth.URL), nil
	}

	_, err = tp.b.Twitter().PostTweet(ctx, &tokens[0], args.Code)
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
