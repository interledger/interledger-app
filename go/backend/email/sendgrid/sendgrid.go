package sendgrid

import (
	"context"
	"errors"
	"fmt"

	"github.com/sendgrid/rest"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type Client interface {
	SendTemplate(ctx context.Context, subject string, to []Email, templateID string, templateData map[string]interface{}, attachments []mail.Attachment) error
}

type Email mail.Email

var ErrExternal = errors.New("sendgrid_external: failed to send email")

type client struct {
	from   *mail.Email
	mailer *sendgrid.Client
}

func NewClient(apiKey string) Client {
	// Override the default API HTTP client. The lib doesn't seem to have a nice way to set this...
	rest.DefaultClient.HTTPClient = otelhttp.DefaultClient

	return &client{
		from:   mail.NewEmail("Interledger", "support@interledger.app"),
		mailer: sendgrid.NewSendClient(apiKey),
	}
}

func (c *client) SendTemplate(ctx context.Context, subject string, to []Email, templateID string, templateData map[string]interface{}, attachments []mail.Attachment) error {
	msg := new(mail.SGMailV3)
	msg.SetFrom(c.from)
	msg.Subject = subject
	msg.SetTemplateID(templateID)

	for _, t := range to {
		p := mail.NewPersonalization()
		p.DynamicTemplateData = templateData
		p.AddTos(mail.NewEmail(t.Name, t.Address))
		msg.AddPersonalizations(p)
	}

	for _, attachment := range attachments {
		msg.AddAttachment(&attachment)
	}

	_, err := c.mailer.SendWithContext(ctx, msg)
	if err != nil {
		return fmt.Errorf("%w %s", ErrExternal, err)
	}

	return nil
}
