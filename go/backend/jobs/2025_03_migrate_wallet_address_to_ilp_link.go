package jobs

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

type MigrateWalletAddressesParams struct {
	Revert string `json:"revert"`
}

type MigrationParams struct {
	CurrentWalletAddress string
	NewWalletAddress     string
	CurrentAppDomain     string
	NewAppDomain         string
	CurrentNamespace     string
	NewNamespace         string
}

func MigrateWalletAddressesToIlpLinkJob(ctx workflow.Context, params MigrateWalletAddressesParams) error {
	var domainInfo MigrationParams
	if params.Revert == "true" {
		domainInfo = getRevertedMigrationParams()
	} else {
		domainInfo = getMigrationParams()
	}

	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	log.Info("Starting job MigrateWalletAddressesToIlpLinkJob update wallet to " + domainInfo.NewWalletAddress)

	if err := executeActivity(ctx, a.UpdateBackendWalletRootToIlpActivity, "backend", domainInfo); err != nil {
		return err
	}
	if err := executeActivity(ctx, a.UpdateRafikiWalletRootToIlpActivity, "rafiki", domainInfo); err != nil {
		return err
	}
	if err := executeActivity(ctx, a.UpdateRafikiAuthWalletRootToIlpActivity, "rafiki auth", domainInfo); err != nil {
		return err
	}
	log.Info("Completed job MigrateWalletAddressesToIlpLinkJob update wallet to " + domainInfo.NewWalletAddress)
	return nil
}

func executeActivity(ctx workflow.Context, activityFunc interface{}, activityName string, domainInfo MigrationParams) error {
	logger := workflow.GetLogger(ctx)
	err := workflow.ExecuteActivity(ctx, activityFunc, domainInfo).Get(ctx, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("Migrate [%s] to %s failed", activityName, domainInfo.NewWalletAddress), "Error", err)
		return err
	}
	logger.Info(fmt.Sprintf("Migrate [%s] to %s completed", activityName, domainInfo.NewWalletAddress))
	return nil
}

func (a *Activity) UpdateBackendWalletRootToIlpActivity(ctx context.Context, domainInfo MigrationParams) error {
	log.Info(fmt.Sprintf("Starting [backend] wallet update to %s", domainInfo.NewWalletAddress))

	queries := []struct {
		query string
		args  []interface{}
	}{
		{"UPDATE openpayments_incoming_payment SET sender_wallet_address = replace(sender_wallet_address, $1, $2) WHERE sender_wallet_address ILIKE '%' || $1 || '%';", []interface{}{domainInfo.CurrentWalletAddress, domainInfo.NewWalletAddress}},
		{"UPDATE openpayments_incoming_payment SET receiver_wallet_address = replace(receiver_wallet_address, $1, $2) WHERE receiver_wallet_address ILIKE '%' || $1 || '%';", []interface{}{domainInfo.CurrentWalletAddress, domainInfo.NewWalletAddress}},
		{"UPDATE openpayments_incoming_payment SET created_by = replace(created_by, $1, $2) WHERE created_by ILIKE '%' || $1 || '%';", []interface{}{domainInfo.CurrentWalletAddress, domainInfo.NewWalletAddress}},

		{"UPDATE openpayments_outgoing_payment SET sender_wallet_address = replace(sender_wallet_address, $1, $2) WHERE sender_wallet_address ILIKE '%' || $1 || '%';", []interface{}{domainInfo.CurrentWalletAddress, domainInfo.NewWalletAddress}},
		{"UPDATE openpayments_outgoing_payment SET receiver_wallet_address = replace(receiver_wallet_address, $1, $2) WHERE receiver_wallet_address ILIKE '%' || $1 || '%';", []interface{}{domainInfo.CurrentWalletAddress, domainInfo.NewWalletAddress}},
		{"UPDATE openpayments_outgoing_payment SET created_by = replace(created_by, $1, $2) WHERE created_by ILIKE '%' || $1 || '%';", []interface{}{domainInfo.CurrentWalletAddress, domainInfo.NewWalletAddress}},

		{"UPDATE openpayments_quotes SET sender_wallet_address = replace(sender_wallet_address, $1, $2) WHERE sender_wallet_address ILIKE '%' || $1 || '%';", []interface{}{domainInfo.CurrentWalletAddress, domainInfo.NewWalletAddress}},
		{"UPDATE openpayments_quotes SET receiver_wallet_address = replace(receiver_wallet_address, $1, $2) WHERE receiver_wallet_address ILIKE '%' || $1 || '%';", []interface{}{domainInfo.CurrentWalletAddress, domainInfo.NewWalletAddress}},
		{"UPDATE openpayments_quotes SET created_by = replace(created_by, $1, $2) WHERE created_by ILIKE '%' || $1 || '%';", []interface{}{domainInfo.CurrentWalletAddress, domainInfo.NewWalletAddress}},

		{"UPDATE payments SET receiver_id = replace(receiver_id, $1, $2) WHERE receiver_id ILIKE '%' || $1 || '%';", []interface{}{domainInfo.CurrentWalletAddress, domainInfo.NewWalletAddress}},
		{"UPDATE contacts SET payment_pointer = replace(payment_pointer, $1, $2) WHERE payment_pointer ILIKE '%' || $1 || '%';", []interface{}{domainInfo.CurrentWalletAddress, domainInfo.NewWalletAddress}},

		{"UPDATE transactions SET source = replace(source, $1, $2) WHERE source ILIKE '%' || $1 || '%';", []interface{}{domainInfo.CurrentWalletAddress, domainInfo.NewWalletAddress}},
		{"UPDATE transactions SET destination_identity = replace(destination_identity, $1, $2) WHERE destination_identity ILIKE '%' || $1 || '%';", []interface{}{domainInfo.CurrentWalletAddress, domainInfo.NewWalletAddress}},
		{"UPDATE transactions SET destination = replace(destination, $1, $2) WHERE destination ILIKE '%' || $1 || '%';", []interface{}{domainInfo.CurrentWalletAddress, domainInfo.NewWalletAddress}},
		{"UPDATE transactions SET title = replace(title, $1, $2) WHERE title ILIKE '%' || $1 || '%';", []interface{}{domainInfo.CurrentWalletAddress, domainInfo.NewWalletAddress}},

		{"UPDATE wallet_addresses SET url = replace(url, $1, $2) WHERE url ILIKE '%' || $1 || '%';", []interface{}{domainInfo.CurrentWalletAddress, domainInfo.NewWalletAddress}},
		{"UPDATE wallet_keys SET name = replace(name, $1, $2);", []interface{}{domainInfo.CurrentNamespace, domainInfo.NewNamespace}},

		{"UPDATE slack_authorizations SET redirect_url = replace(redirect_url, $1, $2) WHERE redirect_url ILIKE '%' || $1 || '%';", []interface{}{domainInfo.CurrentAppDomain, domainInfo.NewAppDomain}},
		{"UPDATE twitter_authorizations SET redirect_url = replace(redirect_url, $1, $2) WHERE redirect_url ILIKE '%' || $1 || '%';", []interface{}{domainInfo.CurrentAppDomain, domainInfo.NewAppDomain}},
	}

	err := ExecuteTransaction(a.b.DB(), queries)
	if err != nil {
		return err
	}
	return nil
}

func (a *Activity) UpdateRafikiWalletRootToIlpActivity(ctx context.Context, domainInfo MigrationParams) error {
	log.Info(fmt.Sprintf("Starting [rafiki] wallet update to %s", domainInfo.NewWalletAddress))
	db, err := DbConnection(a.cfg.RafikiDBURL)
	if err != nil {
		log.Error("Error establishing db connection: %v", zap.Error(err))
		return err
	}
	defer db.Close()

	queries := []struct {
		query string
		args  []interface{}
	}{
		{"UPDATE \"incomingPayments\" SET client = replace(client, $1, $2) WHERE client ILIKE '%' || $1 || '%';", []interface{}{domainInfo.CurrentWalletAddress, domainInfo.NewWalletAddress}},
		{"UPDATE quotes SET receiver = replace(receiver, $1, $2) WHERE receiver ILIKE '%' || $1 || '%';", []interface{}{domainInfo.CurrentWalletAddress, domainInfo.NewWalletAddress}},
		{"UPDATE quotes SET client = replace(client, $1, $2) WHERE client ILIKE '%' || $1 || '%';", []interface{}{domainInfo.CurrentWalletAddress, domainInfo.NewWalletAddress}},
		{"UPDATE \"walletAddresses\" SET url = replace(url, $1, $2) WHERE url ILIKE '%' || $1 || '%';", []interface{}{domainInfo.CurrentWalletAddress, domainInfo.NewWalletAddress}},
		{"UPDATE \"webhookEvents\" SET data = REPLACE(data::TEXT, $1, $2)::JSON WHERE data::TEXT LIKE '%' || $1 || '%';", []interface{}{domainInfo.CurrentWalletAddress, domainInfo.NewWalletAddress}},
	}

	trErr := ExecuteTransaction(db, queries)
	if trErr != nil {
		return trErr
	}
	return nil
}

func (a *Activity) UpdateRafikiAuthWalletRootToIlpActivity(ctx context.Context, domainInfo MigrationParams) error {
	log.Info(fmt.Sprintf("Starting [rafiki auth] wallet update to %s", domainInfo.NewWalletAddress))
	db, err := DbConnection(a.cfg.RafikiAuthDBURL)
	if err != nil {
		log.Error("Error establishing db connection: %v", zap.Error(err))
		return err
	}
	defer db.Close()

	queries := []struct {
		query string
		args  []interface{}
	}{
		{"UPDATE accesses SET identifier = replace(identifier, $1, $2) WHERE identifier ILIKE '%' || $1 || '%';", []interface{}{domainInfo.CurrentWalletAddress, domainInfo.NewWalletAddress}},
		{"UPDATE grants SET client = replace(client, $1, $2) WHERE client ILIKE '%' || $1 || '%';", []interface{}{domainInfo.CurrentWalletAddress, domainInfo.NewWalletAddress}},
	}

	trErr := ExecuteTransaction(db, queries)
	if trErr != nil {
		return trErr
	}
	return nil
}

func getMigrationParams() MigrationParams {
	prod := MigrationParams{
		CurrentWalletAddress: "fynbos.me",
		NewWalletAddress:     "ilp.link",
		CurrentAppDomain:     "fynbos.app",
		NewAppDomain:         "interledger.app",
		CurrentNamespace:     "Fynbos",
		NewNamespace:         "Interledger",
	}

	dev := MigrationParams{
		CurrentWalletAddress: "eu1.fynbos.me",
		NewWalletAddress:     "sandbox.ilp.link",
		CurrentAppDomain:     "eu1.fynbos.dev",
		NewAppDomain:         "sandbox.interledger.app",
		CurrentNamespace:     "Fynbos",
		NewNamespace:         "Interledger",
	}

	local := prod

	switch {
	case env.IsLocal():
		return local
	case env.IsDev():
		return dev
	case env.IsProd():
		return prod
	default:
		log.Error("Environment not set")
		return MigrationParams{}
	}
}

func getRevertedMigrationParams() MigrationParams {
	prod := MigrationParams{
		CurrentWalletAddress: "ilp.link",
		NewWalletAddress:     "fynbos.me",
		CurrentAppDomain:     "interledger.app",
		NewAppDomain:         "fynbos.app",
		CurrentNamespace:     "Interledger",
		NewNamespace:         "Fynbos",
	}

	dev := MigrationParams{
		CurrentWalletAddress: "sandbox.ilp.link",
		NewWalletAddress:     "eu1.fynbos.me",
		CurrentAppDomain:     "sandbox.interledger.app",
		NewAppDomain:         "eu1.fynbos.dev",
		CurrentNamespace:     "Interledger",
		NewNamespace:         "Fynbos",
	}

	local := prod

	switch {
	case env.IsLocal():
		return local
	case env.IsDev():
		return dev
	case env.IsProd():
		return prod
	default:
		log.Error("Environment not set")
		return MigrationParams{}
	}
}
