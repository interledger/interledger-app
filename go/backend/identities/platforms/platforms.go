package platforms

import (
	"context"
	"fmt"

	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/env"
	// "gitlab.com/fynbos/env"
)

type (
	SignedClaimArgs struct {
		Identifier string
		WalletID   string
	}

	VerifyInstructionsArgs struct {
		Identifier string
		WalletID   string
		Identity   *identities.Identity
	}

	GeneratedSignedClaim struct {
		Claim         identities.Claim
		Signature     []byte
		SignatureHash []byte
	}
)

type Platform interface {
	VerifyWorkflow() interface{} // Return the child workflow func to call with the identity ID, only args the workflow must expect is the identityID and the proof URL
	GenerateSignedClaim(ctx context.Context, args *SignedClaimArgs) (*GeneratedSignedClaim, error)
	VerifyInstructions(ctx context.Context, args *VerifyInstructionsArgs) (string, error)
}

func Get(b Backends, platform identities.Platform) (Platform, error) {
	if !env.IsProd() && !env.IsDev() {
		return newDev(platform), nil
	}

	switch platform {
	case identities.PlatformTwitter:
		return newTwitter(b, platform), nil
	}

	return nil, fmt.Errorf("unknown platform: %s", platform)
}
