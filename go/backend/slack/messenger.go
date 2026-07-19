package slack

import (
	"context"
	"sync"

	"github.com/interledger/interledger-app/go/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"

	ext_slack "github.com/slack-go/slack"
)

type Channel string

const (
	ChannelSignupKYC   Channel = "signup_kyc"
	ChannelTransaction Channel = "transaction"
	ChannelError       Channel = "error"
)

var allChannels = []Channel{ChannelSignupKYC, ChannelTransaction, ChannelError}

var api *ext_slack.Client
var initOnce sync.Once
var slackToken string
var channelIDs map[Channel]string

func Init(token string, channels map[Channel]string) {
	slackToken = token
	channelIDs = channels
	for _, ch := range allChannels {
		if channelIDs[ch] == "" {
			log.Info("slack channel not configured", zap.String("channel", string(ch)))
		}
	}
}

// workflow code must notify slack through this activity
// rather than calling SendToChanneldirectly,
// which would run a blocking HTTP call on the workflow goroutine and re-fire on every replay
func SendToChannelActivity(ctx context.Context, channel Channel, fromUser, message string) error {
	SendToChannel(ctx, channel, fromUser, message)
	return nil
}

func SendToChannel(ctx context.Context, channel Channel, fromUser, message string) {
	// mirror every notification in every environment,
	// regardless of whether slack is configured or reachable
	var sendErr error
	defer func() {
		fields := []zap.Field{
			zap.String("channel", string(channel)),
			zap.String("from_user", fromUser),
			zap.String("message", message),
		}

		if sendErr != nil {
			log.Warn("failed to send message to slack", append(fields, zap.Error(sendErr))...)
			return
		}

		log.Info("slack_msg", fields...)
	}()

	channelID := channelIDs[channel]
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

	_, _, sendErr = api.PostMessageContext(ctx, channelID, ext_slack.MsgOptionUsername(fromUser), ext_slack.MsgOptionAttachments(attachment))
}
