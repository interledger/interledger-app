package slack

import (
	"context"
	"os"
	"sync"

	"gitlab.com/fynbos/log"
	"go.uber.org/zap"

	ext_slack "github.com/slack-go/slack"
)

var api *ext_slack.Client
var initOnce sync.Once

type Channel string

const (
	PersonaChannel Channel = "C053HA9ANCF"
	NotifyGMT      Channel = "C05A6PKHVUY"
	NotifyTabapay  Channel = "C05D4TTGREE"
)

func SendToChannel(ctx context.Context, channel Channel, fromUser, message string) {
	initOnce.Do(func() {
		api = ext_slack.New(os.Getenv("SLACK_TOKEN"))
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
