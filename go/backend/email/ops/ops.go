package ops

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/email/sendgrid"
)

func SendMailTemplate(ctx context.Context, b Backends, walletID string, template email.TemplateID, templateData map[string]interface{}, attachments []email.Attachment) error {
	if !template.IsValid() {
		return fmt.Errorf("%w %s is not a known template ID", email.ErrInvalidTemplate, template)
	}

	users, err := b.Users().ListUsers(ctx, walletID)
	if err != nil {
		return err
	}
	if len(users) < 1 {
		return fmt.Errorf("%w wallet has (%d) users associated", email.ErrInternal, len(users))
	}

	id, err := b.KYC().GetIndividualDetails(ctx, walletID)
	if err != nil {
		return err
	}

	var emails []sendgrid.Email
	for _, u := range users {
		emails = append(emails, sendgrid.Email{
			Name:    id.FirstName + " " + id.LastName,
			Address: u.Email,
		})
	}

	mailAttachments := make([]mail.Attachment, len(attachments))
	for i, attachment := range attachments {
		mailAttachments[i] = mail.Attachment{
			Content:     base64.StdEncoding.EncodeToString(attachment.Content),
			Type:        attachment.ContentType,
			Filename:    attachment.Name,
			Disposition: "attachment",
		}
	}

	err = b.External().SendTemplate(ctx, template.Subject(), emails, template.String(), templateData, mailAttachments)
	if err != nil {
		return fmt.Errorf("%w %s", email.ErrInternal, err)
	}

	return nil
}

func SendPlainText(ctx context.Context, b Backends, subject, body string, to []string, attachments []email.Attachment) error {
	var emails []sendgrid.Email
	for _, address := range to {
		emails = append(emails, sendgrid.Email{
			Address: address,
		})
	}

	mailAttachments := make([]mail.Attachment, len(attachments))
	for i, attachment := range attachments {
		mailAttachments[i] = mail.Attachment{
			Content:     base64.StdEncoding.EncodeToString(attachment.Content),
			Type:        attachment.ContentType,
			Filename:    attachment.Name,
			Disposition: "attachment",
		}
	}

	err := b.External().SendPlainText(ctx, subject, body, emails, mailAttachments)
	if err != nil {
		return fmt.Errorf("%w %s", email.ErrInternal, err)
	}

	return nil
}
