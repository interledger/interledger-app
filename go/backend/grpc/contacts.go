package grpc

import (
	"context"

	"github.com/interledger/interledger-app/go/backend/wallets"

	"github.com/interledger/interledger-app/go/backend/contacts"
	"github.com/interledger/interledger-app/go/backend/db"
	backendv1 "github.com/interledger/interledger-app/go/proto/backend/v1"
)

func (s *rpcService) CreateContact(
	ctx context.Context,
	req *backendv1.CreateContactRequest,
) (*backendv1.Contact, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wa, err := wallets.ParseAddress(req.GetPaymentPointer())
	if err != nil {
		return nil, toGRPCError(err)
	}

	w, err := s.b.Wallets().GetFromAddress(ctx, wa.String())
	if err != nil {
		return nil, toGRPCError(err)
	}

	c, err := s.b.Contacts().Create(ctx, contacts.CreateContactArgs{
		Name:           w.Name,
		PaymentPointer: wa,
		WalletID:       wallet.ID,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.Contact{
		Id:             c.ID,
		PaymentPointer: c.PaymentPointer.ShortString(),
		Name:           c.Name,
		WalletId:       c.WalletID,
	}, nil
}

func (s *rpcService) ListContacts(
	ctx context.Context,
	req *backendv1.ListContactsRequest,
) (*backendv1.ListContactsResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	pageToken := req.GetPageToken()
	page := db.PaginationFromPB(&backendv1.PaginationRequest{
		PageSize:  req.GetPageSize(),
		PageToken: &pageToken,
	})

	c, err := s.b.Contacts().List(ctx, wallet.ID, page, req.GetOrderBy())
	if err != nil {
		return nil, toGRPCError(err)
	}

	var nextPageToken string
	var res []*backendv1.Contact

	for i, contact := range c {

		if i == page.PageSize {
			// Use the PageSize+1 tx.ID as the start of the next page.
			nextPageToken = contact.ID
			break
		}

		res = append(res, &backendv1.Contact{
			Id:             contact.ID,
			PaymentPointer: contact.PaymentPointer.ShortString(),
			Name:           contact.Name,
			WalletId:       contact.WalletID,
		})
	}

	return &backendv1.ListContactsResponse{
		Contacts:      res,
		NextPageToken: nextPageToken,
	}, nil
}
