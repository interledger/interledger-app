package ops

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/kyc/smileid"
)

func GetSmileIDToken(ctx context.Context, sc smileid.Client, walletID string) (string, error) {
	jobID := uuid.NewString()
	token, err := sc.GetToken(ctx, walletID, jobID, smileid.EnhancedKYCProduct)
	if err != nil {
		return "", fmt.Errorf("%w %s", kyc.ErrInternal, err)
	}

	return token, nil
}
