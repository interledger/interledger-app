package client

import (
	"context"

	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/email/ops"
	"gitlab.com/fynbos/backend/email/sendgrid"
)

var _ email.Client = &client{}

type client struct {
	b ops.Backends
}

func New(b Backends, sendgridAPIKey string) email.Client {

	externalClient := sendgrid.NewClient(sendgridAPIKey)

	ob := &opsBackends{
		Backends: b,
		external: externalClient,
	}

	return &client{
		b: ob,
	}
}

func (c *client) SendMailTemplate(ctx context.Context, walletID string, template email.TemplateID, templateData map[string]interface{}, attachments []email.Attachment) error {
	return ops.SendMailTemplate(ctx, c.b, walletID, template, templateData, attachments)
}

func (c *client) SendPlainText(ctx context.Context, subject, body string, to []string, attachments []email.Attachment) error {
	return ops.SendPlainText(ctx, c.b, subject, body, to, attachments)
}
