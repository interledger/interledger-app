package workflows_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	aws_mock "gitlab.com/fynbos/backend/aws/client/mock"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/providers/tabapay/workflows"
)

func TestActivity_ProcessChargebacksReports(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	b := workflows.NewTestBackends(func(tb *workflows.TestBackends) {
		tb.AWSImpl = aws_mock.NewMockClient(ctrl)
		tb.Db = db.MigrateTestDB(t, ctx)
	})
	a := workflows.NewActivity(b)

	fd, err := os.Open("testdata/chargebacks.csv")
	require.NoError(t, err)
	defer fd.Close()

	b.AWSImpl.EXPECT().S3GetObjectData(ctx, "tabapayreports", "chargebacks.csv").Return(fd, nil)

	err = a.ProcessChargebacksReports(ctx, "chargebacks.csv")
	require.NoError(t, err)

	res, err := b.DB().QueryContext(ctx, "SELECT * FROM tabapay_report_chargebacks")
	require.NoError(t, err)

	var rows [][]string
	for res.Next() {
		row := make([]string, 34)
		err = res.Scan(&row[0], &row[1], &row[2], &row[3], &row[4], &row[5], &row[6], &row[7], &row[8], &row[9], &row[10], &row[11], &row[12], &row[13], &row[14], &row[15], &row[16], &row[17], &row[18], &row[19], &row[20], &row[21], &row[22], &row[23], &row[24], &row[25], &row[26], &row[27], &row[28], &row[29], &row[30], &row[31], &row[32], &row[33])
		require.NoError(t, err)
		rows = append(rows, row)
	}

	// 17 valid lines in the file 4 identical so only inserted the first one
	assert.Len(t, rows, 14)

	for _, row := range rows {
		fmt.Println(row)
	}
}

func TestActivity_ProcessAMLTransactionsReport(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	b := workflows.NewTestBackends(func(tb *workflows.TestBackends) {
		tb.AWSImpl = aws_mock.NewMockClient(ctrl)
		tb.Db = db.MigrateTestDB(t, ctx)
	})
	a := workflows.NewActivity(b)

	fd, err := os.Open("testdata/aml_transactions.csv")
	require.NoError(t, err)
	defer fd.Close()

	b.AWSImpl.EXPECT().S3GetObjectData(ctx, "tabapayreports", "aml_transactions.csv").Return(fd, nil)

	err = a.ProcessAMLTransactionsReport(ctx, "aml_transactions.csv")
	require.NoError(t, err)

	res, err := b.DB().QueryContext(ctx, "SELECT * FROM tabapay_report_aml_transaction")
	require.NoError(t, err)

	var rows [][]string
	for res.Next() {
		row := make([]string, 24)
		err = res.Scan(&row[0], &row[1], &row[2], &row[3], &row[4], &row[5], &row[6], &row[7], &row[8], &row[9], &row[10], &row[11], &row[12], &row[13], &row[14], &row[15], &row[16], &row[17], &row[18], &row[19], &row[20], &row[21], &row[22], &row[23])
		require.NoError(t, err)
		rows = append(rows, row)
	}

	// 4 valid lines in the file all identical so only inserted the first one
	assert.Len(t, rows, 1)

	for _, row := range rows {
		fmt.Println(row)
	}
}
