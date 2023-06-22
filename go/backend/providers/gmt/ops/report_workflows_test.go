package ops_test

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
	"gitlab.com/fynbos/backend/providers/gmt/ops"
)

func TestActivity_ProcessDailyReport(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)

	awsImpl := aws_mock.NewMockClient(ctrl)

	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx), func(tb *ops.TestBackends) {
		tb.AWSImpl = awsImpl
	})
	a := ops.NewActivity(b)

	fd, err := os.Open("testdata/daily_report.csv")
	require.NoError(t, err)
	defer fd.Close()

	awsImpl.EXPECT().S3GetObjectData(ctx, "fynbos-gmt", "daily_report.csv").Return(fd, nil)

	err = a.ProcessDailyReport(ctx, "daily_report.csv")
	require.NoError(t, err)

	res, err := b.DB().QueryContext(ctx, "SELECT * FROM gmt_daily_report")
	require.NoError(t, err)

	var rows [][]string
	for res.Next() {
		row := make([]string, 19)
		err = res.Scan(&row[0], &row[1], &row[2], &row[3], &row[4], &row[5], &row[6], &row[7], &row[8], &row[9], &row[10], &row[11], &row[12], &row[13], &row[14], &row[15], &row[16], &row[17], &row[18])
		require.NoError(t, err)
		rows = append(rows, row)
	}

	assert.Len(t, rows, 8)

	for _, row := range rows {
		fmt.Println(row)
	}
}
