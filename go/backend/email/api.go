package email

import (
	"context"
)

type Client interface {
	SendMailTemplate(ctx context.Context, userID string, template TemplateID, personalization map[string]interface{}) error
}
