package gatehub

import (
	"strings"

	"github.com/interledger/interledger-app/go/backend/providers/gatehub/external"
)

// IsUserInKYCEditMode reports whether GateHub considers the user able to edit or
// resubmit KYC. See GateHub docs: edit mode is status 0/state 0; resubmission is
// status 10/state 0.
func IsUserInKYCEditMode(user *User) bool {
	if user == nil {
		return false
	}

	if user.IsProfileCreationDisabled {
		return false
	}

	verification := sumsubVerification(user.Verifications)
	if verification == nil {
		return false
	}

	if verification.State != 0 {
		return false
	}

	return verification.Status == 0 || verification.Status == 10
}

func sumsubVerification(verifications []external.Verification) *external.Verification {
	for i := range verifications {
		verification := &verifications[i]
		if isSumsubProvider(verification) {
			return verification
		}
	}

	return nil
}

func isSumsubProvider(verification *external.Verification) bool {
	return strings.EqualFold(verification.Provider, "Sumsub") ||
		strings.EqualFold(verification.ProviderType, "sumsub")
}
