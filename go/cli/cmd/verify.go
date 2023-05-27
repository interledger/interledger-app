package cmd

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	twitterscraper "github.com/n0madic/twitter-scraper"
	"net/http"
	"regexp"
)

func VerifyIdentity(ctx context.Context, b Backends, args VerifyClaimArgs) error {
	// fetch wallet details
	resp, err := http.Get(args.WalletAddress)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	var walletDetails WalletDetails
	switch {
	case resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices:
		err := json.NewDecoder(resp.Body).Decode(&walletDetails)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unsuccessful get request to wallet details. statusCode=%d url=%s", resp.StatusCode, args.WalletAddress)
	}

	// find the identity in the wallet details using the given identifier and type
	var identity Identity
	for _, id := range walletDetails.Identities {
		if id.Type == args.Type && id.Identifier == args.Identifier {
			identity = id
			break
		}
	}

	// fetch public proof using the public proof URL from the identity
	tweetID, err := getTweetIDFromURL(identity.PublicProof)
	if err != nil {
		return err
	}

	scraper := twitterscraper.New()
	tweet, err := scraper.GetTweet(tweetID)
	if err != nil {
		return fmt.Errorf("Error fetching public proof tweet: %s", err)
	}

	// check identifier and public proof content
	// TODO: check tweet content and format
	if tweet.Username != args.Identifier {
		return fmt.Errorf("Public proof author identifier does not match identifier. author=%s identifier=%s", tweet.Username, args.Identifier)
	}

	// fetch wallet public keys
	resp, err = http.Get(args.WalletAddress + "/jwks.json")
	if err != nil {
		return fmt.Errorf("Error fetching wallet public keys: %s", err)
	}

	defer resp.Body.Close()

	var jwks JWKS
	switch {
	case resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices:
		err := json.NewDecoder(resp.Body).Decode(&jwks)
		if err != nil {
			return fmt.Errorf("Error decoding wallet public keys: %s", err)
		}
	default:
		return fmt.Errorf("Unsuccessful get request to wallet public keys. statusCode=%d url=%s", resp.StatusCode, args.WalletAddress)
	}

	var jwk Jwk
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
	claim := IdentityClaim{
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

	signature, err := base64.RawURLEncoding.DecodeString(identity.Signature)
	if err != nil {
		return fmt.Errorf("Error decoding identity signature: %s", err)
	}

	// TODO: i'm not sure if this is the right way to parse the public key
	publicKey := ed25519.PublicKey(jwk.X)

	// verify the signature
	verified := ed25519.Verify(publicKey, jsonClaim, signature)
	if !verified {
		return fmt.Errorf("Identity verification failed.")
	}

	return nil
}

func getTweetIDFromURL(url string) (string, error) {
	pattern := `^https?://(?:www\.)?twitter\.com/(?:#!/)?[^/]+/status/(\d+).*`

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}

	matches := regex.FindStringSubmatch(url)
	if len(matches) < 2 {
		return "", fmt.Errorf("Invalid public proof tweet URL")
	}

	tweetID := matches[1]
	return tweetID, nil
}
