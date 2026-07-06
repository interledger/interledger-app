package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/interledger/interledger-app/go/backend/kyc"
	"github.com/interledger/interledger-app/go/backend/slack"

	"github.com/interledger/interledger-app/go/backend/notify"

	"github.com/interledger/interledger-app/go/backend/wallets"

	"github.com/interledger/interledger-app/go/backend/linkedaccounts"

	"github.com/google/uuid"
	"github.com/interledger/interledger-app/go/backend/identities"
	"github.com/interledger/interledger-app/go/backend/identities/platforms"
	"github.com/interledger/interledger-app/go/env"
	"github.com/interledger/interledger-app/go/log"
	"github.com/jmoiron/sqlx"
	"go.temporal.io/api/enums/v1"
	temporal_client "go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

const cols = ` id, wallet_id, platform, identifier, state, public, key_id, proof, signature, signature_hash, created_at, verified_at `

func List(ctx context.Context, b Backends, walletID string) ([]identities.Identity, error) {
	var res []identities.Identity
	err := b.DB().SelectContext(ctx, &res, fmt.Sprintf("SELECT %s FROM identities WHERE wallet_id=$1", cols), walletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return res, nil
}

func ListPublic(ctx context.Context, b Backends, walletID string) ([]identities.Identity, error) {
	var res []identities.Identity
	err := b.DB().SelectContext(ctx, &res, fmt.Sprintf("SELECT %s FROM identities WHERE wallet_id=$1 AND public=true AND state=$2", cols),
		walletID, identities.StateVerified)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return res, nil
}

func Add(ctx context.Context, b Backends, args identities.AddArgs) (*identities.Identity, error) {
	err := b.Validator().StructCtx(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInvalidArgument, err)
	}

	p, err := platforms.Get(b, args.Platform)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInvalidArgument, err)
	}

	var existing identities.Identity
	err = b.DB().GetContext(ctx, &existing, fmt.Sprintf("SELECT %s FROM identities WHERE platform=$1 AND lower(identifier)=$2", cols),
		args.Platform, strings.ToLower(args.Identifier))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}
	if existing.ID != "" {
		return nil, fmt.Errorf("%w %s identifier %s has already been created", identities.ErrAlreadyExists, args.Platform, args.Identifier)
	}

	id := uuid.NewString()
	c, err := p.GenerateSignedClaim(ctx, &platforms.SignedClaimArgs{
		Identifier: args.Identifier,
		WalletID:   args.WalletID,
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	err = p.GenerateImages(ctx, &platforms.GenerateImagesArgs{
		Identifier:    c.Claim.Identifier,
		SignatureHash: c.SignatureHash,
		WalletURL:     strings.TrimPrefix(c.Claim.Wallet, "https://"),
	})
	if err != nil {
		log.Error("error generating images", zap.Error(err))
	}

	ts := time.Unix(c.Claim.Ctime, 0)
	var identity identities.Identity
	err = b.DB().GetContext(ctx, &identity, "INSERT INTO identities(id, wallet_id, state, public, platform, key_id, identifier,proof, signature, signature_hash, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING "+cols,
		id, args.WalletID, identities.StateUnverified, true, args.Platform, c.Claim.Kid, args.Identifier, "", c.Signature, c.SignatureHash, ts)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	err = b.Notify().NotifyWallet(ctx, args.WalletID, notify.NotificationTypeIdentity)
	if err != nil {
		log.Error("error notifying wallet", zap.Error(err), zap.String("type", notify.NotificationTypeIdentity))
	}
	// TODO discuss with DEVOPS what will be the new admin url
	if args.Platform == identities.PlatformSlack {
		slack.SendToChannel(ctx, slack.ChannelSignupKYC, "wallet-info-bot", fmt.Sprintf(":troll: *New identity created*\n*Identifier:* %s\n*Platform:* %s\n*Wallet:* https://admin.interledger.tech/wallet/%s/profile", args.Identifier, args.Platform, args.WalletID))
	}

	return &identity, nil
}

func Get(ctx context.Context, b Backends, id string) (*identities.Identity, error) {
	var res identities.Identity
	err := b.DB().GetContext(ctx, &res, fmt.Sprintf("SELECT %s FROM identities WHERE id=$1", cols), id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w %s", identities.ErrNotFound, err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return &res, nil
}

// FIXME: Potentially remove
func VerifyInstructions(ctx context.Context, b Backends, id string) (*identities.VerifyInstructions, error) {
	ident, err := Get(ctx, b, id)
	if err != nil {
		return nil, err
	}

	p, err := platforms.Get(b, ident.Platform)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInvalidArgument, err)
	}

	verifyInstructions, err := p.VerifyInstructions(ctx, &platforms.VerifyInstructionsArgs{
		Identifier: ident.Identifier,
		Identity:   ident,
		WalletID:   ident.WalletID,
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return &identities.VerifyInstructions{
		IdentityID:   id,
		Code:         "",
		Instructions: verifyInstructions,
	}, nil
}

func Delete(ctx context.Context, b Backends, id, walletID string) error {
	res, err := b.DB().ExecContext(ctx, "DELETE FROM identities WHERE id=$1 AND wallet_id=$2", id, walletID)
	if err != nil {
		return fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	if rows != 1 {
		return fmt.Errorf("%w wrong number of rows deleted (%d)", identities.ErrInternal, rows)
	}

	err = b.Notify().NotifyWallet(ctx, walletID, notify.NotificationTypeIdentity)
	if err != nil {
		log.Error("error notifying wallet", zap.Error(err), zap.String("type", notify.NotificationTypeIdentity))
	}

	return err
}

func SetPublic(ctx context.Context, b Backends, id, walletID string, public bool) (*identities.Identity, error) {
	_, err := b.DB().ExecContext(ctx, "UPDATE identities SET public=$1, updated_at=now() WHERE id=$2 AND wallet_id=$3", public, id, walletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	err = b.Notify().NotifyWallet(ctx, walletID, notify.NotificationTypeIdentity)
	if err != nil {
		log.Error("error notifying wallet", zap.Error(err), zap.String("type", notify.NotificationTypeIdentity))
	}

	return Get(ctx, b, id)
}

func StartVerification(ctx context.Context, b Backends, id, proof string) (*identities.Identity, error) {
	ident, err := Get(ctx, b, id)
	if err != nil {
		return nil, err
	}

	p, err := platforms.Get(b, ident.Platform)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	workflowOptions := temporal_client.StartWorkflowOptions{
		ID:                       "identities_verify_" + id,
		TaskQueue:                "backend",
		WorkflowExecutionTimeout: time.Hour * 24, // Workflow has a day to complete
		WorkflowIDReusePolicy:    enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
	}

	_, err = b.Temporal().ExecuteWorkflow(ctx, workflowOptions, p.VerifyWorkflow(), id, proof)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return ident, nil
}

func UpdateState(ctx context.Context, b Backends, id string, state identities.State, proof string) error {
	ident, err := Get(ctx, b, id)
	if err != nil {
		return err
	}

	// Only update the verified at if the state is verified
	var verifiedAt time.Time
	if state == identities.StateVerified {
		verifiedAt = time.Now()
	}

	row := b.DB().QueryRowContext(ctx,
		`UPDATE identities 
    SET proof=$1, state=$2, updated_at=now(), verified_at=$3 
    WHERE id=$4 
    RETURNING wallet_id`,
		proof, state, verifiedAt, ident.ID)

	var walletID string
	err = row.Scan(&walletID)
	if err != nil {
		return fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	err = b.Notify().NotifyWallet(ctx, walletID, notify.NotificationTypeIdentity)
	if err != nil {
		log.Error("error notifying wallet", zap.Error(err), zap.String("type", notify.NotificationTypeIdentity))
	}

	err = b.Payments().SignalIdentityCreated(ctx, ident.Identifier)
	if err != nil {
		log.Error("error notifying payments of new identity", zap.Error(err), zap.String("id", id))
	}

	return nil
}

func GetBySignatureHash(ctx context.Context, b Backends, sigHash []byte) (*identities.Identity, error) {
	var res identities.Identity
	err := b.DB().GetContext(ctx, &res, fmt.Sprintf("SELECT %s FROM identities WHERE signature_hash=$1 and public=true", cols), sigHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w %s", identities.ErrNotFound, err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return &res, nil
}

func GetByIdentifier(ctx context.Context, b Backends, identifier string) (*identities.Identity, error) {
	var res identities.Identity
	err := b.DB().GetContext(ctx, &res, fmt.Sprintf("SELECT %s FROM identities WHERE lower(identifier) = lower($1) and public=true", cols), identifier)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w %s", identities.ErrNotFound, err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return &res, nil
}

func Search(ctx context.Context, b Backends, walletID, term string) ([]identities.SearchResult, error) {
	var dbRes, allCandidates, candidates []identities.SearchResult

	// Strip the URL prefix if it is a wallet address or a twitter account URL
	wa, err := wallets.ParseAddress(term)
	if err == nil {
		// Valid URL, possibly wallet address
		term = strings.TrimPrefix(wa.String(), env.OpenPaymentsURL()+"/")
		term = strings.TrimPrefix(term, "https://twitter.com/")
	}
	// Trim the twitter '@' prefix as we don't store it
	term = strings.TrimPrefix(term, "@")

	if len(term) < 3 {
		return dbRes, nil
	}

	// Lookup twitter and other external identities. (Not Slack)
	err = b.DB().SelectContext(ctx, &dbRes, `SELECT wallet_id, identifier, platform as identifier_type, similarity(identifier, $1) as rank
               FROM identities
               WHERE wallet_id<>$3 AND public = true AND state = 'verified' AND identifier ILIKE $2 AND platform<>$4 ORDER BY RANK DESC LIMIT 100`, term, "%"+term+"%", walletID, identities.PlatformSlack)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)

	}
	allCandidates = append(allCandidates, dbRes...)

	// Lookup Wallet addresses.
	dbRes = nil
	err = b.DB().SelectContext(ctx, &dbRes, `SELECT wallet_addresses.wallet_id, wallet_addresses.url as identifier, 'wallet_url' as identifier_type, coalesce(similarity(substring(wallet_addresses.url, $4), $1), 0) as rank
               FROM wallet_addresses  INNER JOIN wallet_kyc_status on  wallet_addresses.wallet_id=wallet_kyc_status.wallet_id
               WHERE wallet_addresses.wallet_id<>$3 AND wallet_addresses.url ILIKE $2 AND wallet_kyc_status.status in ($5,$6,$7) ORDER BY rank, wallet_addresses.wallet_id DESC LIMIT 100`, term, "%"+term+"%", walletID, len(env.OpenPaymentsURL())+1, kyc.StatusApproved, kyc.StatusLevel1, kyc.StatusLevel2)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}
	allCandidates = append(allCandidates, dbRes...)

	// Lookup Wallet names.
	dbRes = nil
	err = b.DB().SelectContext(ctx, &dbRes, `SELECT wallets.id as wallet_id,  wallets.name as identifier, 'wallet' as identifier_type, similarity(wallets.name, $1) as rank
               FROM wallets INNER JOIN wallet_kyc_status on wallets.id=wallet_kyc_status.wallet_id
               WHERE wallets.id<>$3 AND wallets.name ILIKE $2 AND wallet_kyc_status.status in ($4,$5,$6) ORDER BY rank,  wallets.id DESC LIMIT 100`, term, "%"+term+"%", walletID, kyc.StatusApproved, kyc.StatusLevel1, kyc.StatusLevel2)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}
	allCandidates = append(allCandidates, dbRes...)

	if len(allCandidates) == 0 {
		return allCandidates, nil
	}

	// Get the walletIDs
	allWalletIDs := make([]interface{}, len(allCandidates))
	for i, sr := range allCandidates {
		allWalletIDs[i] = sr.WalletID
	}

	// Lookup all the wallets that can receive payments from the list of results
	var canRecv []string
	canRecvQuery, canRecvArgs, err := sqlx.In("SELECT DISTINCT wallet_id FROM linked_accounts WHERE wallet_id IN (?) AND state=? AND can_receive=?", allWalletIDs, linkedaccounts.Verified, true)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}
	err = b.DB().SelectContext(ctx, &canRecv, b.DB().Rebind(canRecvQuery), canRecvArgs...)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	if len(canRecv) == 0 {
		return nil, nil
	}

	canRecvWalletIDs := make([]interface{}, len(canRecv))
	for i, wid := range canRecv {
		canRecvWalletIDs[i] = wid
	}

	walletCanRecv := make(map[string]bool)
	for _, cr := range canRecv {
		walletCanRecv[cr] = true
	}

	// Loop over all the possible results and only include the ones that can receive payment in the candidates
	for _, r := range allCandidates {
		if walletCanRecv[r.WalletID] {
			candidates = append(candidates, r)
		}
	}

	// Lookup URLs for the candidate walletIDs
	ppUrlQuery, ppUrlArgs, err := sqlx.In(`SELECT wallet_id, url FROM wallet_addresses WHERE wallet_id IN(?)`, canRecvWalletIDs)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}
	dbRes = nil
	err = b.DB().SelectContext(ctx, &dbRes, b.DB().Rebind(ppUrlQuery), ppUrlArgs...)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	walletURLs := make(map[string]string)
	for _, r := range dbRes {
		walletURLs[r.WalletID] = r.WalletUrl
	}

	// Lookup the Wallet names for walletIDs
	nameQuery, nameArgs, err := sqlx.In(`SELECT id as wallet_id, name FROM wallets WHERE id IN(?)`, canRecvWalletIDs)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}
	dbRes = nil
	err = b.DB().SelectContext(ctx, &dbRes, b.DB().Rebind(nameQuery), nameArgs...)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	walletNames := make(map[string]string)
	for _, r := range dbRes {
		walletNames[r.WalletID] = r.WalletName
	}

	// Set the wallet names and URLs
	for i, r := range candidates {
		candidates[i].WalletName = walletNames[r.WalletID]
		candidates[i].WalletUrl = walletURLs[r.WalletID]
	}

	// Group by duplicate wallet addresses
	mp := make(map[string][]identities.SearchResult)
	for _, r := range candidates {
		mp[r.WalletID] = append(mp[r.WalletID], r)
	}

	var resp []identities.SearchResult
	for wid, r := range mp {
		if len(r) == 0 {
			// Shouldn't happen but I'm just paranoid
			continue
		}

		group := identities.SearchResult{
			WalletID:       wid,
			WalletUrl:      r[0].WalletUrl,
			WalletName:     r[0].WalletName,
			IdentifierType: "wallet",
			Identifier:     r[0].WalletName,
		}
		for _, sr := range r {
			// Pick the highest possible rank of the sub results
			if sr.Rank > group.Rank {
				group.Rank = sr.Rank
			}
			// The wallet is already the grouping, don't include as one of the sub results
			if sr.IdentifierType == "wallet" {
				continue
			}
			group.SubResults = append(group.SubResults, sr)
		}
		resp = append(resp, group)
	}

	sort.Slice(resp, func(i, j int) bool {
		return resp[i].Rank < resp[j].Rank
	})

	if len(resp) > 20 {
		resp = resp[:20]
	}

	return resp, nil
}
