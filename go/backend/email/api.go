package email

import (
	"context"
)

type Client interface {
	SendMailTemplate(ctx context.Context, walletID string, template TemplateID, personalization map[string]interface{}, attachments []Attachment) error
	SendPlainText(ctx context.Context, subject, body string, to []string, attachments []Attachment) error
}
