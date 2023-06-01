package ops

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gitlab.com/fynbos/cli/identities"
	"gitlab.com/fynbos/cli/identities/platforms"
	"net/http"
)

func VerifyIdentity(ctx context.Context, args *identities.VerifyArgs) error {
	// fetch wallet details
	resp, err := http.Get(args.WalletAddress)
	if err != nil {
		return fmt.Errorf("Error fetching wallet details: %s", err)
	}

	defer resp.Body.Close()

	var walletDetails identities.WalletDetails
	switch {
	case resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices:
		err := json.NewDecoder(resp.Body).Decode(&walletDetails)
		if err != nil {
			return fmt.Errorf("Error decoding wallet details: %s", err)
		}
	default:
		return fmt.Errorf("Unsuccessful get request to wallet details. statusCode=%d url=%s", resp.StatusCode, args.WalletAddress)
	}

	// find the identity in the wallet details using the given identifier and type
	var identity identities.Identity
	for _, id := range walletDetails.Identities {
		if id.Type == args.Type && id.Identifier == args.Identifier {
			identity = id
			break
		}
	}
	if identity.Identifier == "" {
		return fmt.Errorf("No matching identity found in wallet. identifier=%s type=%s", args.Identifier, args.Type)
	}

	// fetch public proof using the public proof URL from the identity
	platform, err := platforms.Get(identities.Platform(args.Type))
	if err != nil {
		return fmt.Errorf("Error getting identity platform: %s", err)
	}

	publicProof, err := platform.FetchPublicProof(ctx, identity.PublicProof)
	if err != nil {
		return fmt.Errorf("Error fetching public proof: %s", err)
	}

	// check identifier and public proof content
	// TODO: check tweet content and format
	if publicProof.Author != args.Identifier {
		return fmt.Errorf("Public proof author identifier does not match identifier. author=%s identifier=%s", publicProof.Author, args.Identifier)
	}

	// fetch wallet public keys
	resp, err = http.Get(args.WalletAddress + "/jwks.json")
	if err != nil {
		return fmt.Errorf("Error fetching wallet public keys: %s", err)
	}

	defer resp.Body.Close()

	var jwks identities.JWKS
	switch {
	case resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices:
		err := json.NewDecoder(resp.Body).Decode(&jwks)
		if err != nil {
			return fmt.Errorf("Error decoding wallet public keys: %s", err)
		}
	default:
		return fmt.Errorf("Unsuccessful get request to wallet public keys. statusCode=%d url=%s", resp.StatusCode, args.WalletAddress)
	}

	var jwk identities.Jwk
	for _, k := range jwks.Keys {
		if k.Kid == identity.KeyID {
			jwk = k
			break
		}
	}
	if jwk.Kid == "" {
		return fmt.Errorf("No matching key found for identity. keyID=%s", identity.KeyID)
	}

	// construct the identity claim
	claim := identities.IdentityClaim{
		Wallet:       args.WalletAddress,
		Identifier:   args.Identifier,
		Type:         args.Type,
		KeyID:        identity.KeyID,
		CreationTime: identity.CreationTime,
	}
	jsonClaim, err := json.Marshal(claim)
	if err != nil {
		return fmt.Errorf("Error marshalling identity claim: %s", err)
	}

	signature, err := base64.URLEncoding.DecodeString(identity.Signature)
	if err != nil {
		return fmt.Errorf("Error decoding identity signature: %s", err)
	}

	publicKey, err := base64.URLEncoding.DecodeString(jwk.X)
	if err != nil {
		return fmt.Errorf("Error decoding identity public key: %s", err)
	}

	// verify the signature
	verified := ed25519.Verify(publicKey, jsonClaim, signature)
	if !verified {
		return fmt.Errorf("Identity verification failed.")
	}

	return nil
}
