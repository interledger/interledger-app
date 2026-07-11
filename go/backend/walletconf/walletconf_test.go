package walletconf_test

import (
	"testing"

	"github.com/interledger/interledger-app/go/backend/entityconf"
	"github.com/interledger/interledger-app/go/backend/walletconf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfsAreRegistered(t *testing.T) {
	defs := entityconf.DefinitionsFor(walletconf.EntityWallet)
	require.Len(t, defs, 12)

	byKey := map[string]entityconf.Definition{}
	for _, d := range defs {
		assert.Equal(t, walletconf.EntityWallet, d.EntityType)
		assert.Equal(t, entityconf.TypeBool, d.Type)
		assert.NotEmpty(t, d.DisplayName, "conf %q must have a display name", d.Key)
		assert.NotEmpty(t, d.Description, "conf %q must have a description", d.Key)
		byKey[d.Key] = d
	}

	accountsTab, ok := byKey["wallet.accounts_tab_enabled"]
	require.True(t, ok)
	assert.Equal(t, true, accountsTab.CodeDefault, "accounts_tab_enabled mirrors wallet_features' true column default")

	send, ok := byKey["wallet.send_enabled"]
	require.True(t, ok)
	assert.Equal(t, false, send.CodeDefault, "send_enabled mirrors wallet_features' false column default")
}
