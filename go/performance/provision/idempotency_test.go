package provision

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsAlreadyExistsError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "plain already exists", err: errors.New("identity already exists"), want: true},
		{name: "wrapped", err: errors.New("register failed: email already in use"), want: true},
		{name: "kratos duplicate identifier", err: errors.New("an account with the same identifier (email, phone, username, ...) exists already"), want: true},
		{name: "different error", err: errors.New("permission denied"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, isAlreadyExistsError(tt.err))
		})
	}
}

func TestTargetMinorAmount(t *testing.T) {
	require.Equal(t, int64(500000), targetMinorAmount(5000, 2))
	require.Equal(t, int64(5000), targetMinorAmount(5000, 0))
	require.Equal(t, int64(50000), targetMinorAmount(5000, 1))
}

func TestWalletAddressLabel(t *testing.T) {
	require.Equal(t, "perfza001", walletAddressLabel("perf-za-001"))
	require.Equal(t, "perf_za_001", walletAddressLabel("perf_za_001"))
	require.Equal(t, "perf", walletAddressLabel("---"))
}

type stubKratosAuthenticator struct {
	attempts     []string
	loginErrs    map[string]error
	loginTokens  map[string]string
	whoAmITokens map[string]string
	whoAmIErrs   map[string]error
}

func (s *stubKratosAuthenticator) Login(_ context.Context, identifier, _ string) (string, error) {
	s.attempts = append(s.attempts, identifier)
	if err, ok := s.loginErrs[identifier]; ok {
		return "", err
	}
	if token, ok := s.loginTokens[identifier]; ok {
		return token, nil
	}
	return "", nil
}

func (s *stubKratosAuthenticator) WhoAmI(_ context.Context, token string) (string, error) {
	if err, ok := s.whoAmIErrs[token]; ok {
		return "", err
	}
	if userID, ok := s.whoAmITokens[token]; ok {
		return userID, nil
	}
	return "", nil
}

func TestLoginExistingIdentityFallsBackToPhone(t *testing.T) {
	auth := &stubKratosAuthenticator{
		loginErrs: map[string]error{
			"perf-001@perf.interledger.test": errors.New("invalid credentials"),
		},
		loginTokens: map[string]string{
			"+27820000001": "token-phone",
		},
		whoAmITokens: map[string]string{
			"token-phone": "user-phone",
		},
	}

	userID, token, err := loginExistingIdentity(context.Background(), auth, "perf-001@perf.interledger.test", "secret", "+27820000001")
	require.NoError(t, err)
	require.Equal(t, "user-phone", userID)
	require.Equal(t, "token-phone", token)
	require.Equal(t, []string{"perf-001@perf.interledger.test", "+27820000001"}, auth.attempts)
}
