package slack

import (
	"context"
	"os"
	"sync"

	"gitlab.com/fynbos/env"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"gitlab.com/fynbos/log"
	"go.uber.org/zap"

	ext_slack "github.com/slack-go/slack"
)

var api *ext_slack.Client
var initOnce sync.Once

type Channel string

const (
	ChannelPersona      Channel = "C053HA9ANCF"
	ChannelNotifyReview Channel = "C05KABR3Z8U"
	ChannelNotifyEvents Channel = "C05L0Q20RJ9"
	ChannelNotifyForms  Channel = "C05RA9HSNKG"
)

func SendToChannel(ctx context.Context, channel Channel, fromUser, message string) {
	if channel == ChannelNotifyEvents && !env.IsProd() {
		return
	}

	initOnce.Do(func() {
		api = ext_slack.New(os.Getenv("SLACK_TOKEN"), ext_slack.OptionHTTPClient(otelhttp.DefaultClient))
	})
	if api == nil {
		return
	}

	// Create the Slack attachment that we will send to the channel
	attachment := ext_slack.Attachment{
		Pretext: message,
	}

	_, _, err := api.PostMessageContext(ctx, string(channel), ext_slack.MsgOptionUsername(fromUser), ext_slack.MsgOptionAttachments(attachment))

	if err != nil {
		log.Error("failed to send message to slack", zap.Error(err))
	}
}
