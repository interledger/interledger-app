package provision

import (
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
