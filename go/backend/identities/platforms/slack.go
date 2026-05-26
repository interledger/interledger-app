package platforms

import (
	"context"
	"crypto"
	"encoding/json"
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/keys"
)

type slackPlatform struct {
	b        Backends
	platform identities.Platform
}

func (s *slackPlatform) VerifyWorkflow() interface{} {
	return nil
}

func (s *slackPlatform) GenerateSignedClaim(ctx context.Context, args *SignedClaimArgs) (*GeneratedSignedClaim, error) {
	walletKeys, err := s.b.Keys().List(ctx, args.WalletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	// Get first custodial key
	var signingKey *keys.Key
	for _, k := range walletKeys {
		if k.Type == keys.Custodial {
			signingKey = &k
			break
		}
	}

	if signingKey == nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, "no custodial key found")
	}

	wallet, err := s.b.Wallets().Get(ctx, args.WalletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	claim := identities.Claim{
		Wallet:     wallet.AddressString(),
		Type:       string(identities.PlatformSlack),
		Identifier: args.Identifier,
		Kid:        signingKey.ID,
		Ctime:      time.Now().Unix(),
	}

	jsonClaim, err := json.Marshal(claim)

	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	signature, err := s.b.Keys().Sign(ctx, signingKey.ID, args.WalletID, jsonClaim)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	signatureHash := crypto.SHA256.New()
	signatureHash.Write(signature)
	hash := signatureHash.Sum(nil)

	return &GeneratedSignedClaim{
		Claim:         claim,
		Signature:     signature,
		SignatureHash: hash,
	}, nil
}

func (s *slackPlatform) VerifyInstructions(_ context.Context, _ *VerifyInstructionsArgs) (string, error) {
	return "Successful", nil
}

func (s *slackPlatform) GenerateImages(_ context.Context, _ *GenerateImagesArgs) error {
	return nil
}
