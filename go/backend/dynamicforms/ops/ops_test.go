package ops_test

import (
	"bytes"
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
	t.Parallel()
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

func TestListDynamicFroms(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	b := &TestBackends{
		Db: db.MigrateTestDB(t, ctx),
	}

	testForms := []dynamicforms.CreateFormArgs{
		{
			FormID:   "testForm1",
			FormData: `{ "test1": "data" }`,
		},
		{
			FormID:   "testForm1",
			FormData: `{ "test2": "data" }`,
		},
		{
			FormID:   "testForm3",
			FormData: `{ "test3": "data" }`,
		},
	}

	for _, form := range testForms {
		_, err := ops.CreateForm(ctx, b, &form)
		if err != nil {
			t.Fatal(err)
		}
	}

	formCounts, err := ops.ListFormCounts(ctx, b, db.Pagination{})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(formCounts))
}

func TestExportFormResults(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b := &TestBackends{
		Db: db.MigrateTestDB(t, ctx),
	}

	testForms := []dynamicforms.CreateFormArgs{
		{
			FormID:   "testForm1",
			FormData: `{ "name": "Omer" }`,
		},
		{
			FormID:   "testForm1",
			FormData: `{ "name": "Matt" }`,
		},
		{
			FormID:   "testForm3",
			FormData: `{ "test3": { "test4": "data" } }`,
		},
	}
	for _, form := range testForms {
		_, err := ops.CreateForm(ctx, b, &form)
		if err != nil {
			t.Fatal(err)
		}
	}

	buf := &bytes.Buffer{}
	err := ops.ExportFormResults(ctx, b, "testForm1", buf)

	assert.NoError(t, err)
	assert.Equal(t, buf.String(), "Name\nOmer\nMatt\n")
}
