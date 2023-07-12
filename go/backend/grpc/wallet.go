package grpc

import (
	"context"
	"errors"
	"sync"

	"gitlab.com/fynbos/backend/linkedaccounts"

	"gitlab.com/fynbos/backend/openpayments"

	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/paymentpointers"
	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/user"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) GetCurrentWallet(ctx context.Context, req *pb.Empty) (*pb.GetCurrentWalletResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, ForbiddenError("Unauthenticated.")
	}

	w, err := s.b.Users().WalletForContext(ctx)

	if w == nil {
		return nil, NotFoundError("wallet not found")
	}

	return &pb.GetCurrentWalletResponse{
		Id: w.ID,
	}, toGRPCError(err)
}

func (s *rpcService) SetWalletName(ctx context.Context, req *pb.SetWalletNameRequest) (*pb.Empty, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Users().WalletForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, ForbiddenError("Unauthenticated.")
	}

	err = s.b.Users().SetWalletName(ctx, w.ID, req.Name)

	return &pb.Empty{}, toGRPCError(err)
}

func (s *rpcService) GetPublicWalletDetails(ctx context.Context, req *pb.GetPublicWalletDetailsRequest) (*pb.GetPublicWalletDetailsResponse, error) {
	w, err := s.b.Users().GetWallet(ctx, req.GetId())

	return &pb.GetPublicWalletDetailsResponse{
		Id:         w.ID,
		PublicName: w.Name,
	}, toGRPCError(err)
}

func (s *rpcService) GetWalletInfo(ctx context.Context, _ *pb.Empty) (*pb.WalletInfo, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, ForbiddenError("Unauthenticated.")
	}

	w, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	var hasWalletAddress, hasCard, hasBank, hasIdentities, hasTxs bool
	var anyErr error
	var wg sync.WaitGroup
	var pp paymentpointers.PaymentPointer

	wg.Add(4)

	go func() {
		defer wg.Done()
		opp, err := s.b.OpenPayments().GetWalletPaymentPointer(ctx, w.ID)
		if errors.Is(err, openpayments.ErrPaymentPointerNotFound) {
			return
		}
		if err != nil {
			anyErr = err
			return
		}

		pp, err = paymentpointers.Parse(opp.URL)
		if err != nil {
			anyErr = err
			return
		}
		hasWalletAddress = true
	}()

	go func() {
		defer wg.Done()
		lal, err := s.b.LinkedAccounts().ListByWalletId(ctx, w.ID)
		if err != nil {
			anyErr = err
			return
		}

		for _, la := range lal {
			if la.Provider == mx.ProviderName {
				hasBank = true
			}
			if la.Provider == tabapay.ProviderName {
				hasCard = true
			}
		}
	}()

	go func() {
		defer wg.Done()
		ids, err := s.b.Identities().List(ctx, w.ID)
		if err != nil {
			anyErr = err
			return
		}
		hasIdentities = len(ids) > 0
	}()

	go func() {
		defer wg.Done()
		txs, err := s.b.Transactions().List(ctx, db.Pagination{PageSize: 1}, w.ID)
		if err != nil {
			anyErr = err
			return
		}
		hasTxs = len(txs) > 0
	}()

	wg.Wait()
	if anyErr != nil {
		return nil, toGRPCError(anyErr)
	}

	return &pb.WalletInfo{
		WalletID:         w.ID,
		Url:              pp.String(),
		FormattedURL:     pp.ShortString(),
		HasCard:          hasCard,
		HasBank:          hasBank,
		HasIdentities:    hasIdentities,
		HasTransacted:    hasTxs,
		HasWalletAddress: hasWalletAddress,
	}, nil
}

func (s *rpcService) SearchWallets(ctx context.Context, req *pb.SearchWalletsRequest) (*pb.SearchWalletsResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, ForbiddenError("Unauthenticated.")
	}

	w, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	results, err := s.b.Identities().Search(ctx, w.ID, req.Term)
	if err != nil {
		return nil, toGRPCError(err)
	}

	walletCanRecv := make(map[string]bool)
	// Deduplicate as we could have the same result more than once if multiple identities match for the same wallet
	for _, r := range results {
		walletCanRecv[r.WalletID] = false
	}

	var mutex sync.Mutex
	var wg sync.WaitGroup
	var anyErr error

	for wid := range walletCanRecv {
		// Can't send to yourself so don't bother checking, search results should exclude it but let us be paranoid.
		if wid == w.ID {
			continue
		}
		wg.Add(1)
		go func(walletID string) {
			defer wg.Done()
			accounts, err := s.b.LinkedAccounts().ListByWalletId(ctx, walletID)
			if err != nil && !errors.Is(err, linkedaccounts.ErrNotFound) {
				anyErr = err
				return
			}

			for _, acc := range accounts {
				if acc.CanReceive {
					mutex.Lock()
					defer mutex.Unlock()
					walletCanRecv[walletID] = true
					return
				}
			}
		}(wid)
	}

	wg.Wait()

	if anyErr != nil {
		return nil, toGRPCError(err)
	}

	res := make([]*pb.SearchResult, len(results))
	for i, r := range results {
		res[i] = &pb.SearchResult{
			WalletID:       r.WalletID,
			Identifier:     r.Identifier,
			IdentifierType: r.IdentifierType,
			CanSend:        walletCanRecv[r.WalletID],
		}
	}

	return &pb.SearchWalletsResponse{Results: res}, nil
}
