package platforms

import (
	"context"
	"crypto"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gitlab.com/fynbos/backend/cdn"
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/backend/linkedin"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
	"net/url"
	"path"
	"time"
)

type linkedinPlatform struct {
	platform identities.Platform
	b        Backends
}

func newLinkedinPlatform(b Backends, platform identities.Platform) *linkedinPlatform {
	return &linkedinPlatform{
		platform: platform,
		b:        b,
	}
}

func (lp *linkedinPlatform) VerifyWorkflow() interface{} {
	return LinkedinVerifyWorkflow
}

func (lp *linkedinPlatform) GenerateSignedClaim(ctx context.Context, args *SignedClaimArgs) (*GeneratedSignedClaim, error) {
	walletKeys, err := lp.b.Keys().List(ctx, args.WalletID)
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

	pp, err := lp.b.OpenPayments().GetWalletPaymentPointer(ctx, args.WalletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	claim := identities.Claim{
		Wallet:     pp.URL,
		Type:       "linkedin",
		Identifier: args.Identifier,
		Kid:        signingKey.ID,
		Ctime:      time.Now().Unix(),
	}

	jsonClaim, err := json.Marshal(claim)

	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	signature, err := lp.b.Keys().Sign(ctx, signingKey.ID, args.WalletID, jsonClaim)
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

func (lp *linkedinPlatform) GenerateImages(ctx context.Context, args *GenerateImagesArgs) error {
	sigHashBase64 := base64.URLEncoding.EncodeToString(args.SignatureHash)

	img, err := lp.b.Images().GenerateTwitterIdentity(ctx, args.WalletURL, args.Identifier)
	if err != nil {
		return err
	}
	err = cdn.Put(ctx, cdn.PutArgs{
		Data:        img,
		ContentType: "image/png",
		Path:        "identities/" + sigHashBase64 + "/twitter.png",
	})
	if err != nil {
		return err
	}

	imgOG, err := lp.b.Images().GenerateTwitterIdentityOG(ctx, args.WalletURL, args.Identifier)
	if err != nil {
		return err
	}
	err = cdn.Put(ctx, cdn.PutArgs{
		Data:        imgOG,
		ContentType: "image/png",
		Path:        "identities/" + sigHashBase64 + "/twitter-og.png",
	})
	if err != nil {
		return err
	}

	return nil
}

func (lp *linkedinPlatform) VerifyInstructions(ctx context.Context, args *VerifyInstructionsArgs) (string, error) {
	return "", nil
}

type LinkedinActivityBackends interface {
	Identities() identities.Client
	Linkedin() linkedin.Client
}

type LinkedinActivity struct {
	b LinkedinActivityBackends
}

func NewLinkedinActivity(b LinkedinActivityBackends) *LinkedinActivity {
	return &LinkedinActivity{
		b: b,
	}
}

func LinkedinVerifyWorkflow(ctx workflow.Context, identityID, proofURL string) (string, error) {
	var a *LinkedinActivity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("LinkedinVerifyWorkflow for linkedin platform started", "identityID", identityID, "proofURL", proofURL)

	var linkedinProof linkedin.Post
	err := workflow.ExecuteActivity(ctx, a.FetchPublicProof, proofURL).Get(ctx, &linkedinProof)
	if err != nil {
		return "", fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	var identity identities.Identity
	err = workflow.ExecuteActivity(ctx, a.GetIdentity, identityID).Get(ctx, &identity)
	if err != nil {
		return "", fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	err = workflow.ExecuteActivity(ctx, a.VerifyLinkedinProof, identity, linkedinProof).Get(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	err = workflow.ExecuteActivity(ctx, a.VerifyIdentity, identityID, proofURL).Get(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return "OK", nil
}

func (a *LinkedinActivity) FetchPublicProof(ctx context.Context, proofURL string) (*linkedin.Post, error) {
	return a.b.Linkedin().GetPost(ctx, proofURL)
}

func (a *LinkedinActivity) GetIdentity(ctx context.Context, id string) (identities.Identity, error) {
	identity, err := a.b.Identities().Get(ctx, id)
	if err != nil {
		return identities.Identity{}, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return *identity, nil
}

func (a *LinkedinActivity) VerifyLinkedinProof(ctx context.Context, identity identities.Identity, proof linkedin.Post) error {
	activity.GetLogger(ctx).Info("Verifying proof", "identity", identity.ID, "linkedin post", proof.URLs[0])

	parsedUrl, err := url.Parse(proof.URLs[0])
	if err != nil {
		return fmt.Errorf("%w %s", identities.ErrInternal, err)
	}
	sigHash := path.Base(parsedUrl.Path)
	base64SigHash := base64.URLEncoding.EncodeToString(identity.SignatureHash)

	if sigHash != base64SigHash {
		return fmt.Errorf("%w %s", identities.ErrInternal, "proof sighash doesn't match identity sighash")
	}

	// verify the username
	if identity.Identifier != proof.Author {
		return fmt.Errorf("%w %s", identities.ErrInternal, "linkedin username doesn't match identity username")
	}

	return nil
}

func (a *LinkedinActivity) VerifyIdentity(ctx context.Context, id, proof string) error {
	err := a.b.Identities().UpdateState(ctx, id, identities.StateVerified, proof)
	if err != nil {
		return fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return nil
}
