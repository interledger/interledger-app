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

// add json object for input


func SendMigrationEmailJob(ctx workflow.Context, ) ([]string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}

	ctx = workflow.WithActivityOptions(ctx, ao)
	unverified := []string{}
	err := workflow.ExecuteActivity(ctx, a.SendMigrationEmails).Get(ctx, &unverified)
	if err != nil {
		return nil, err
	}

	return unverified, nil
}

func (a *Activity) SendMigrationEmails(ctx context.Context) ([]string, error) {
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
	

	identities, _, err := client.IdentityAPI.ListIdentities(ctx).PerPage(500).Execute()
	if err != nil {
		return nil, err
	}

	emails := []string{}

	for _, id := range identities {
		traits, ok := id.Traits.(map[string]any)
		if !ok {
			continue
		}
		idEmail, ok := traits["email"].(string)
		if !ok {
			continue
		}
		FirstName, ok := traits["firstNme"].(string)
		if !ok {
			continue
		}
		country, ok := traits["country"].(string)
		if !ok {
			continue
		}
		if country == "CA" {
			continue
		}
		// send email message

		err = a.b.Email().SendMigrationEmail(ctx, idEmail,FirstName,idEmail,idEmail)
		if err != nil {
			// cumulative string array of errors
			emails = append(emails, idEmail)
		}

		log.Warn("log info about verification: ", zap.String("email", idEmail), zap.Any("verifiable_addresses", id.VerifiableAddresses))

	}


	return emails, nil
}
