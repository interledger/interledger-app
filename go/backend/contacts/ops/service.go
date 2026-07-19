package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/interledger/interledger-app/go/backend/contacts"
	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/interledger/interledger-app/go/backend/wallets"
)

const contactCols = ` id, name, payment_pointer, wallet_id, last_paid_at `

func Create(ctx context.Context, b Backends, args contacts.CreateContactArgs) (*contacts.Contact, error) {
	err := b.Validator().Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", contacts.ErrInvalidArgument, err)
	}

	var c contacts.Contact
	err = b.DB().GetContext(ctx, &c,
		fmt.Sprintf("INSERT INTO contacts (name, payment_pointer, wallet_id) values ($1, $2, $3) returning %s", contactCols),
		args.Name, args.PaymentPointer, args.WalletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", contacts.ErrInternal, err)
	}

	return &c, nil
}

func List(ctx context.Context, b Backends, walletID string, page db.Pagination, orderBy string) ([]contacts.Contact, error) {
	if orderBy == "" {
		orderBy = "name asc"
	}

	ob, err := db.NewOrderBy(orderBy, []string{"name", "last_paid_at"}, "contacts")
	if err != nil {
		return nil, fmt.Errorf("%w %s", contacts.ErrInternal, err)
	}

	sqlStmt := fmt.Sprintf("SELECT %s FROM contacts WHERE wallet_id=$1 %s %s", contactCols, ob.SQLOrderBy(), page.SQL())
	sqlArgs := []interface{}{walletID}

	if page.PageToken != "" {
		sqlStmt = fmt.Sprintf("SELECT %s FROM contacts %s and wallet_id=$1 %s %s", contactCols, ob.SQLWhere("$2"), ob.SQLOrderBy(), page.SQL())
		sqlArgs = []interface{}{walletID, page.PageToken}
	}

	var cl []contacts.Contact
	err = b.DB().SelectContext(ctx, &cl, sqlStmt, sqlArgs...)
	if err != nil {
		return nil, fmt.Errorf("%w %s", contacts.ErrInternal, err)
	}

	return cl, nil
}

func Get(ctx context.Context, b Backends, walletID string, wa wallets.Address) (*contacts.Contact, error) {
	var c contacts.Contact
	err := b.DB().GetContext(ctx, &c,
		fmt.Sprintf("SELECT %s from contacts where wallet_id = $1 and payment_pointer = $2", contactCols),
		walletID, wa.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w %s", contacts.ErrNotFound, err)
		}
		return nil, fmt.Errorf("%w %s", contacts.ErrInternal, err)
	}

	return &c, nil
}

func SetLastPaidAtNow(ctx context.Context, b Backends, walletID string, wa wallets.Address) error {
	_, err := b.DB().ExecContext(ctx,
		"UPDATE contacts set last_paid_at = now()::TIMESTAMP where wallet_id = $1 and payment_pointer = $2",
		walletID, wa.String())
	if err != nil {
		return fmt.Errorf("%w %s", contacts.ErrInternal, err)
	}

	return nil
}
