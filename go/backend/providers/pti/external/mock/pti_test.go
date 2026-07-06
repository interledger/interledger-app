package mock_test

import (
	"encoding/json"
	"io/fs"
	"os"
	"testing"

	"github.com/interledger/interledger-app/go/backend/providers/pti/external"
	"github.com/interledger/interledger-app/go/backend/providers/pti/external/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatesFileStorage(t *testing.T) {
	t.Cleanup(func() {
		_ = os.Remove("pti_mock.json")
	})
	_, err := os.Stat("pti_mock.json")
	var pathError *fs.PathError
	require.ErrorAs(t, err, &pathError)

	pti := mock.NewPTI()
	_, err = os.Stat("pti_mock.json")
	require.NoError(t, err)
	assert.Empty(t, pti.Users)
}

func TestHydratesFromFileStorage(t *testing.T) {
	file, err := os.CreateTemp("", "pti_mock_*.json")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = file.Close()
	})

	data := mock.PTI{
		Users: map[string]external.CreateUserArgs{
			"1": {
				ID:   "1",
				Type: "PERSON",
				Name: external.Name{
					First: "James",
					Last:  "Bond",
				},
			},
		},
		Wallets: map[string]external.Wallet{
			"a": {
				WalletID:  "a",
				Reference: "gg",
				Balance:   1,
			},
		},
		WalletToUser: map[string]string{
			"a": "1",
		},
	}
	rawData, err := json.Marshal(data)
	require.NoError(t, err)
	_, err = file.Write(rawData)
	require.NoError(t, err)

	t.Setenv("DATA_FILE_PATH", file.Name())

	pti := mock.NewPTI()
	assert.True(t, assert.ObjectsAreEqual(data.Users, pti.Users))
	assert.True(t, assert.ObjectsAreEqual(data.Wallets, pti.Wallets))
	assert.True(t, assert.ObjectsAreEqual(data.WalletToUser, pti.WalletToUser))
}

func TestSave(t *testing.T) {
	file, err := os.CreateTemp("", "pti_mock_*.json")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = file.Close()
	})
	t.Setenv("DATA_FILE_PATH", file.Name())

	pti := mock.NewPTI()
	pti.Users["1"] = external.CreateUserArgs{
		ID:   "1",
		Type: "PERSON",
		Name: external.Name{
			First: "James",
			Last:  "Bond",
		},
	}
	require.NoError(t, pti.Save())

	data, err := os.ReadFile(file.Name())
	require.NoError(t, err)

	var mock mock.PTI
	err = json.Unmarshal(data, &mock)
	require.NoError(t, err)
	assert.True(t, assert.ObjectsAreEqual(mock.Users, pti.Users))
	assert.True(t, assert.ObjectsAreEqual(mock.Wallets, pti.Wallets))
	assert.True(t, assert.ObjectsAreEqual(mock.WalletToUser, pti.WalletToUser))
}
