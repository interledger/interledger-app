package slack

import (
	"context"
	"sync"

	"gitlab.com/fynbos/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"

	ext_slack "github.com/slack-go/slack"
)

var api *ext_slack.Client
var initOnce sync.Once
var slackToken string
var channelID string

// Init configures the Slack notifier. The channel ID is environment-specific
// (set via the SLACK_CHANNEL env var) so that sandbox and production route to
// separate channels. When channel is empty (e.g. local development) no
// notifications are sent.
func Init(token, channel string) {
	slackToken = token
	channelID = channel
}

// SendToChannel posts a notification to the configured Slack channel. It is a
// best-effort, fire-and-forget operation: failures are logged, not returned.
// When no channel is configured (e.g. local development) it is a no-op.
func SendToChannel(ctx context.Context, fromUser, message string) {
	if channelID == "" {
		log.Debug("slack channel not configured, skipping notification", zap.String("from_user", fromUser))
		return
	}

	initOnce.Do(func() {
		api = ext_slack.New(slackToken, ext_slack.OptionHTTPClient(otelhttp.DefaultClient))
	})
	if api == nil {
		return
	}

	// Create the Slack attachment that we will send to the channel
	attachment := ext_slack.Attachment{
		Pretext: message,
	}

	_, _, err := api.PostMessageContext(ctx, channelID, ext_slack.MsgOptionUsername(fromUser), ext_slack.MsgOptionAttachments(attachment))
	if err != nil {
		log.Warn("failed to send message to slack", zap.Error(err))
	}
}
