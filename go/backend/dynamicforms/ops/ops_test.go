package ops_test

import (
	"context"
	"encoding/json"
	"github.com/dgryski/trifles/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/dynamicforms"
	"gitlab.com/fynbos/backend/dynamicforms/ops"
	"testing"
)

func TestCreateDynamicForm(t *testing.T) {
	ctx := context.Background()

	b := &TestBackends{
		Db: db.MigrateTestDB(t, ctx),
	}

	walletID := uuid.UUIDv4()
	data := `{ "test": "data" }`
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}

	form, err := ops.CreateForm(ctx, b, &dynamicforms.CreateFormArgs{
		FormID:   "testForm",
		FormData: `{ "test": "data" }`,
		WalletID: walletID,
	})

	assert.NoError(t, err)
	assert.Equal(t, "testForm", form.FormID)
	assert.Equal(t, jsonData, form.Data)
	assert.Equal(t, walletID, form.WalletID)
}
