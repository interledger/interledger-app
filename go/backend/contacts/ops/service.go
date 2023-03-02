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

type dbContact struct {
	ID             string
	Name           string
	PaymentPointer string `db:"payment_pointer"`
	WalletID       string `db:"wallet_id"`
}

func Create(ctx context.Context, b Backends, args contacts.CreateContactArgs) (*contacts.Contact, error) {
	err := b.Validator().Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", contacts.ErrInvalidArgument, err)
	}

	var c dbContact
	err = b.DB().GetContext(ctx, &c,
		fmt.Sprintf("INSERT INTO contacts (name, payment_pointer, wallet_id) values ($1, $2, $3) returning %s", contactCols),
		args.Name, args.PaymentPointer.String(), args.WalletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", contacts.ErrInternal, err)
	}

	pp, err := paymentpointers.Parse(c.PaymentPointer)
	if err != nil {
		return nil, fmt.Errorf("%w %s error parsing payment pointer", contacts.ErrInternal, err)
	}

	return &contacts.Contact{
		ID:             c.ID,
		Name:           c.Name,
		PaymentPointer: pp,
		WalletID:       c.WalletID,
	}, nil
}

func List(ctx context.Context, b Backends, walletID string, page db.Pagination) ([]contacts.Contact, error) {
	sqlStmt := fmt.Sprintf("SELECT %s FROM contacts WHERE wallet_id=$1 ORDER BY name %s", contactCols, page.SQL())
	sqlArgs := []interface{}{walletID}

	if page.PageToken != "" {
		sqlStmt = fmt.Sprintf("SELECT %s FROM contacts WHERE wallet_id=$1 and id > $2 ORDER BY name %s", contactCols, page.SQL())
		sqlArgs = []interface{}{walletID, page.PageToken}
	}

	var cl []dbContact
	err := b.DB().SelectContext(ctx, &cl, sqlStmt, sqlArgs...)
	if err != nil {
		return nil, fmt.Errorf("%w %s", contacts.ErrInternal, err)
	}

	var contactList []contacts.Contact
	for _, c := range cl {

		pp, err := paymentpointers.Parse(c.PaymentPointer)
		if err != nil {
			return nil, fmt.Errorf("%w %s error parsing payment pointer", contacts.ErrInternal, err)
		}

		contactList = append(contactList, contacts.Contact{
			ID:             c.ID,
			Name:           c.Name,
			PaymentPointer: pp,
			WalletID:       c.WalletID,
		})
	}

	return contactList, nil
}

func Get(ctx context.Context, b Backends, walletID string, pp *paymentpointers.PaymentPointer) (*contacts.Contact, error) {
	var c dbContact
	err := b.DB().GetContext(ctx, &c,
		fmt.Sprintf("SELECT %s from contacts where wallet_id = $1 and payment_pointer = $2", contactCols),
		walletID, pp.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w %s", contacts.ErrNotFound, err)
		}
		return nil, fmt.Errorf("%w %s", contacts.ErrInternal, err)
	}

	ppDb, err := paymentpointers.Parse(c.PaymentPointer)
	if err != nil {
		return nil, fmt.Errorf("%w %s error parsing payment pointer", contacts.ErrInternal, err)
	}

	return &contacts.Contact{
		ID:             c.ID,
		Name:           c.Name,
		PaymentPointer: ppDb,
		WalletID:       c.WalletID,
	}, nil
}
