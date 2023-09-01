package platforms

import (
	"context"
	"crypto"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/cdn"
	"gitlab.com/fynbos/backend/keys"

	"gitlab.com/fynbos/backend/identities"
)

type discordPlatform struct {
	b        Backends
	platform identities.Platform
}

func (d *discordPlatform) VerifyWorkflow() interface{} {
	return nil
}

func (d *discordPlatform) GenerateSignedClaim(ctx context.Context, args *SignedClaimArgs) (*GeneratedSignedClaim, error) {
	walletKeys, err := d.b.Keys().List(ctx, args.WalletID)
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

	wallet, err := d.b.Wallets().Get(ctx, args.WalletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	claim := identities.Claim{
		Wallet:     wallet.AddressString(),
		Type:       "discord",
		Identifier: args.Identifier,
		Kid:        signingKey.ID,
		Ctime:      time.Now().Unix(),
	}

	jsonClaim, err := json.Marshal(claim)

	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	signature, err := d.b.Keys().Sign(ctx, signingKey.ID, args.WalletID, jsonClaim)
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
