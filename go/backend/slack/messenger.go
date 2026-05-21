package slack

import (
	"context"
	"fmt"
	"os"
	"sync"

	"gitlab.com/fynbos/env"

	"gitlab.com/fynbos/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"

	ext_slack "github.com/slack-go/slack"
)

var api *ext_slack.Client
var initOnce sync.Once
var slackToken string

type Channel string

const (
	ChannelPersona      Channel = "C091T8JD0DS"
	ChannelNotifyReview Channel = "C091T8JD0DS"
	ChannelNotifyEvents Channel = "C091T8JD0DS"
	ChannelNotifyForms  Channel = "C091T8JD0DS"
	ChannelNotifyErrors Channel = "C091T8JD0DS"
)

func Init(token string) {
	slackToken = token
}

func SendToChannel(ctx context.Context, channel Channel, fromUser, message string) {
	if !env.FeatureSlackNotifications() {
		return
	}

	initOnce.Do(func() {
		api = ext_slack.New(slackToken, ext_slack.OptionHTTPClient(otelhttp.DefaultClient))
	})
	if api == nil {
		return
	}

	message = formatMessageForEnvironment(message)

	// Create the Slack attachment that we will send to the channel
	attachment := ext_slack.Attachment{
		Pretext: message,
	}

	_, _, err := api.PostMessageContext(ctx, string(channel), ext_slack.MsgOptionUsername(fromUser), ext_slack.MsgOptionAttachments(attachment))

	log.Warn("failed to send message to slack", zap.Error(err))
}

func formatMessageForEnvironment(message string) string {
	if label := os.Getenv("SENTRY_ENV_LABEL"); label != "" && label != "prod" {
		return fmt.Sprintf("%s\n*[%s]*", message, label)
	}
	return message
}
