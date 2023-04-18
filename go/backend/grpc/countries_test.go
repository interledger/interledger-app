package grpc

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func TestGetCountries(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	rpc, err := client.GetCountries(context.Background(), &backendv1.Empty{})
	require.NoError(t, err)

	require.Len(t, rpc.Countries, 250)
}
