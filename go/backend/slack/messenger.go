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

func Init(token, channel string) {
	slackToken = token
	channelID = channel
	if channelID == "" {
		log.Info("slack notifications disabled: SLACK_CHANNEL not set")
	}
}

func SendToChannel(ctx context.Context, fromUser, message string) {
	// mirror every notification in every environment,
	// regardless of whether slack is  configured or reachable
	log.Info("slack_msg", zap.String("from_user", fromUser), zap.String("message", message))

	if channelID == "" {
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
