package jobs

import (
	"context"
	"errors"
	"time"

	"github.com/interledger/interledger-app/go/log"
	kratos "github.com/ory/kratos-client-go"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type Identities struct {
	IdentityID string
	Email      string
}

func EnableSendVerificationEmailToUnverifiedUserJob(ctx workflow.Context, email string) ([]string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}

	ctx = workflow.WithActivityOptions(ctx, ao)
	unverified := []string{}
	err := workflow.ExecuteActivity(ctx, a.EnableSendVerificationEmails, email).Get(ctx, &unverified)
	if err != nil {
		return nil, err
	}

	return unverified, nil
}

func (a *Activity) EnableSendVerificationEmails(ctx context.Context, email string) ([]string, error) {
	kratosURL := a.cfg.KratosURL
	kratosAdminURL := a.cfg.KratosAdminURL
	if kratosURL == "" || kratosAdminURL == "" {
		return nil, errors.New("kratos URLs are not set")
	}

	config := kratos.NewConfiguration()
	config.HTTPClient = otelhttp.DefaultClient
	config.Servers = kratos.ServerConfigurations{
		{URL: kratosURL, Description: "Public Kratos"},
		{URL: kratosAdminURL, Description: "Admin Kratos"},
	}

	client := kratos.NewAPIClient(config)
	noVerificationAddresses := []kratos.Identity{}

	identities, _, err := client.IdentityApi.ListIdentities(ctx).PerPage(500).Execute()
	if err != nil {
		return nil, err
	}

	for _, id := range identities {
		traits, ok := id.Traits.(map[string]any)
		if !ok {
			continue
		}
		idEmail, ok := traits["email"].(string)
		if !ok {
			continue
		}

		if email != "" && idEmail != email {
			continue
		}

		log.Warn("log info about verification: ", zap.String("email", idEmail), zap.Any("verifiable_addresses", id.VerifiableAddresses))
		if len(id.VerifiableAddresses) == 0 {
			noVerificationAddresses = append(noVerificationAddresses, id)

		}
	}

	//update identities without verifiable addresses we need to update traits to add an verifiable_addresses
	emails := []string{}
	for _, id := range noVerificationAddresses {
		traits, ok := id.Traits.(map[string]any)
		if !ok {
			log.Warn("Invalid traits format", zap.String("identity_id", id.Id))
			continue
		}
		email, ok := traits["email"].(string)
		if !ok {
			log.Warn("Email not found in traits", zap.String("identity_id", id.Id))
			continue
		}

		update := kratos.UpdateIdentityBody{Traits: traits}
		_, _, err = kratos.IdentityApi.UpdateIdentity(client.IdentityApi, ctx, id.Id).
			UpdateIdentityBody(update).
			Execute()
		if err != nil {
			log.Warn("Failed to update identity", zap.Error(err), zap.String("identity_id", id.Id))
		} else {
			emails = append(emails, email)
		}

	}

	return emails, nil
}
