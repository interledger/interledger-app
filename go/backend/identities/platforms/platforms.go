package platforms

import (
	"errors"
	"fmt"

	"gitlab.com/fynbos/backend/identities/platforms/dev"

	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/env"
)

type Platform interface {
	VerifyWorkflow() interface{} // Return the child workflow func to call with the identity ID, only arg the workflow must expect is the identityID
	NewVerifyCode() string
	VerifyInstructions() string
}

func Get(platform identities.Platform) (Platform, error) {
	if !env.IsProd() && !env.IsDev() {
		return dev.New(platform), nil
	}

	switch platform {
	case identities.PlatformTwitter:
		// TODO: Add twitter platform interface
		return nil, errors.New("TODO make twitter")
	}

	return nil, fmt.Errorf("unknown platform: %s", platform)
}
