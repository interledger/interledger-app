package jobs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"gitlab.com/fynbos/backend/country"
	"go.temporal.io/sdk/workflow"
)

type DisabledAccountsTabParams struct {
	// Region — disables accounts tab for all countries in a region. Supported values: "EU"
	Region string
	// Country — disables accounts tab for a single ISO country code (e.g. "DE", "FR", "US", "GB")
	Country string
}

func DisabledAccountsTabWorkflow(ctx workflow.Context, params DisabledAccountsTabParams) error {
	if _, err := resolveCountries(params); err != nil {
		return err
	}

	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("DisabledAccountsTabWorkflow started", "region", params.Region, "country", params.Country)

	var rowsAffected int64
	err := workflow.ExecuteActivity(ctx, a.DisableAccountsTab, params).Get(ctx, &rowsAffected)
	if err != nil {
		logger.Error("DisabledAccountsTabWorkflow failed", "error", err)
		return err
	}

	logger.Info("DisabledAccountsTabWorkflow complete", "region", params.Region, "country", params.Country, "rowsAffected", rowsAffected)
	return nil
}

func (a *Activity) DisableAccountsTab(ctx context.Context, params DisabledAccountsTabParams) (int64, error) {
	countries, err := resolveCountries(params)
	if err != nil {
		return 0, err
	}

	codes := make([]string, len(countries))
	for i, c := range countries {
		codes[i] = string(c)
	}

	result, err := a.b.DB().ExecContext(ctx, `
		UPDATE wallet_features
		SET accounts_tab_enabled = false
		FROM wallets
		WHERE wallet_features.wallet_id = wallets.id
		AND wallets.country = ANY($1)
	`, pq.Array(codes))
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func resolveCountries(params DisabledAccountsTabParams) ([]country.Country, error) {
	if params.Region != "" && params.Country != "" {
		return nil, fmt.Errorf("provide either Region or Country, not both")
	}

	if params.Country != "" {
		c := country.Country(strings.ToUpper(params.Country))
		if !c.Valid() {
			return nil, fmt.Errorf("unknown country code %q", params.Country)
		}
		return []country.Country{c}, nil
	}

	switch strings.ToUpper(params.Region) {
	case "EU":
		countries := make([]country.Country, 0, len(country.EUCountries))
		for c := range country.EUCountries {
			countries = append(countries, c)
		}
		return countries, nil
	default:
		return nil, fmt.Errorf("unknown region %q — supported values: EU", params.Region)
	}
}
