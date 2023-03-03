package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gitlab.com/fynbos/backend/contacts"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/paymentpointers"
)

const contactCols = ` id, name, payment_pointer, wallet_id `

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

func List(ctx context.Context, b Backends, walletID string, page db.Pagination) ([]contacts.Contact, error) {
	sqlStmt := fmt.Sprintf("SELECT %s FROM contacts WHERE wallet_id=$1 ORDER BY name %s", contactCols, page.SQL())
	sqlArgs := []interface{}{walletID}

	if page.PageToken != "" {
		sqlStmt = fmt.Sprintf("SELECT %s FROM contacts WHERE wallet_id=$1 and id > $2 ORDER BY name %s", contactCols, page.SQL())
		sqlArgs = []interface{}{walletID, page.PageToken}
	}

	var cl []contacts.Contact
	err := b.DB().SelectContext(ctx, &cl, sqlStmt, sqlArgs...)
	if err != nil {
		return nil, fmt.Errorf("%w %s", contacts.ErrInternal, err)
	}

	return cl, nil
}

func Get(ctx context.Context, b Backends, walletID string, pp paymentpointers.PaymentPointer) (*contacts.Contact, error) {
	var c contacts.Contact
	err := b.DB().GetContext(ctx, &c,
		fmt.Sprintf("SELECT %s from contacts where wallet_id = $1 and payment_pointer = $2", contactCols),
		walletID, pp.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w %s", contacts.ErrNotFound, err)
		}
		return nil, fmt.Errorf("%w %s", contacts.ErrInternal, err)
	}

	return &c, nil
}
