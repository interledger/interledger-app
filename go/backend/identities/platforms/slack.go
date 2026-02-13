package platforms

import (
	"context"
	"encoding/base64"

	"gitlab.com/fynbos/backend/cdn"
	"gitlab.com/fynbos/backend/identities"
)

type slackPlatform struct {
	b        Backends
	platform identities.Platform
}

func (s *slackPlatform) VerifyWorkflow() interface{} {
	return nil
}

func (s *slackPlatform) VerifyInstructions(_ context.Context, _ *VerifyInstructionsArgs) (string, error) {
	return "Successful", nil
}

func (s *slackPlatform) GenerateImages(ctx context.Context, args *GenerateImagesArgs) error {
	sigHashBase64 := base64.URLEncoding.EncodeToString(args.SignatureHash)

	img, err := s.b.Images().GenerateSlackIdentity(ctx, args.WalletURL, args.Identifier)
	if err != nil {
		return err
	}
	err = cdn.Put(ctx, cdn.PutArgs{
		Data:        img,
		ContentType: "image/png",
		Path:        "identities/" + sigHashBase64 + "/slack.png",
	})
	if err != nil {
		return err
	}

	imgOG, err := s.b.Images().GenerateSlackIdentityOG(ctx, args.WalletURL, args.Identifier)
	if err != nil {
		return err
	}
	err = cdn.Put(ctx, cdn.PutArgs{
		Data:        imgOG,
		ContentType: "image/png",
		Path:        "identities/" + sigHashBase64 + "/slack-og.png",
	})
	if err != nil {
		return err
	}

	return nil
}
