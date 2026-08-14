package jobs

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/interledger/interledger-app/go/backend/country"
	"github.com/interledger/interledger-app/go/log"
	kratos "github.com/ory/kratos-client-go"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

const (
	migrationEmailPageSize    = 500
	migrationEmailMaxPages    = 100 // safety valve: 50k identities
	migrationEmailConcurrency = 20
	migrationEmailSendTimeout = 2 * time.Minute

	// migrationRegionAll targets every user.
	migrationRegionAll = "ALL"
)

// SendMigrationEmailParams is the input for SendMigrationEmailJob.
//
// Targeting:
//   - Email set: send only to those addresses, comma separated. Region is ignored,
//     and every address must match a user or the job fails.
//   - Email empty: Region is required — "ALL", "EU" or an ISO country code ("US", "ZA").
//
// Paragraphs are SendGrid template blocks, e.g. {"paragraph": "..."}, {"heading": "..."}.
// A greeting with the user's first name is prepended; the CTA is always login.
type SendMigrationEmailParams struct {
	Subject    string                   `json:"subject"`
	Paragraphs []map[string]interface{} `json:"paragraphs"`
	Region     string                   `json:"region"`
	Email      string                   `json:"email"`
}

// MigrationEmailRecipient is a user selected to receive a migration email.
type MigrationEmailRecipient struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
}

// SendMigrationEmailJob emails users about a migration. Returns the addresses whose
// send failed; re-run with those addresses in Email to retry them.
func SendMigrationEmailJob(ctx workflow.Context, params SendMigrationEmailParams) ([]string, error) {
	if err := validateSendMigrationEmailParams(params); err != nil {
		return nil, err
	}

	var a *Activity
	listCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	})

	var recipients []MigrationEmailRecipient
	if err := workflow.ExecuteActivity(listCtx, a.ListMigrationEmailRecipients, params).Get(listCtx, &recipients); err != nil {
		return nil, err
	}

	sendCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: migrationEmailSendTimeout,
		RetryPolicy: &temporal.RetryPolicy{
			// At most once: a retry after an ambiguous timeout would email twice.
			MaximumAttempts: 1,
		},
	})

	return dispatchMigrationEmails(sendCtx, a, recipients, params.Subject, params.Paragraphs), nil
}

func dispatchMigrationEmails(ctx workflow.Context, a *Activity, recipients []MigrationEmailRecipient, subject string, paragraphs []map[string]interface{}) []string {
	type pendingEmail struct {
		email  string
		future workflow.Future
	}
	var pending []pendingEmail
	var failed []string
	drain := func() {
		for _, p := range pending {
			if err := p.future.Get(ctx, nil); err != nil {
				workflow.GetLogger(ctx).Warn("failed to send migration email", zap.String("email", p.email), zap.Error(err))
				failed = append(failed, p.email)
			}
		}
		pending = nil
	}

	for _, r := range recipients {
		pending = append(pending, pendingEmail{
			email:  r.Email,
			future: workflow.ExecuteActivity(ctx, a.SendMigrationEmailToRecipient, subject, r.Email, r.FirstName, paragraphs),
		})
		if len(pending) >= migrationEmailConcurrency {
			drain()
		}
	}
	drain()
	return failed
}

func (a *Activity) ListMigrationEmailRecipients(ctx context.Context, params SendMigrationEmailParams) ([]MigrationEmailRecipient, error) {
	if err := validateSendMigrationEmailParams(params); err != nil {
		return nil, err
	}

	targetEmails := parseMigrationEmails(params.Email)

	// nil countries = no country filter (address-targeted send, or region ALL).
	var countries map[country.Country]bool
	if len(targetEmails) == 0 {
		var err error
		countries, err = resolveMigrationCountries(params.Region)
		if err != nil {
			return nil, err
		}
	}

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
	kratosClient := kratos.NewAPIClient(config)

	// Admin API is at server index 1 (same as user/ops ListIdentities callers).
	adminCtx := context.WithValue(ctx, kratos.ContextServerIndex, 1)

	var recipients []MigrationEmailRecipient
	seen := map[string]bool{}
	collect := func(identities []kratos.Identity) {
		for _, id := range identities {
			recipient, ok := migrationRecipientFromIdentity(id.Traits, targetEmails, countries)
			if !ok {
				continue
			}
			key := strings.ToLower(recipient.Email)
			if seen[key] {
				continue
			}
			seen[key] = true
			recipients = append(recipients, recipient)
		}
	}

	if len(targetEmails) > 0 {
		// Email is the password credential identifier, so Kratos can look each up directly.
		var missing []string
		for address := range targetEmails {
			identities, _, err := kratosClient.IdentityAPI.ListIdentities(adminCtx).
				CredentialsIdentifier(address).
				Execute()
			if err != nil {
				return nil, err
			}
			collect(identities)
			if !seen[address] {
				missing = append(missing, address)
			}
		}
		if len(missing) > 0 {
			sort.Strings(missing)
			return nil, fmt.Errorf("no user found for: %s", strings.Join(missing, ", "))
		}
		log.Info("listed migration email recipients", zap.Int("count", len(recipients)))
		return recipients, nil
	}

	pageToken := ""
	for page := 0; ; page++ {
		if page == migrationEmailMaxPages {
			return nil, fmt.Errorf("stopped after %d pages of identities: refusing to email a partial list", migrationEmailMaxPages)
		}

		req := kratosClient.IdentityAPI.ListIdentities(adminCtx).PageSize(migrationEmailPageSize)
		if pageToken != "" {
			req = req.PageToken(pageToken)
		}
		identities, resp, err := req.Execute()
		if err != nil {
			return nil, err
		}
		collect(identities)

		next := nextPageToken(resp)
		if next == "" || next == pageToken || len(identities) == 0 {
			break
		}
		pageToken = next
	}

	log.Info("listed migration email recipients",
		zap.Int("count", len(recipients)),
		zap.String("region", params.Region),
	)
	return recipients, nil
}

// nextPageToken returns the Link header's rel="next" page_token, or "" on the last page.
func nextPageToken(resp *http.Response) string {
	if resp == nil {
		return ""
	}
	// Kratos sends one comma-separated Link header, but HTTP allows repeats.
	for _, header := range resp.Header.Values("Link") {
		for _, link := range strings.Split(header, ",") {
			target, rest, ok := strings.Cut(link, ";")
			if !ok {
				continue
			}
			isNext := false
			for _, param := range strings.Split(rest, ";") {
				if strings.EqualFold(strings.TrimSpace(param), `rel="next"`) {
					isNext = true
					break
				}
			}
			if !isNext {
				continue
			}
			u, err := url.Parse(strings.Trim(strings.TrimSpace(target), "<>"))
			if err != nil {
				continue
			}
			return u.Query().Get("page_token")
		}
	}
	return ""
}

func (a *Activity) SendMigrationEmailToRecipient(ctx context.Context, subject, sendTo, firstName string, paragraphs []map[string]interface{}) error {
	return a.b.Email().SendMigrationEmail(ctx, subject, sendTo, firstName, paragraphs)
}

func validateSendMigrationEmailParams(params SendMigrationEmailParams) error {
	if strings.TrimSpace(params.Subject) == "" {
		return errors.New("subject is required")
	}
	if len(params.Paragraphs) == 0 {
		return errors.New("paragraphs are required")
	}
	if len(parseMigrationEmails(params.Email)) > 0 {
		return nil
	}
	if strings.TrimSpace(params.Region) == "" {
		return errors.New("region is required when email is not set")
	}
	if _, err := resolveMigrationCountries(params.Region); err != nil {
		return err
	}
	return nil
}

// parseMigrationEmails normalises the Email param: one address, or several comma separated.
func parseMigrationEmails(email string) map[string]bool {
	addresses := map[string]bool{}
	for _, address := range strings.Split(email, ",") {
		address = strings.ToLower(strings.TrimSpace(address))
		if address != "" {
			addresses[address] = true
		}
	}
	return addresses
}

// resolveMigrationCountries maps a region to its countries; nil means no filter.
func resolveMigrationCountries(region string) (map[country.Country]bool, error) {
	region = strings.ToUpper(strings.TrimSpace(region))
	if region == "" {
		return nil, errors.New("region is required")
	}
	if region == migrationRegionAll {
		return nil, nil
	}
	if region == "EU" {
		out := make(map[country.Country]bool, len(country.EUCountries))
		for c := range country.EUCountries {
			out[c] = true
		}
		return out, nil
	}
	c := country.Country(region)
	if !c.Valid() {
		return nil, fmt.Errorf("unknown region %q — use ALL, EU or a valid ISO country code", region)
	}
	return map[country.Country]bool{c: true}, nil
}

func migrationRecipientFromIdentity(traits any, targetEmails map[string]bool, countries map[country.Country]bool) (MigrationEmailRecipient, bool) {
	traitsMap, ok := traits.(map[string]any)
	if !ok {
		return MigrationEmailRecipient{}, false
	}

	email, _ := traitsMap["email"].(string)
	email = strings.TrimSpace(email)
	if email == "" {
		return MigrationEmailRecipient{}, false
	}

	if len(targetEmails) > 0 {
		if !targetEmails[strings.ToLower(email)] {
			return MigrationEmailRecipient{}, false
		}
	} else if countries != nil { // nil = no filter (region ALL)
		code, _ := traitsMap["countryCode"].(string)
		c := country.Country(strings.ToUpper(strings.TrimSpace(code)))
		if !countries[c] {
			return MigrationEmailRecipient{}, false
		}
	}

	firstName, _ := traitsMap["firstName"].(string)
	return MigrationEmailRecipient{
		Email:     email,
		FirstName: strings.TrimSpace(firstName),
	}, true
}
