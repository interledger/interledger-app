package admin

import (
	"context"
	"fmt"

	"github.com/interledger/interledger-app/go/backend/entityconf"
	"github.com/interledger/interledger-app/go/backend/walletconf"
	pb "github.com/interledger/interledger-app/go/proto/backend/admin/v1"
)

// GetWalletConfs and SetWalletConf back the botanist "Wallet Settings" tab.
// This is a parallel, independent surface to GetWalletFeatures/
// SetWalletFeatures: it reads and writes entityconf's own store, never
// wallet_features, and nothing else in the application consults these
// values yet.

func (s *AdminRpcService) GetWalletConfs(ctx context.Context, req *pb.GetWalletConfsRequest) (*pb.WalletConfsResponse, error) {
	return s.walletConfsResponse(ctx, req.WalletID)
}

func (s *AdminRpcService) SetWalletConf(ctx context.Context, req *pb.SetWalletConfRequest) (*pb.WalletConfsResponse, error) {
	store := s.b.EntityConfStore()

	def, err := store.StoredDefinition(ctx, req.Key)
	if err != nil {
		return nil, err
	}

	var value any
	switch def.Type {
	case entityconf.TypeBool:
		value = req.BoolValue
	case entityconf.TypeInt:
		value = int(req.IntValue)
	case entityconf.TypeString:
		value = req.StringValue
	default:
		return nil, fmt.Errorf("wallet conf %q has unsupported type %q", req.Key, def.Type)
	}

	if err := store.SetValue(ctx, walletconf.EntityWallet, req.WalletID, req.Key, value); err != nil {
		return nil, err
	}

	return s.walletConfsResponse(ctx, req.WalletID)
}

func (s *AdminRpcService) walletConfsResponse(ctx context.Context, walletID string) (*pb.WalletConfsResponse, error) {
	defs := entityconf.DefinitionsFor(walletconf.EntityWallet)

	keys := make([]string, len(defs))
	for i, d := range defs {
		keys[i] = d.Key
	}

	values, err := s.b.EntityConfStore().ResolveAll(ctx, walletconf.EntityWallet, walletID, keys)
	if err != nil {
		return nil, err
	}

	confs := make([]*pb.WalletConf, 0, len(defs))
	for _, d := range defs {
		conf := &pb.WalletConf{
			Key:         d.Key,
			DisplayName: d.DisplayName,
			Description: d.Description,
			Type:        string(d.Type),
		}

		switch v := values[d.Key].(type) {
		case bool:
			conf.BoolValue = v
		case int:
			conf.IntValue = int64(v)
		case string:
			conf.StringValue = v
		}

		confs = append(confs, conf)
	}

	return &pb.WalletConfsResponse{
		WalletID: walletID,
		Confs:    confs,
	}, nil
}
