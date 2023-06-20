package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/tabapay"

	"gitlab.com/fynbos/backend/notify"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"

	"gitlab.com/fynbos/backend/linkedaccounts"

	"github.com/google/uuid"
)

const (
	allFields = "id, wallet_id, name, nickname, mask, provider, provider_id, type, can_send, can_receive, state, created_at, updated_at"

	// If you update this, then remember to update the create and createBatch functions.
	insertFields  = "id, wallet_id, name, nickname, mask, provider, provider_id, type, can_send, can_receive, state"
	insertColumns = 11
)

const (
	reviewAllFields    = "id, linked_account_id, state, new_state, reason, reviewed_by, created_at, completed_at"
	reviewInsertFields = "linked_account_id, state"
)

func Create(ctx context.Context, b Backends, args *linkedaccounts.CreateArgs) (*linkedaccounts.LinkedAccount, error) {
	err := b.Validator().Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInvalidArgument, err.Error())
	}

	linkedAccountID := args.ID
	if linkedAccountID == "" {
		linkedAccountID = uuid.NewString()
	}

	state := args.State
	if state == "" {
		state = linkedaccounts.Verified
	}

	var linkedAccount linkedaccounts.LinkedAccount
	err = b.DB().GetContext(
		ctx,
		&linkedAccount,
		fmt.Sprintf("INSERT INTO linked_accounts (%s) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING %s;", insertFields, allFields),
		linkedAccountID,
		args.WalletID,
		args.Name,
		args.Nickname,
		args.Mask,
		args.Provider,
		args.ProviderID,
		args.Type,
		args.CanSend,
		args.CanReceive,
		state,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err.Error())
	}

	err = b.Notify().NotifyWallet(ctx, args.WalletID, notify.NotificationTypeLinkedAccount)
	if err != nil {
		log.Error("notify failed for linked account", zap.String("walletId", args.WalletID), zap.Error(err))
	}

	return &linkedAccount, nil
}

func CreateBatch(ctx context.Context, b Backends, args []linkedaccounts.CreateArgs) ([]linkedaccounts.LinkedAccount, error) {
	if len(args) == 0 {
		return nil, nil
	}

	for _, arg := range args {
		err := b.Validator().Struct(arg)
		if err != nil {
			return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInvalidArgument, err.Error())
		}
	}

	var placeholders []string
	var values []interface{}
	for i, arg := range args {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", i*insertColumns+1, i*insertColumns+2, i*insertColumns+3, i*insertColumns+4, i*insertColumns+5, i*insertColumns+6, i*insertColumns+7, i*insertColumns+8, i*insertColumns+9, i*insertColumns+10, i*insertColumns+11))

		linkedAccountID := arg.ID
		if linkedAccountID == "" {
			linkedAccountID = uuid.NewString()
		}
		state := arg.State
		if state == "" {
			state = linkedaccounts.Verified
		}

		values = append(values, linkedAccountID, arg.WalletID, arg.Name, arg.Nickname, arg.Mask, arg.Provider, arg.ProviderID, arg.Type, arg.CanSend, arg.CanReceive, state)
	}

	var linkedAccounts []linkedaccounts.LinkedAccount
	err := b.DB().SelectContext(
		ctx,
		&linkedAccounts,
		fmt.Sprintf("INSERT INTO linked_accounts (%s) VALUES %s RETURNING %s;", insertFields, strings.Join(placeholders, ","), allFields),
		values...,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err.Error())
	}

	notifiedWallets := map[string]bool{}
	for _, la := range linkedAccounts {
		notified := notifiedWallets[la.WalletID]
		if notified {
			continue
		}

		err = b.Notify().NotifyWallet(ctx, la.WalletID, notify.NotificationTypeLinkedAccount)
		if err != nil {
			log.Error("notify failed for linked account", zap.String("walletId", la.WalletID), zap.Error(err))
		}

		notifiedWallets[la.WalletID] = true
	}

	return linkedAccounts, nil
}

func Get(ctx context.Context, b Backends, id string) (*linkedaccounts.LinkedAccount, error) {
	if id == "" {
		return nil, fmt.Errorf("%w ID is required.", linkedaccounts.ErrInvalidArgument)
	}

	var linkedAccount linkedaccounts.LinkedAccount
	err := b.DB().GetContext(
		ctx,
		&linkedAccount,
		fmt.Sprintf("SELECT %s FROM linked_accounts where id=$1 and deleted_at IS NULL LIMIT 1;", allFields),
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, linkedaccounts.ErrNotFound
		}

		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err.Error())
	}

	return &linkedAccount, nil
}

func Delete(ctx context.Context, b Backends, id string) error {
	_, err := b.DB().ExecContext(ctx, "UPDATE linked_accounts SET deleted_at=now() WHERE id=$1", id)
	if err != nil {
		return fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err.Error())
	}

	return nil
}

func MarkNotDeleted(ctx context.Context, b Backends, id string) (*linkedaccounts.LinkedAccount, error) {
	result, err := b.DB().ExecContext(ctx, "UPDATE linked_accounts SET deleted_at=NULL WHERE id=$1", id)
	if err != nil {
		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err.Error())
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err.Error())
	}

	if rows < 1 {
		return nil, linkedaccounts.ErrNotFound
	}

	return Get(ctx, b, id)
}

func GetByProviderID(ctx context.Context, b Backends, args linkedaccounts.GetByProviderIDArgs) (*linkedaccounts.LinkedAccount, error) {
	var linkedAccount linkedaccounts.LinkedAccount
	err := b.DB().GetContext(
		ctx,
		&linkedAccount,
		fmt.Sprintf("SELECT %s FROM linked_accounts WHERE provider=$1 AND provider_id=$2;", allFields),
		args.Provider,
		args.ProviderID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, linkedaccounts.ErrNotFound
		}

		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err.Error())
	}

	return &linkedAccount, nil
}

func ListByWalletId(ctx context.Context, b Backends, walletId string) ([]linkedaccounts.LinkedAccount, error) {

	var linkedAccounts []linkedaccounts.LinkedAccount
	err := b.DB().SelectContext(
		ctx,
		&linkedAccounts,
		fmt.Sprintf("SELECT %s FROM linked_accounts WHERE deleted_at IS NULL AND wallet_id=$1;", allFields),
		walletId,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, linkedaccounts.ErrNotFound
		}

		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err.Error())
	}

	return linkedAccounts, nil
}

func ListMXBankAccounts(ctx context.Context, b Backends) ([]linkedaccounts.LinkedAccount, error) {
	var linkedAccounts []linkedaccounts.LinkedAccount
	err := b.DB().SelectContext(
		ctx,
		&linkedAccounts,
		fmt.Sprintf("SELECT %s FROM linked_accounts WHERE deleted_at IS NULL AND provider=$1 AND type=$2;", allFields),
		mx.ProviderName, mx.TypeBankAccount,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err)
	}

	return linkedAccounts, nil
}

func SetNickname(ctx context.Context, b Backends, id, nickname string) error {
	result, err := b.DB().ExecContext(ctx, `UPDATE linked_accounts set nickname = $1 where id = $2`, nickname, id)
	if err != nil {
		return fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err.Error())
	}
	ra, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err.Error())
	}
	if ra == 0 {
		return linkedaccounts.ErrNotFound
	}

	return nil
}

func Requires3DS(ctx context.Context, b Backends, id string) (bool, error) {
	la, err := Get(ctx, b, id)
	if err != nil {
		return false, err
	}

	return la.Provider == tabapay.ProviderName && la.Type == tabapay.TypeCard, nil
}

type dbReview struct {
	ID              string
	LinkedAccountID string `db:"linked_account_id"`
	State           linkedaccounts.State
	NewState        linkedaccounts.State `db:"new_state"`
	Reason          string
	ReviewedBy      string       `db:"reviewed_by"`
	CreatedAt       time.Time    `db:"created_at"`
	CompletedAt     sql.NullTime `db:"completed_at"`
}

func toReview(record dbReview) linkedaccounts.Review {
	return linkedaccounts.Review{
		ID:              record.ID,
		LinkedAccountID: record.LinkedAccountID,
		State:           record.State,
		NewState:        record.NewState,
		Reason:          record.Reason,
		ReviewedBy:      record.ReviewedBy,
		CreatedAt:       record.CreatedAt,
		CompletedAt:     record.CompletedAt.Time,
	}
}

func CreateReviews(ctx context.Context, b Backends, reviewsArgs []linkedaccounts.CreateReviewArgs) ([]linkedaccounts.Review, error) {
	var placeholders []string
	var values []interface{}
	insertColumns := 2
	for i, review := range reviewsArgs {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", i*insertColumns+1, i*insertColumns+2))
		values = append(values, review.LinkedAccountID, review.State)
	}

	var dbReviews []dbReview
	err := b.DB().SelectContext(
		ctx,
		&dbReviews,
		fmt.Sprintf("INSERT INTO linked_account_reviews (%s) VALUES %s RETURNING %s;", reviewInsertFields, strings.Join(placeholders, ","), reviewAllFields),
		values...,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err)
	}

	reviews := make([]linkedaccounts.Review, len(dbReviews))
	for i, record := range dbReviews {
		reviews[i] = toReview(record)
	}

	return reviews, nil
}

func GetReview(ctx context.Context, b Backends, id string) (*linkedaccounts.Review, error) {
	var dbReview dbReview
	err := b.DB().GetContext(
		ctx,
		&dbReview,
		fmt.Sprintf("SELECT %s FROM linked_account_reviews WHERE id=$1;", reviewAllFields),
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err)
	}

	review := toReview(dbReview)
	return &review, nil
}

func UpdateReviewState(ctx context.Context, b Backends, reviewID string, newState linkedaccounts.State) (*linkedaccounts.Review, error) {
	var dbReview dbReview
	err := b.DB().GetContext(
		ctx,
		&dbReview,
		fmt.Sprintf("UPDATE linked_account_reviews SET new_state=$1 WHERE id=$2  RETURNING %s;", reviewAllFields),
		newState,
		reviewID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err)
	}

	review := toReview(dbReview)
	return &review, nil
}

func UpdateReviewReason(ctx context.Context, b Backends, reviewID string, reason string) (*linkedaccounts.Review, error) {
	var dbReview dbReview
	err := b.DB().GetContext(
		ctx,
		&dbReview,
		fmt.Sprintf("UPDATE linked_account_reviews SET reason=$1 WHERE id=$2  RETURNING %s;", reviewAllFields),
		reason,
		reviewID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err)
	}

	review := toReview(dbReview)
	return &review, nil
}

func CompleteReview(ctx context.Context, b Backends, reviewID string, reviewedBy string) (*linkedaccounts.Review, error) {
	var dbReview dbReview
	err := b.DB().GetContext(
		ctx,
		&dbReview,
		fmt.Sprintf("UPDATE linked_account_reviews SET reviewed_by=$1, completed_at=now() WHERE id=$2  RETURNING %s;", reviewAllFields),
		reviewedBy,
		reviewID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err)
	}

	result, err := b.DB().ExecContext(
		ctx,
		"UPDATE linked_accounts SET state=$1 WHERE id=$2;",
		dbReview.NewState,
		dbReview.LinkedAccountID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err)
	}
	if rows, _ := result.RowsAffected(); rows < 1 {
		return nil, fmt.Errorf("%w Failed to update linked account's new state after review.", linkedaccounts.ErrInternal)
	}

	review := toReview(dbReview)
	return &review, nil
}
