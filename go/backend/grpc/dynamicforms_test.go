package grpc

import (
	"context"
	"gitlab.com/fynbos/backend/dynamicforms"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func TestSubmitForm(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	c.walletImpl.EXPECT().ForContext(gomock.Any()).Return(nil, nil).AnyTimes()
	c.dynamicforms.EXPECT().Submit(gomock.Any(), &dynamicforms.SubmitArgs{
		FormID: "test",
		Data:   `{ "test": "data" }`,
	}).Return(nil, nil).AnyTimes()

	_, err := client.SubmitForm(context.Background(), &backendv1.SubmitFormRequest{
		FormId: "test",
		Data:   `{ "test": "data" }`,
	})

	require.NoError(t, err)
}
