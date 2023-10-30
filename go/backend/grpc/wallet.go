package grpc

import (
	"context"
	"errors"
	"strings"
	"sync"

	"gitlab.com/fynbos/backend/db"
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

	w, err := s.b.Wallets().ForContext(ctx)

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

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, ForbiddenError("Unauthenticated.")
	}

	_, err = s.b.Wallets().SetWalletName(ctx, w.ID, req.Name)

	return &pb.Empty{}, toGRPCError(err)
}

func (s *rpcService) GetPublicWalletDetails(ctx context.Context, req *pb.GetPublicWalletDetailsRequest) (*pb.GetPublicWalletDetailsResponse, error) {
	w, err := s.b.Wallets().Get(ctx, req.GetId())

	return &pb.GetPublicWalletDetailsResponse{
		Id:         w.ID,
		PublicName: w.Name,
	}, toGRPCError(err)
}

func (s *rpcService) GetPublicWalletInfo(ctx context.Context, req *pb.GetPublicWalletInfoRequest) (*pb.PublicWalletInfo, error) {
	wallet, err := s.b.Wallets().GetFromAddress(ctx, req.WalletAddress)
	if err != nil {
		return nil, toGRPCError(err)
	}

	ids, err := s.b.Identities().ListPublic(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	idsResp := make([]*pb.Identity, len(ids))
	for i, id := range ids {
		idsResp[i] = identityToPB(&id, wallet.AddressString())
	}

	lal, err := s.b.LinkedAccounts().ListByWalletId(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	var canRecv bool
	for _, la := range lal {
		if la.CanReceive && !la.DeletedAt.Valid {
			canRecv = true
			break
		}
	}

	return &pb.PublicWalletInfo{
		WalletID:     wallet.ID,
		Address:      wallet.AddressString(),
		ShortAddress: wallet.AddressShortString(),
		Identities:   idsResp,
		CanReceive:   canRecv,
		PublicName:   wallet.Name,
	}, nil
}

func (s *rpcService) GetWalletInfo(ctx context.Context, _ *pb.Empty) (*pb.WalletInfo, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	var hasWalletAddress, hasCard, hasBank, hasIdentities, hasTxs bool
	var anyErr error
	var wg sync.WaitGroup

	wg.Add(3)

	if len(w.Addresses) > 0 {
		hasWalletAddress = true
	}

	go func() {
		defer wg.Done()
		lal, err := s.b.LinkedAccounts().ListByWalletId(ctx, w.ID)
		if err != nil {
			anyErr = err
			return
		}

		for _, la := range lal {
			if la.DeletedAt.Valid {
				continue
			}

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
		Url:              w.AddressString(),
		FormattedURL:     w.AddressShortString(),
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
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	term := req.GetTerm()
	term = strings.TrimSpace(term)

	results, err := s.b.Identities().Search(ctx, w.ID, term)
	if err != nil {
		return nil, toGRPCError(err)
	}

	res := make([]*pb.SearchResult, len(results))
	for i, r := range results {
		res[i] = &pb.SearchResult{
			WalletID:       r.WalletID,
			WalletUrl:      r.WalletUrl,
			Identifier:     strings.TrimPrefix(r.Identifier, "https://"),
			IdentifierType: r.IdentifierType,
			CanSend:        true,
		}
		for _, sr := range r.SubResults {
			res[i].SubResults = append(res[i].SubResults, &pb.SearchResult{
				WalletID:       sr.WalletID,
				WalletUrl:      sr.WalletUrl,
				Identifier:     strings.TrimPrefix(sr.Identifier, "https://"),
				IdentifierType: sr.IdentifierType,
				CanSend:        true,
			})
		}
	}

	return &pb.SearchWalletsResponse{Results: res}, nil
}
