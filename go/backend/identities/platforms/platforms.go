package platforms

import (
	"context"
	"fmt"

	"gitlab.com/fynbos/backend/identities"
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

	GenerateImagesArgs struct {
		Identifier    string
		WalletURL     string
		SignatureHash []byte
	}
)

type Platform interface {
	VerifyWorkflow() interface{} // Return the child workflow func to call with the identity ID, only args the workflow must expect is the identityID and the proof URL
	GenerateSignedClaim(ctx context.Context, args *SignedClaimArgs) (*GeneratedSignedClaim, error)
	VerifyInstructions(ctx context.Context, args *VerifyInstructionsArgs) (string, error)
	GenerateImages(ctx context.Context, args *GenerateImagesArgs) error
}

func Get(b Backends, platform identities.Platform) (Platform, error) {
	switch platform {
	case identities.PlatformTwitter:
		return newTwitter(b, platform), nil
	case identities.PlatformDomain:
		return newDomainPlatform(b, platform), nil
	case identities.PlatformSlack:
		return &slackPlatform{b, platform}, nil
	}

	return nil, fmt.Errorf("unknown platform: %s", platform)
}
