package ops_test

import (
	"context"
	"testing"

	"gitlab.com/fynbos/backend/fundingsources"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestFundingSources(s *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(s)
	c, err := NewTestContainer(ctx, s, ctrl)
	if err != nil {
		s.Fatal(err)
	}

	s.Run("can create a funding source", func(t *testing.T) {
		walletId := uuid.NewString()

		fs, err := c.Fs.Create(ctx, &fundingsources.CreateArgs{
			WalletID: walletId,
			Name:     "Test",
			Mask:     "1234",
			Type:     "mx",
			SubType:  "bank",
		})
		if err != nil {
			t.Fatal("Should be able to create a funding source")
		}

		assert.NotNil(t, fs)
	})
}
