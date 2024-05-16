package ops_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/dgryski/trifles/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/dynamicforms"
	"gitlab.com/fynbos/backend/dynamicforms/ops"
)

func TestSubmitForm(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	b := &TestBackends{
		Db: db.MigrateTestDB(t, ctx),
	}

	walletID := uuid.UUIDv4()

	form, err := ops.SubmitForm(ctx, b, &dynamicforms.SubmitArgs{
		FormID:   "testForm",
		Data:     `{ "testform1": "data" }`,
		WalletID: walletID,
	})

	assert.NoError(t, err)
	assert.Equal(t, "testForm", form.FormID)
	assert.Equal(t, walletID, form.WalletID.String)
}

func TestListSubmissionCount(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	b := &TestBackends{
		Db: db.MigrateTestDB(t, ctx),
	}

	testForms := []dynamicforms.SubmitArgs{
		{
			FormID: "testForm1",
			Data:   `{ "test1": "data" }`,
		},
		{
			FormID: "testForm1",
			Data:   `{ "test2": "data" }`,
		},
		{
			FormID: "testForm3",
			Data:   `{ "test3": "data" }`,
		},
	}

	for _, form := range testForms {
		_, err := ops.SubmitForm(ctx, b, &form)
		if err != nil {
			t.Fatal(err)
		}
	}

	formCounts, err := ops.ListSubmissionCount(ctx, b, db.Pagination{})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(formCounts))
}

func TestExportSubmissions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b := &TestBackends{
		Db: db.MigrateTestDB(t, ctx),
	}

	testSubmissions := []dynamicforms.SubmitArgs{
		{
			FormID: "testForm1",
			Data:   `{ "name": "Omer" }`,
		},
		{
			FormID: "testForm1",
			Data:   `{ "name": "Matt" }`,
		},
		{
			FormID: "testForm3",
			Data:   `{ "test3": { "test4": "data" } }`,
		},
	}
	for _, form := range testSubmissions {
		_, err := ops.SubmitForm(ctx, b, &form)
		if err != nil {
			t.Fatal(err)
		}
	}

	buf := &bytes.Buffer{}
	err := ops.ExportSubmissions(ctx, b, "testForm1", buf)

	assert.NoError(t, err)
	assert.Equal(t, buf.String(), "Name\nMatt\nOmer\n")
}
