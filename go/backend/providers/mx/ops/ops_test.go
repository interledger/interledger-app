package ops_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/providers/mx/external"
	"gitlab.com/fynbos/backend/providers/mx/external/client/mock"
	"gitlab.com/fynbos/backend/providers/mx/ops"
	"gotest.tools/assert"
)

func TestGetWidget(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	b := ops.NewTestBackends(t, func(tb *ops.TestBackends) {
		tb.ExternalClient = mock.NewMockClient(ctrl)
	})

	walletID := uuid.NewString()
	externalUserGuid := uuid.NewString()
	widgetURL := "widgeturl"
	b.ExternalClient.EXPECT().ListUsersByID(gomock.Any(), walletID).Return(
		&external.ListUsersResponse{
			Users: []external.User{
				{
					Guid: externalUserGuid,
					ID:   walletID,
				},
			},
		},
		nil,
	).AnyTimes()
	b.ExternalClient.EXPECT().GetWidgetURL(gomock.Any(), external.GetWidgetURLArgs{
		UserGuid:            externalUserGuid,
		IncludeTransactions: false,
		IncludeIdentity:     true,
		Mode:                "verification",
		WidgetType:          "connect_widget",
	}).Return(
		&external.WidgetURL{
			URL: widgetURL,
		},
		nil,
	).AnyTimes()

	url, err := ops.GetWidget(context.Background(), b, walletID)
	require.NoError(t, err)
	assert.Equal(t, widgetURL, url)
}
