package ops

import (
	"context"
	"fmt"

	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/email/sendgrid"
)

func SendMailTemplate(ctx context.Context, b Backends, walletID string, template email.TemplateID, templateData map[string]interface{}) error {
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

	err = b.External().SendTemplate(ctx, template.Subject(), emails, template.String(), templateData)
	if err != nil {
		return fmt.Errorf("%w %s", email.ErrInternal, err)
	}

	return nil
}
