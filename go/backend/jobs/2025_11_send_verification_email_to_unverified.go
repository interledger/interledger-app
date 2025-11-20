package jobs

import (
	"context"
	"errors"
	"os"
	"time"

	kratos "github.com/ory/kratos-client-go"
	"gitlab.com/fynbos/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)


 type Identities struct {
	IdentityID string
	Email      string
}
func SendVerificationEmailToUnverifiedUserJob(ctx workflow.Context, email string) ([]string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}

	ctx = workflow.WithActivityOptions(ctx, ao)
	unverified := []string{}
	err := workflow.ExecuteActivity(ctx, a.SendVerificationEmails, email).Get(ctx, &unverified)
		if err != nil {
		return nil, err
	}
	
	return unverified, nil
}

func (a *Activity) SendVerificationEmails(ctx context.Context, email string) ([]string, error) {
	kratosURL := os.Getenv("KRATOS_URL")
	kratosAdminURL := os.Getenv("KRATOS_ADMIN_URL")
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
	unverified := []Identities{}
	noVerificationAddresses := []kratos.Identity{}

	identities, _, err := client.IdentityApi.ListIdentities(ctx).Execute()
	if err != nil {
		return nil, err
	}

	for _, id := range identities {
		traits, ok := id.Traits.(map[string]interface{})
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

		if id.VerifiableAddresses == nil {
			noVerificationAddresses = append(noVerificationAddresses, id)
			unverified = append(unverified, Identities{
				IdentityID: id.Id,
				Email:      idEmail,
			})
			continue
		}

		for _, addr := range id.VerifiableAddresses {
			if addr.Via == "email" && !addr.Verified {
				if email == "" || addr.Value == email {
					unverified = append(unverified, Identities{
						IdentityID: id.Id,
						Email:      addr.Value,
					})
				}
			}
		}
	}

	//update identities without verifiable addresses we need to update traits to add an verifiable_addresses
	for _, id := range noVerificationAddresses {
		traits, ok := id.Traits.(map[string]interface{})
		if !ok {
			log.Warn("Invalid traits format", zap.String("identity_id", id.Id))
			continue
		}

		update := kratos.UpdateIdentityBody{Traits: traits}
		_, _, err = kratos.IdentityApi.UpdateIdentity(client.IdentityApi, ctx, id.Id).
			UpdateIdentityBody(update).
			Execute()
		if err != nil {
			log.Warn("Failed to update identity", zap.Error(err), zap.String("identity_id", id.Id))
		}
	}

	//send verification flows
	emails:= []string{}
	for _, u := range unverified {
		flow, _, err := client.FrontendApi.CreateNativeVerificationFlow(ctx).Execute()
		if err != nil {
			log.Warn("Failed to create verification flow", zap.Error(err), zap.String("user_email", u.Email))
			continue
		}

		body := kratos.UpdateVerificationFlowBody{
			UpdateVerificationFlowWithLinkMethod: &kratos.UpdateVerificationFlowWithLinkMethod{
				Method: "link",
				Email:  u.Email,
			},
		}

		_, _, err = client.FrontendApi.UpdateVerificationFlow(ctx).
			Flow(flow.Id).
			UpdateVerificationFlowBody(body).
			Execute()
		if err != nil {
			log.Warn("Failed to send verification email", zap.Error(err), zap.String("user_email", u.Email))
		}
		emails = append(emails, u.Email)
	}

	return emails, nil
}

