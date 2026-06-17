package grpc

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/interledger/interledger-app/go/backend/country"
	backendv1 "github.com/interledger/interledger-app/go/proto/backend/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCountries(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	rpc, err := client.GetCountries(context.Background(), &backendv1.Empty{})
	require.NoError(t, err)
	require.Len(t, rpc.Countries, 250)
	assert.Equal(t, country.US.String(), rpc.Countries[0].Id)
	assert.Equal(t, country.ZA.String(), rpc.Countries[1].Id)
	assert.Equal(t, country.CA.String(), rpc.Countries[2].Id)
}
