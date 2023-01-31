package httpmessagesignatures

import (
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"strings"

	"github.com/dunglas/httpsfv"
)

var supportedAlgorithms = "sha-256,sha-512"

// Creates content-digests of the body using the specified algorithms. A serialized dictionary
// of the digests are returned.
// https://datatracker.ietf.org/doc/html/rfc8941#section-3.2
// https://www.ietf.org/archive/id/draft-ietf-httpbis-digest-headers-10.html#content-digest
func CreateContentDigest(ctx context.Context, body []byte, algorithms []string) (string, error) {
	dictionary := httpsfv.NewDictionary()
	for _, algo := range algorithms {
		var hasher hash.Hash
		switch algo {
		case "sha-256":
			hasher = sha256.New()
		case "sha-512":
			hasher = sha512.New()
		default:
			return "", fmt.Errorf("Unsupported algorithm=%s", algo)
		}

		_, err := hasher.Write(body)
		if err != nil {
			return "", err
		}
		dictionary.Add(algo, httpsfv.NewItem(hasher.Sum(nil)))
	}

	ret, err := httpsfv.Marshal(dictionary)
	if err != nil {
		return "", err
	}

	return ret, nil
}

func VerifyContentDigest(ctx context.Context, digest string, body []byte) error {
	dictionary, err := httpsfv.UnmarshalDictionary([]string{digest})
	if err != nil {
		return err
	}

	algorithms := dictionary.Names()
	for _, algo := range algorithms {
		if !strings.Contains(supportedAlgorithms, algo) {
			dictionary.Del(algo)
		}
	}

	computedDigest, err := CreateContentDigest(ctx, body, algorithms)
	if err != nil {
		return err
	}

	if computedDigest != digest {
		return fmt.Errorf("Content digest failed verification.")
	}

	return nil
}
