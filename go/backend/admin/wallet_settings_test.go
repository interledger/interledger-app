package admin

import (
	"context"
	"testing"

	"github.com/interledger/interledger-app/go/backend/entityconf"
	"github.com/interledger/interledger-app/go/backend/walletconf"
	pb "github.com/interledger/interledger-app/go/proto/backend/admin/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeBackends embeds the (unexported, package-internal) Backends interface
// as a nil value: only EntityConfStore is overridden, which is all
// GetWalletConfs/SetWalletConf call. Any other method would panic on the
// embedded nil interface if it were ever (incorrectly) invoked.
type fakeBackends struct {
	Backends
	store entityconf.Store
}

func (f fakeBackends) EntityConfStore() entityconf.Store {
	return f.store
}

func newTestAdminService(t *testing.T) *AdminRpcService {
	t.Helper()

	store := entityconf.NewInMemoryStore()
	require.NoError(t, store.SyncDefinitions(context.Background(), entityconf.DefinitionsFor(walletconf.EntityWallet)))

	return &AdminRpcService{b: fakeBackends{store: store}}
}

func confsByKey(confs []*pb.WalletConf) map[string]*pb.WalletConf {
	out := make(map[string]*pb.WalletConf, len(confs))
	for _, c := range confs {
		out[c.Key] = c
	}
	return out
}

func TestGetWalletConfs_ReturnsDefaultsWhenNoOverride(t *testing.T) {
	s := newTestAdminService(t)

	resp, err := s.GetWalletConfs(context.Background(), &pb.GetWalletConfsRequest{WalletID: "wallet-1"})
	require.NoError(t, err)
	assert.Equal(t, "wallet-1", resp.WalletID)
	require.Len(t, resp.Confs, 12)

	byKey := confsByKey(resp.Confs)

	accountsTab, ok := byKey["wallet.accounts_tab_enabled"]
	require.True(t, ok)
	assert.Equal(t, "bool", accountsTab.Type)
	assert.True(t, accountsTab.BoolValue)
	assert.Equal(t, "Accounts Tab", accountsTab.DisplayName)

	send, ok := byKey["wallet.send_enabled"]
	require.True(t, ok)
	assert.False(t, send.BoolValue)
}

func TestSetWalletConf_OverridesJustThatWallet(t *testing.T) {
	s := newTestAdminService(t)
	ctx := context.Background()

	resp, err := s.SetWalletConf(ctx, &pb.SetWalletConfRequest{
		WalletID:  "wallet-1",
		Key:       "wallet.send_enabled",
		BoolValue: true,
	})
	require.NoError(t, err)
	assert.True(t, confsByKey(resp.Confs)["wallet.send_enabled"].BoolValue)

	other, err := s.GetWalletConfs(ctx, &pb.GetWalletConfsRequest{WalletID: "wallet-2"})
	require.NoError(t, err)
	assert.False(t, confsByKey(other.Confs)["wallet.send_enabled"].BoolValue,
		"a different wallet must not see wallet-1's override")
}

func TestSetWalletConf_UnknownKeyErrors(t *testing.T) {
	s := newTestAdminService(t)

	_, err := s.SetWalletConf(context.Background(), &pb.SetWalletConfRequest{
		WalletID: "wallet-1",
		Key:      "wallet.nonexistent",
	})
	assert.Error(t, err)
}
