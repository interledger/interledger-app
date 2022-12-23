package email

import (
	"context"
)

type Client interface {
	SendMailTemplate(ctx context.Context, walletID string, template TemplateID, personalization map[string]interface{}) error
}
