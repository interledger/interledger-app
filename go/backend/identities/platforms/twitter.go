package platforms

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	twitterscraper "github.com/n0madic/twitter-scraper"
	"gitlab.com/fynbos/backend/identities"
	"go.temporal.io/sdk/workflow"
)

type twitter struct {
	platform identities.Platform
}

func newTwitter(platform identities.Platform) *twitter {
	return &twitter{platform: platform}
}

func (t *twitter) VerifyWorkflow() interface{} {
	return TwitterVerifyWorkflow
}

func (t *twitter) NewVerifyCode() string {
	return uuid.NewString()
}

func (t *twitter) VerifyInstructions() string {
	return `dude tweet this thing`
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
