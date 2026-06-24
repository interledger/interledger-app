package sendgrid

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/interledger/interledger-app/go/log"
	"github.com/sendgrid/rest"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"

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

func NewClient(apiKey, fromName, fromEmail string) Client {
	// Override the default API HTTP client. The lib doesn't seem to have a nice way to set this...
	rest.DefaultClient.HTTPClient = otelhttp.DefaultClient

	return &client{
		from:   mail.NewEmail(fromName, fromEmail),
		mailer: sendgrid.NewSendClient(apiKey),
	}
}

func (c *client) SendTemplate(ctx context.Context, subject string, to []Email, templateID string, templateData map[string]interface{}, attachments []mail.Attachment) error {
	msg := new(mail.SGMailV3)
	msg.SetFrom(c.from)
	msg.Subject = subject
	msg.SetTemplateID(templateID)

	for _, t := range to {
		log.Info("Dispatching email via SendGrid",
			zap.String("to", maskEmailAddress(t.Address)),
			zap.String("templateID", templateID),
			zap.String("subject", subject),
		)

		p := mail.NewPersonalization()
		p.DynamicTemplateData = templateData
		p.AddTos(mail.NewEmail(t.Name, t.Address))
		msg.AddPersonalizations(p)
	}

	for _, attachment := range attachments {
		msg.AddAttachment(&attachment)
	}

	resp, err := c.mailer.SendWithContext(ctx, msg)
	if err != nil {
		return fmt.Errorf("%w %s", ErrExternal, err)
	}

	if resp == nil {
		return fmt.Errorf("%w nil response from sendgrid", ErrExternal)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf(
			"%w status=%d body=%q",
			ErrExternal,
			resp.StatusCode,
			truncateBody(resp.Body, 512),
		)
	}

	messageID := firstHeaderValue(resp.Headers, "X-Message-Id")
	log.Info("SendGrid accepted email dispatch",
		zap.Int("status", resp.StatusCode),
		zap.String("messageID", messageID),
		zap.Strings("to", maskEmailAddresses(to)),
		zap.String("templateID", templateID),
		zap.String("subject", subject),
	)

	return nil
}

func firstHeaderValue(headers map[string][]string, key string) string {
	for headerKey, values := range headers {
		if strings.EqualFold(headerKey, key) {
			if len(values) == 0 {
				return ""
			}
			return strings.TrimSpace(values[0])
		}
	}

	return ""
}

func maskEmailAddresses(to []Email) []string {
	masked := make([]string, 0, len(to))
	for _, recipient := range to {
		masked = append(masked, maskEmailAddress(recipient.Address))
	}

	return masked
}

func maskEmailAddress(addr string) string {
	parts := strings.SplitN(addr, "@", 2)
	if len(parts) != 2 {
		if len(addr) <= 3 {
			return addr + "***"
		}
		return addr[:3] + "***"
	}

	localPart := parts[0]
	domain := parts[1]

	prefixLen := 3
	if len(localPart) < prefixLen {
		prefixLen = len(localPart)
	}

	return localPart[:prefixLen] + "***@" + domain
}

func truncateBody(body string, maxLen int) string {
	if maxLen <= 0 || len(body) <= maxLen {
		return body
	}

	return body[:maxLen] + "..."
}
