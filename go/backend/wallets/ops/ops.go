package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/interledger/interledger-app/go/backend/country"
	"github.com/interledger/interledger-app/go/backend/db"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/interledger/interledger-app/go/backend/wallets"
)

// A user may have many wallets but only one default
// (partial unique index on user_wallets(user_id) WHERE is_default),
// so ON CONFLICT targets that partial index.
func Create(ctx context.Context, b Backends, args wallets.CreateArgs) (*wallets.Wallet, error) {
	err := b.Validator().StructCtx(ctx, args)
	if err != nil {
		return nil, err
	}

	userID := args.UserID
	name := args.Name
	if name == "" {
		name = "default"
	}

	walletID := args.ID
	if walletID == "" {
		walletID = uuid.NewString()
	}

	ctry := args.Country
	if ctry.String() == "" {
		ctry = country.US
	}

	err = crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(ctx, "INSERT INTO wallets (id, name, country) VALUES ($1, $2, $3)", walletID, name, ctry)
		if err != nil {
			return fmt.Errorf("%w %s", wallets.ErrInternal, err)
		}

		// claim the user's default-wallet slot
		res, err := tx.ExecContext(ctx, "INSERT INTO user_wallets (user_id, wallet_id, is_default) VALUES ($1, $2, true) ON CONFLICT (user_id) WHERE is_default DO NOTHING", userID, walletID)
		if err != nil {
			return fmt.Errorf("%w %s", wallets.ErrInternal, err)
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("%w %s", wallets.ErrInternal, err)
		}
		if rows == 0 {
			// if 0 rows are affected the user already has a default wallet,
			// so we roll back (no orphan wallets row) and return ErrDuplicateWallet for the caller to handle
			return wallets.ErrDuplicateWallet
		}

		for _, wa := range args.Addresses {
			_, err = tx.ExecContext(ctx, "INSERT INTO wallet_addresses(wallet_id, url) VALUES ($1, $2)", walletID, wa)
			if err != nil {
				return fmt.Errorf("%w %s", wallets.ErrInternal, err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	b.Analytics().TrackWalletCreated(walletID, userID)
	for range args.Addresses {
		b.Analytics().TrackWalletPaymentPointerCreated(walletID)
	}

	return Get(ctx, b, walletID)
}

func AddAddress(ctx context.Context, b Backends, id, url string) (*wallets.Wallet, error) {
	// Check if it already exists
	w, err := GetFromAddress(ctx, b, url)
	if err != nil && !errors.Is(err, wallets.ErrNoWalletFound) {
		return nil, err
	}

	if w != nil && w.ID == id {
		// The address already belongs to this wallet, nothing to see here.
		return w, nil
	}

	address, err := wallets.ParseAddress(url)
	if err != nil {
		return nil, err
	}
	_, err = b.DB().ExecContext(ctx, "INSERT INTO wallet_addresses(wallet_id, url) VALUES ($1, $2)", id, address)
	if db.IsErrorCode(err, db.UniqueViolationError) {
		return nil, fmt.Errorf("%w %s", wallets.ErrAddressExists, err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", wallets.ErrInternal, err)
	}

	b.Analytics().TrackWalletPaymentPointerCreated(id)

	return Get(ctx, b, id)
}

func GetFromAddress(ctx context.Context, b Backends, url string) (*wallets.Wallet, error) {
	address, err := wallets.ParseAddress(url)
	if err != nil {
		return nil, fmt.Errorf("%w address(%s) invalid", wallets.ErrNoWalletFound, url)
	}

	var wid string
	err = b.DB().GetContext(ctx, &wid, "SELECT wallet_id FROM wallet_addresses WHERE lower(url)=lower($1)", address)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w address(%s) not found", wallets.ErrNoWalletFound, address.String())
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", wallets.ErrInternal, err)
	}

	return Get(ctx, b, wid)
}

func Get(ctx context.Context, b Backends, id string) (*wallets.Wallet, error) {
	var wallet wallets.Wallet
	err := b.DB().GetContext(ctx, &wallet,
		"SELECT id, name, country, exceeded_limits FROM wallets WHERE id=$1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, wallets.ErrNoWalletFound
	}
	if err != nil {
		return nil, err
	}

	var wa []wallets.Address
	err = b.DB().SelectContext(ctx, &wa, "SELECT url FROM wallet_addresses WHERE wallet_id=$1", id)
	if err != nil {
		return nil, err
	}

	wallet.Addresses = wa

	return &wallet, nil
}

func List(ctx context.Context, b Backends, userID string) ([]wallets.Wallet, error) {
	var wl []wallets.Wallet
	err := b.DB().SelectContext(ctx, &wl, "SELECT w.id, w.name, w.country, w.exceeded_limits FROM wallets w INNER JOIN user_wallets uw ON w.id = uw.wallet_id WHERE user_id=$1", userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	for i, w := range wl {
		var wa []wallets.Address
		err = b.DB().SelectContext(ctx, &wa, "SELECT url FROM wallet_addresses WHERE wallet_id=$1", w.ID)
		if err != nil {
			return nil, err
		}

		wl[i].Addresses = wa
	}

	return wl, nil
}

func ListAll(ctx context.Context, b Backends, page db.Pagination) ([]wallets.Wallet, error) {
	args := map[string]any{}
	conditions := []string{}

	if page.PageToken != "" {
		conditions = append(conditions, "(created_at, id) < ( select created_at, id from wallets where id = :pagetoken )")
		args["pagetoken"] = page.PageToken
	}

	if page.Search != "" {
		args["search"] = "%" + page.Search + "%"
		if _, err := uuid.Parse(page.Search); err == nil {
			// Exact UUID match hits the primary-key index; also search by name.
			conditions = append(conditions, "(id = :searchid OR name ILIKE :search)")
			args["searchid"] = page.Search
		} else {
			conditions = append(conditions, "name ILIKE :search")
		}
	}

	filter := page.Filter

	if filter.FirstName != "" || filter.LastName != "" {
		// Single EXISTS so both names (when both are set) must match the SAME
		// individual_kyc_details row
		nameConds := []string{}
		if filter.FirstName != "" {
			nameConds = append(nameConds, "lower(k.first_name) LIKE :firstname")
			args["firstname"] = escapeLike(strings.ToLower(filter.FirstName)) + "%"
		}
		if filter.LastName != "" {
			nameConds = append(nameConds, "lower(k.last_name) LIKE :lastname")
			args["lastname"] = escapeLike(strings.ToLower(filter.LastName)) + "%"
		}
		conditions = append(conditions, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM individual_kyc_details k WHERE k.wallet_id = wallets.id AND %s)",
			strings.Join(nameConds, " AND "),
		))
	}

	if filter.WalletAddress != "" {
		conditions = append(conditions, "EXISTS (SELECT 1 FROM wallet_addresses a WHERE a.wallet_id = wallets.id AND lower(a.url) LIKE :walletaddress)")
		// Use partial search instead of the prefix so you won't need to complete with the host, like https://local.ilp.link/<wallet_address>
		args["walletaddress"] = "%" + escapeLike(strings.ToLower(filter.WalletAddress)) + "%"
	}

	if filter.ProviderID != "" {
		conditions = append(conditions, "EXISTS (SELECT 1 FROM linked_accounts la WHERE la.wallet_id = wallets.id AND la.provider_id = :providerid)")
		args["providerid"] = filter.ProviderID
	}

	if filter.WalletIDs != nil {
		conditions = append(conditions, "id = ANY(:walletids)")
		args["walletids"] = pq.Array(filter.WalletIDs)
	}

	query := "select id, name, country, exceeded_limits from wallets"
	if len(conditions) > 0 {
		query += " where " + strings.Join(conditions, " AND ")
	}
	query += " order by created_at DESC, id DESC %s"
	finalQuery := fmt.Sprintf(query, page.SQL())

	nstmt, err := b.DB().PrepareNamed(finalQuery)
	if err != nil {
		return nil, err
	}

	var wl []wallets.Wallet
	err = nstmt.SelectContext(ctx, &wl, args)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	for i, w := range wl {
		var wa []wallets.Address
		err = b.DB().SelectContext(ctx, &wa, "SELECT url FROM wallet_addresses WHERE wallet_id=$1", w.ID)
		if err != nil {
			return nil, err
		}

		wl[i].Addresses = wa
	}

	return wl, nil
}

// escapeLike escapes the LIKE/ILIKE metacharacters backslash, % and _ (in
// that order) so a user-supplied search term is matched literally.
func escapeLike(s string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(s)
}

func SetWalletName(ctx context.Context, b Backends, id, name string) (*wallets.Wallet, error) {
	// We convert the public name to lowercase as well, to match the wallet address name.
	_, err := b.DB().ExecContext(ctx, "UPDATE wallets set name = $1 where id = $2", strings.ToLower(name), id)
	if err != nil {
		return nil, err
	}

	return Get(ctx, b, id)
}

func SetCountry(ctx context.Context, b Backends, id string, country country.Country) (*wallets.Wallet, error) {
	result, err := b.DB().ExecContext(ctx, "UPDATE wallets set country=$1 WHERE id=$2;", country.String(), id)
	if err != nil {
		return nil, fmt.Errorf("%w %s", wallets.ErrInternal, err)
	}
	if rows, _ := result.RowsAffected(); rows < 1 {
		return nil, wallets.ErrNoWalletFound
	}

	return Get(ctx, b, id)
}

func SetExceededLimits(ctx context.Context, b Backends, id string, exceeded bool) (*wallets.Wallet, error) {
	_, err := b.DB().ExecContext(ctx, "UPDATE wallets SET exceeded_limits=$1 WHERE id=$2", exceeded, id)
	if err != nil {
		return nil, fmt.Errorf("%w %s", wallets.ErrInternal, err)
	}

	return Get(ctx, b, id)
}

func WalletForContext(ctx context.Context) (*wallets.Wallet, error) {
	w, ok := ctx.Value(wallets.CtxKey).(*wallets.Wallet)
	if !ok || w == nil {
		return nil, wallets.ErrNoWalletFound
	}
	return w, nil
}
