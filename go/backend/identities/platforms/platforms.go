package platforms

import (
	"fmt"

	"gitlab.com/fynbos/backend/identities"

	"gitlab.com/fynbos/env"
)

type Platform interface {
	VerifyWorkflow() interface{} // Return the child workflow func to call with the identity ID, only args the workflow must expect is the identityID and the proof URL
	NewVerifyCode() string
	VerifyInstructions() string
}

func Get(platform identities.Platform) (Platform, error) {
	if !env.IsProd() && !env.IsDev() {
		return newDev(platform), nil
	}

	switch platform {
	case identities.PlatformTwitter:
		return newTwitter(platform), nil
	}

	return nil, fmt.Errorf("unknown platform: %s", platform)
}
