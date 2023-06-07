package workflows_test

import (
	"context"
	"testing"

	"gitlab.com/fynbos/backend/db"

	"github.com/golang/mock/gomock"
	aws_mock "gitlab.com/fynbos/backend/aws/client/mock"
	"gitlab.com/fynbos/backend/providers/tabapay/workflows"
)

func TestActivity_ProcessChargebacksReports(t *testing.T) {

	ctrl := gomock.NewController(t)
	b := workflows.NewTestBackends(func(tb *workflows.TestBackends) {
		tb.AWSImpl = aws_mock.NewMockClient(ctrl)
		tb.Db = db.MigrateTestDB(t, context.TODO())
	})
	a := workflows.NewActivity(b)

}
