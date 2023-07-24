package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/paymentpointers"

	"gitlab.com/fynbos/env"

	"gitlab.com/fynbos/log"
	"go.temporal.io/api/enums/v1"
	"go.uber.org/zap"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/identities/platforms"
	temporal_client "go.temporal.io/sdk/client"
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
		Identifier:    "@" + c.Claim.Identifier,
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

	return err
}

func SetPublic(ctx context.Context, b Backends, id, walletID string, public bool) (*identities.Identity, error) {
	_, err := b.DB().ExecContext(ctx, "UPDATE identities SET public=$1, updated_at=now() WHERE id=$2 AND wallet_id=$3", public, id, walletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
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

	_, err = b.DB().ExecContext(ctx, "UPDATE identities SET proof=$1, state=$2, updated_at=now(), verified_at=$3 WHERE id=$4", proof, state, verifiedAt, ident.ID)
	if err != nil {
		return fmt.Errorf("%w %s", identities.ErrInternal, err)
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
	var dbRes []identities.SearchResult

	// Strip the URL prefix if it is a wallet address or a twitter account URL
	wa, err := paymentpointers.Parse(term)
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

	err = b.DB().SelectContext(ctx, &dbRes, `SELECT tmp.*, pp.url, w.name FROM (SELECT wallet_id, identifier, platform as identifier_type, similarity(identifier, $1) as rank
               FROM identities
               WHERE public = true AND state = 'verified' AND identifier ILIKE $2
               UNION
               SELECT wallet_id, url as identifier, 'wallet' as identifier_type, coalesce(similarity(substring(url, $3), $1), 0) as rank
               FROM payment_pointers
               WHERE url ILIKE $2
               UNION 
               SELECT id as wallet_id, name as identifier, 'wallet' as identifier_type, similarity(name, $1) as rank
               FROM wallets
               WHERE name ILIKE $2) tmp 
         INNER JOIN payment_pointers pp ON tmp.wallet_id = pp.wallet_id
         INNER JOIN wallets w ON w.id = tmp.wallet_id
         WHERE tmp.wallet_id<>$4 ORDER BY tmp.rank DESC LIMIT 20`, term, "%"+term+"%", len(env.OpenPaymentsURL())+1, walletID)

	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	// Group by duplicate wallet addresses
	mp := make(map[string][]identities.SearchResult)
	for _, r := range dbRes {
		mp[r.WalletID] = append(mp[r.WalletID], r)
	}

	var res []identities.SearchResult
	for wid, r := range mp {
		if len(r) == 0 {
			// Shouldn't happen but I'm just paranoid
			continue
		}
		if len(r) == 1 {
			res = append(res, r[0])
			continue
		}

		group := identities.SearchResult{
			WalletID:       wid,
			WalletUrl:      r[0].WalletUrl,
			WalletName:     r[0].WalletName,
			IdentifierType: "wallet",
		}
		for _, sr := range r {
			// Pick the highest possible rank of the sub results
			if sr.Rank > group.Rank {
				group.Rank = sr.Rank
			}
			// The wallet is already the grouping, don't include as one of the sub results
			if sr.IdentifierType == "wallet" {
				group.Identifier = sr.Identifier
				continue
			}
			group.SubResults = append(group.SubResults, sr)
		}
		res = append(res, group)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Rank < res[j].Rank
	})

	return res, nil
}
