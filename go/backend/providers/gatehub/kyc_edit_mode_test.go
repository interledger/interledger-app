package gatehub

import (
	"testing"

	"github.com/interledger/interledger-app/go/backend/providers/gatehub/external"
	"github.com/stretchr/testify/assert"
)

func TestIsUserInKYCEditMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		user *User
		want bool
	}{
		{
			name: "nil user",
			user: nil,
			want: false,
		},
		{
			name: "profile creation disabled",
			user: &User{IsProfileCreationDisabled: true},
			want: false,
		},
		{
			name: "missing verification",
			user: &User{},
			want: false,
		},
		{
			name: "edit mode status 0 state 0",
			user: &User{
				Verifications: []external.Verification{{
					Provider: "Sumsub",
					Status:   0,
					State:    0,
				}},
			},
			want: true,
		},
		{
			name: "resubmission status 10 state 0",
			user: &User{
				Verifications: []external.Verification{{
					ProviderType: "sumsub",
					Status:       10,
					State:        0,
				}},
			},
			want: true,
		},
		{
			name: "valid completed status 1 state 1",
			user: &User{
				Verifications: []external.Verification{{
					Provider: "Sumsub",
					Status:   1,
					State:    1,
				}},
			},
			want: false,
		},
		{
			name: "pending review status 1 state 0",
			user: &User{
				Verifications: []external.Verification{{
					Provider: "Sumsub",
					Status:   1,
					State:    0,
				}},
			},
			want: false,
		},
		{
			name: "rejected status 2 state 0",
			user: &User{
				Verifications: []external.Verification{{
					Provider: "Sumsub",
					Status:   2,
					State:    0,
				}},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsUserInKYCEditMode(tt.user))
		})
	}
}
