package platforms

import (
	"context"
	"crypto"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"

	"gitlab.com/fynbos/backend/cdn"
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/keys"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type domainPlatform struct {
	platform identities.Platform
	b        Backends
}

func newDomainPlatform(b Backends, platform identities.Platform) *domainPlatform {
	return &domainPlatform{
		platform: platform,
		b:        b,
	}
}

func (dp *domainPlatform) VerifyWorkflow() interface{} {
	return DomainVerifyWorkflow
}

func (dp *domainPlatform) VerifyInstructions(ctx context.Context, args *VerifyInstructionsArgs) (string, error) {
	return "", nil
}

func (dp *domainPlatform) GenerateImages(ctx context.Context, args *GenerateImagesArgs) error {
	sigHashBase64 := base64.URLEncoding.EncodeToString(args.SignatureHash)

	img, err := dp.b.Images().GenerateDomainIdentity(ctx, args.WalletURL, args.Identifier)
	if err != nil {
		return err
	}
	err = cdn.Put(ctx, cdn.PutArgs{
		Data:        img,
		ContentType: "image/png",
		Path:        "identities/" + sigHashBase64 + "/domain.png",
	})
	if err != nil {
		return err
	}

	imgOG, err := dp.b.Images().GenerateDomainIdentityOG(ctx, args.WalletURL, args.Identifier)
	if err != nil {
		return err
	}
	err = cdn.Put(ctx, cdn.PutArgs{
		Data:        imgOG,
		ContentType: "image/png",
		Path:        "identities/" + sigHashBase64 + "/domain-og.png",
	})
	if err != nil {
		return err
	}

	return nil
}

func (dp *domainPlatform) GenerateSignedClaim(ctx context.Context, args *SignedClaimArgs) (*GeneratedSignedClaim, error) {
	walletKeys, err := dp.b.Keys().List(ctx, args.WalletID)
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

	wallet, err := dp.b.Wallets().Get(ctx, args.WalletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	claim := identities.Claim{
		Wallet:     wallet.AddressString(),
		Type:       string(identities.PlatformDomain),
		Identifier: args.Identifier,
		Kid:        signingKey.ID,
		Ctime:      time.Now().Unix(),
	}

	jsonClaim, err := json.Marshal(claim)

	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	signature, err := dp.b.Keys().Sign(ctx, signingKey.ID, args.WalletID, jsonClaim)
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

func DomainVerifyWorkflow(ctx workflow.Context, id, domain string) (string, error) {
	var a *DomainActivity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    1 * time.Minute,
			BackoffCoefficient: 1.25,
			MaximumAttempts:    5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("DomainVerifyWorkflow started", "id", id, "proof", domain)

	err := workflow.ExecuteActivity(ctx, a.UpdateDomainIdentityState, id, identities.StatePending, "").Get(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	var identity identities.Identity
	err = workflow.ExecuteActivity(ctx, a.GetDomainIdentity, id).Get(ctx, &identity)
	if err != nil {
		return "", fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	err = workflow.ExecuteActivity(ctx, a.CheckTXTRecords, identity.Identifier, identity.SignatureHash).Get(ctx, nil)
	if err != nil && errors.Is(err, identities.ErrNotFound) {
		err = workflow.ExecuteActivity(ctx, a.UpdateDomainIdentityState, id, identities.StateUnverified, domain).Get(ctx, nil)
		if err != nil {
			return "", fmt.Errorf("%w %s", identities.ErrInternal, err)
		}

		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	err = workflow.ExecuteActivity(ctx, a.UpdateDomainIdentityState, id, identities.StateVerified, domain).Get(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return "", nil
}

type DomainActivityBackends interface {
	Identities() identities.Client
}

type DomainActivity struct {
	b DomainActivityBackends
}

func NewDomainActivity(b DomainActivityBackends) *DomainActivity {
	return &DomainActivity{
		b: b,
	}
}

func (a *DomainActivity) GetDomainIdentity(ctx context.Context, id string) (identities.Identity, error) {
	identity, err := a.b.Identities().Get(ctx, id)
	if err != nil {
		return identities.Identity{}, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return *identity, nil
}

func (a *DomainActivity) CheckTXTRecords(ctx context.Context, domain string, proof []byte) error {
	txtRecords, err := net.LookupTXT("_interledger." + domain)
	if err != nil {
		return fmt.Errorf("%w %s", identities.ErrInternal, err)
	}
	if len(txtRecords) == 0 {
		return fmt.Errorf("%w %s", identities.ErrNotFound, "no TXT records found")
	}

	sighash := base64.URLEncoding.EncodeToString(proof)

	for _, r := range txtRecords {
		if r == sighash {
			return nil
		}
	}

	return fmt.Errorf("%w %s", identities.ErrNotFound, "no matching TXT record found")
}

func (a *DomainActivity) UpdateDomainIdentityState(ctx context.Context, id string, state identities.State, proof string) error {
	err := a.b.Identities().UpdateState(ctx, id, state, proof)
	if err != nil {
		return fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return nil
}
