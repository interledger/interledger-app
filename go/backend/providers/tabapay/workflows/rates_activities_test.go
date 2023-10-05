package workflows_test

import (
	"context"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	aws_mock "gitlab.com/fynbos/backend/aws/client/mock"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/providers/tabapay/workflows"
)

func TestActivity_LoadFXRatesFromS3(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	b := workflows.NewTestBackends(func(tb *workflows.TestBackends) {
		tb.AwsCliet = aws_mock.NewMockClient(ctrl)
		tb.Db = db.MigrateTestDB(t, ctx)
	})
	a := workflows.NewActivity(b)

	fd, err := os.Open("test_data/fx_rates.csv")
	require.NoError(t, err)
	defer fd.Close()

	b.AwsCliet.EXPECT().S3GetObjectData(gomock.Any(), gomock.Any(), "testdata/fx_rates.csv").Return(fd, nil)

	err = a.LoadFXRatesFromS3(ctx, "testdata/fx_rates.csv")
	require.NoError(t, err)

	var cnt int
	err = b.DB().GetContext(ctx, &cnt, "SELECT count(*) FROM tabapay_fx_rates WHERE currency_code=$1", currency.EUR)
	require.NoError(t, err)

	assert.Equal(t, 2, cnt)
}
