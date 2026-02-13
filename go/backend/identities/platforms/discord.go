package platforms

import (
	"context"
	"encoding/base64"

	"gitlab.com/fynbos/backend/cdn"

	"gitlab.com/fynbos/backend/identities"
)

type discordPlatform struct {
	b        Backends
	platform identities.Platform
}

func (d *discordPlatform) VerifyWorkflow() interface{} {
	return nil
}

func (d *discordPlatform) GenerateImages(ctx context.Context, args *GenerateImagesArgs) error {
	sigHashBase64 := base64.URLEncoding.EncodeToString(args.SignatureHash)

	img, err := d.b.Images().GenerateDiscordIdentity(ctx, args.WalletURL, args.Identifier)
	if err != nil {
		return err
	}
	err = cdn.Put(ctx, cdn.PutArgs{
		Data:        img,
		ContentType: "image/png",
		Path:        "identities/" + sigHashBase64 + "/discord.png",
	})
	if err != nil {
		return err
	}

	imgOG, err := d.b.Images().GenerateDiscordIdentityOG(ctx, args.WalletURL, args.Identifier)
	if err != nil {
		return err
	}
	err = cdn.Put(ctx, cdn.PutArgs{
		Data:        imgOG,
		ContentType: "image/png",
		Path:        "identities/" + sigHashBase64 + "/discord-og.png",
	})
	if err != nil {
		return err
	}

	return nil
}

func (d *discordPlatform) VerifyInstructions(ctx context.Context, args *VerifyInstructionsArgs) (string, error) {
	return "Successful", nil
}
