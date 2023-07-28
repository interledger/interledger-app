package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gitlab.com/fynbos/backend/slack"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/db"
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
	reviewAllFields      = "id, linked_account_id, state, new_state, reason, reviewed_by, created_at, completed_at"
	reviewDetailedFields = "lar.id, linked_account_id, lar.state, new_state, reason, reviewed_by, lar.created_at, completed_at, mask, wallet_id"
	reviewInsertFields   = "linked_account_id, state"
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
		fmt.Sprintf("SELECT %s FROM linked_accounts WHERE provider=$1 AND provider_id=$2 AND wallet_id=$3;", allFields),
		args.Provider,
		args.ProviderID,
		args.WalletID,
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
	ID                    string `db:"id"`
	LinkedAccountID       string `db:"linked_account_id"`
	State                 linkedaccounts.State
	LinkedAccountMask     string               `db:"mask"`
	LinkedAccountWalletID string               `db:"wallet_id"`
	NewState              linkedaccounts.State `db:"new_state"`
	Reason                string
	ReviewedBy            string       `db:"reviewed_by"`
	CreatedAt             time.Time    `db:"created_at"`
	CompletedAt           sql.NullTime `db:"completed_at"`
}

func toReview(ctx context.Context, b Backends, record dbReview) (*linkedaccounts.Review, error) {
	var walletName string
	if record.LinkedAccountWalletID != "" {
		wallet, err := b.Wallets().Get(ctx, record.LinkedAccountWalletID)
		if err != nil {
			return nil, err
		}

		walletName = wallet.Name
	}

	return &linkedaccounts.Review{
		ID:                record.ID,
		LinkedAccountID:   record.LinkedAccountID,
		State:             record.State,
		NewState:          record.NewState,
		Reason:            record.Reason,
		LinkedAccountMask: record.LinkedAccountMask,
		WalletID:          record.LinkedAccountWalletID,
		WalletName:        walletName,
		ReviewedBy:        record.ReviewedBy,
		CreatedAt:         record.CreatedAt,
		CompletedAt:       record.CompletedAt.Time,
	}, nil
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
		review, err := toReview(ctx, b, record)
		if err != nil {
			return nil, err
		}

		reviews[i] = *review
	}

	for _, review := range reviews {
		slack.SendToChannel(ctx, slack.NotifyCard, "FynBOT", fmt.Sprintf("Review ID: %s\nLinked Account ID: %s\nState: %s\nReason: %s\n", review.ID, review.LinkedAccountID, review.State, review.Reason))
	}

	return reviews, nil
}

func GetReview(ctx context.Context, b Backends, id string) (*linkedaccounts.Review, error) {
	var dbReview dbReview
	err := b.DB().GetContext(
		ctx,
		&dbReview,
		fmt.Sprintf("SELECT %s FROM linked_account_reviews AS lar INNER JOIN linked_accounts AS la ON lar.linked_account_id = la.id WHERE lar.id=$1;", reviewDetailedFields),
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err)
	}

	review, err := toReview(ctx, b, dbReview)
	if err != nil {
		return nil, err
	}

	return review, nil
}

func ListReviews(ctx context.Context, b Backends, pagination db.Pagination) ([]linkedaccounts.Review, error) {
	pageSize := pagination.PageSize
	if pageSize < 1 || pageSize > 50 {
		pageSize = 50
	}

	query := fmt.Sprintf("SELECT %s FROM linked_account_reviews AS lar INNER JOIN linked_accounts AS la ON lar.linked_account_id = la.id WHERE lar.created_at < (SELECT created_at FROM linked_account_reviews WHERE id=$1) OR (lar.created_at = (SELECT created_at FROM linked_account_reviews WHERE id=$1) AND lar.id < $1) ORDER BY lar.created_at desc LIMIT $2;", reviewDetailedFields)
	args := []interface{}{pagination.PageToken, pageSize}
	if pagination.PageToken == "" {
		query = fmt.Sprintf("SELECT %s FROM linked_account_reviews AS lar INNER JOIN linked_accounts AS la ON lar.linked_account_id = la.id ORDER BY lar.created_at desc LIMIT $1;", reviewDetailedFields)
		args = []interface{}{pageSize}
	}

	var dbReviews []dbReview
	err := b.DB().SelectContext(ctx, &dbReviews, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err)
	}

	reviews := make([]linkedaccounts.Review, len(dbReviews))
	for i, record := range dbReviews {
		review, err := toReview(ctx, b, record)
		if err != nil {
			return nil, err
		}

		reviews[i] = *review
	}

	return reviews, nil
}

func ListIncompleteReviews(ctx context.Context, b Backends, pagination db.Pagination) ([]linkedaccounts.Review, error) {
	pageSize := pagination.PageSize
	if pageSize < 1 || pageSize > 50 {
		pageSize = 50
	}

	query := fmt.Sprintf("SELECT %s FROM linked_account_reviews AS lar INNER JOIN linked_accounts AS la ON lar.linked_account_id = la.id WHERE (lar.created_at < (SELECT lar.created_at FROM linked_account_reviews WHERE lar.id=$1) OR (lar.created_at = (SELECT lar.created_at FROM linked_account_reviews WHERE id=$1) AND id < $1)) AND lar.completed_at is null ORDER BY lar.created_at desc LIMIT $2;", reviewDetailedFields)
	args := []interface{}{pagination.PageToken, pageSize}
	if pagination.PageToken == "" {
		query = fmt.Sprintf("SELECT %s FROM linked_account_reviews AS lar INNER JOIN linked_accounts AS la ON lar.linked_account_id = la.id WHERE completed_at is NULL ORDER BY lar.created_at desc LIMIT $1;", reviewDetailedFields)
		args = []interface{}{pageSize}
	}

	var dbReviews []dbReview
	err := b.DB().SelectContext(ctx, &dbReviews, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err)
	}

	reviews := make([]linkedaccounts.Review, len(dbReviews))
	for i, record := range dbReviews {
		review, err := toReview(ctx, b, record)
		if err != nil {
			return nil, err
		}

		reviews[i] = *review
	}

	return reviews, nil
}

func CompleteReview(ctx context.Context, b Backends, args linkedaccounts.CompleteReviewArgs) (*linkedaccounts.Review, error) {
	var old dbReview
	err := b.DB().GetContext(ctx, &old, fmt.Sprintf("SELECT %s FROM linked_account_reviews WHERE id=$1", reviewAllFields), args.ID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err)
	}

	reason := old.Reason
	if args.Reason != "" {
		reason = args.Reason
	}

	newState := old.NewState
	if args.NewState != "" {
		newState = args.NewState
	}

	reviewedBy := old.ReviewedBy
	if args.ReviewedBy != "" {
		reviewedBy = args.ReviewedBy
	}

	var dbReview dbReview
	err = b.DB().GetContext(
		ctx,
		&dbReview,
		fmt.Sprintf("UPDATE linked_account_reviews SET reviewed_by=$1, reason=$2, new_state=$3, completed_at=now() WHERE id=$4  RETURNING %s;", reviewAllFields),
		reviewedBy, reason, newState, args.ID,
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

	review, err := toReview(ctx, b, dbReview)
	if err != nil {
		return nil, err
	}
	return review, nil
}

func ListByProviderID(ctx context.Context, b Backends, provider, providerID string) ([]linkedaccounts.LinkedAccount, error) {
	var linkedAccounts []linkedaccounts.LinkedAccount
	err := b.DB().SelectContext(
		ctx,
		&linkedAccounts,
		fmt.Sprintf("SELECT %s FROM linked_accounts WHERE deleted_at IS NULL AND provider=$1 AND provider_id=$2;", allFields),
		provider, providerID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err)
	}

	return linkedAccounts, nil
}
