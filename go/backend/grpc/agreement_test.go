package grpc

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/agreements"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func TestGetAgreement(t *testing.T) {
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	t.Run("Successfully calls GetAgreement", func(t *testing.T) {
		c.AgreementsService.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&agreements.Agreement{
			Content: "privacy policy content",
		}, nil).Times(1)

		resp, err := client.GetAgreement(
			context.Background(),
			&backendv1.GetAgreementRequest{
				Id: "privacy_policy-2.0.0",
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "privacy policy content", resp.GetContent())
	})

	t.Run("Successfully handle errors from GetAgreement", func(t *testing.T) {
		c.AgreementsService.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, agreements.ErrNotFound).Times(1)

		_, err := client.GetAgreement(
			context.Background(),
			&backendv1.GetAgreementRequest{
				Id: "privacy_policy-3.0.0",
			},
		)
		if err == nil {
			t.Fatal("Expected error but got nil")
		}

		assert.Error(t, err)
		assert.EqualError(t, err, "rpc error: code = NotFound desc = Not found: Failed to find agreement.")
	})
}
