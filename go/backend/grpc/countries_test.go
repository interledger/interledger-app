package grpc

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/country"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func TestGetCountries(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	c.CountriesService.EXPECT().GetAll(gomock.Any()).Return(
		[]country.Country{
			{
				ID:   "1234",
				Name: "South Africa",
			},
		},
		nil,
	)

	rpc, err := client.GetCountries(context.Background(), &backendv1.Empty{})
	require.NoError(t, err)

	require.Len(t, rpc.Countries, 1)
	assert.Equal(t, "1234", rpc.Countries[0].Id)
	assert.Equal(t, "South Africa", rpc.Countries[0].Name)
}
