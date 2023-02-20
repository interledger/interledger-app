package httpmessagesignatures_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/httpmessagesignatures"
)

func TestCreateContentDigest(t *testing.T) {
	cases := []struct {
		Name           string
		Body           []byte
		ExpectedDigest string
		Algorithms     []string
	}{
		{
			Name:           "creates a single digest from an empty body (SHA256)",
			ExpectedDigest: "sha-256=:47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=:",
			Body:           nil,
			Algorithms:     []string{"sha-256"},
		},
		{
			Name:           "creates a single digest from an empty body (SHA512)",
			ExpectedDigest: "sha-512=:z4PhNX7vuL3xVChQ1m2AB9Yg5AULVxXcg/SpIdNs6c5H0NE8XYXysP+DGNKHfuwvY7kxvUdBeoGlODJ6+SfaPg==:",
			Body:           nil,
			Algorithms:     []string{"sha-512"},
		},
		{
			Name:           "creates a single digest from a body (SHA256)",
			ExpectedDigest: "sha-256=:LsWDvMD3TQ5hD1FciIKL6ePw7YR8BVI5dD6NnJwusRs=:",
			Body:           []byte(`{hello:"world"}`),
			Algorithms:     []string{"sha-256"},
		},
		{
			Name:           "creates a single digest from a body (SHA512)",
			ExpectedDigest: "sha-512=:YwRB5Y5G6jIfS1V0gBi59+hVKgu+vFjZKmeXdqMQQjwrwh5hA0vNbwDQi30SCiOK+e2dRs3P4tMo72WT3BfmQg==:",
			Body:           []byte(`{hello:"world"}`),
			Algorithms:     []string{"sha-512"},
		},
		{
			Name:           "creates multiple digests from a body",
			ExpectedDigest: "sha-256=:LsWDvMD3TQ5hD1FciIKL6ePw7YR8BVI5dD6NnJwusRs=:, sha-512=:YwRB5Y5G6jIfS1V0gBi59+hVKgu+vFjZKmeXdqMQQjwrwh5hA0vNbwDQi30SCiOK+e2dRs3P4tMo72WT3BfmQg==:",
			Body:           []byte(`{hello:"world"}`),
			Algorithms:     []string{"sha-256", "sha-512"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(st *testing.T) {
			digest, err := httpmessagesignatures.CreateContentDigest(context.Background(), tc.Body, tc.Algorithms)
			require.NoError(st, err)
			assert.Equal(st, tc.ExpectedDigest, digest)
		})
	}
}

func TestVerifyContentDigest(t *testing.T) {
	cases := []struct {
		Name   string
		Digest string
		Body   []byte
		Fails  bool
	}{
		{
			Name:   "verifies a single digest (SHA256)",
			Digest: "sha-256=:LsWDvMD3TQ5hD1FciIKL6ePw7YR8BVI5dD6NnJwusRs=:",
			Body:   []byte(`{hello:"world"}`),
			Fails:  false,
		},
		{
			Name:   "verifies a single digest with empty body (SHA256)",
			Digest: "sha-256=:47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=:",
			Body:   nil,
			Fails:  false,
		},
		{
			Name:   "verifies a single digest with empty body (SHA512)",
			Digest: "sha-512=:z4PhNX7vuL3xVChQ1m2AB9Yg5AULVxXcg/SpIdNs6c5H0NE8XYXysP+DGNKHfuwvY7kxvUdBeoGlODJ6+SfaPg==:",
			Body:   nil,
			Fails:  false,
		},
		{
			Name:   "verifies a single digest (SHA512)",
			Digest: "sha-512=:YwRB5Y5G6jIfS1V0gBi59+hVKgu+vFjZKmeXdqMQQjwrwh5hA0vNbwDQi30SCiOK+e2dRs3P4tMo72WT3BfmQg==:",
			Body:   []byte(`{hello:"world"}`),
			Fails:  false,
		},
		{
			Name:   "verifies two digests (SHA256 and SHA512)",
			Digest: "sha-256=:LsWDvMD3TQ5hD1FciIKL6ePw7YR8BVI5dD6NnJwusRs=:, sha-512=:YwRB5Y5G6jIfS1V0gBi59+hVKgu+vFjZKmeXdqMQQjwrwh5hA0vNbwDQi30SCiOK+e2dRs3P4tMo72WT3BfmQg==:",
			Body:   []byte(`{hello:"world"}`),
			Fails:  false,
		},
		{
			Name:   "verifies two digests in any order(SHA256 and SHA512)",
			Digest: "sha-512=:YwRB5Y5G6jIfS1V0gBi59+hVKgu+vFjZKmeXdqMQQjwrwh5hA0vNbwDQi30SCiOK+e2dRs3P4tMo72WT3BfmQg==:, sha-256=:LsWDvMD3TQ5hD1FciIKL6ePw7YR8BVI5dD6NnJwusRs=:",
			Body:   []byte(`{hello:"world"}`),
			Fails:  false,
		},
		{
			Name:   "doesn't verify if any digest fails",
			Digest: "sha-512=:YwRB5Y5G6jIfS10gBi59+hVKgu+vFjZKmeXdqMQQjwrwh5hA0vNbwDQi30SCiOK+e2dRs3P4tMo72WT3BfmQg==:, sha-256=:LsWDvMD3TQ5hD1FciIKL6ePw7YR8BVI5dD6NnJwusRs=:",
			Body:   []byte(`{hello:"world"}`),
			Fails:  true,
		},
		{
			Name:   "doesn't verify if empty digest",
			Digest: "",
			Body:   []byte(`{hello:"world"}`),
			Fails:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(st *testing.T) {
			err := httpmessagesignatures.VerifyContentDigest(context.Background(), tc.Digest, tc.Body)
			if tc.Fails {
				require.Error(t, err)
			} else {
				require.NoError(st, err)
			}
		})
	}
}
