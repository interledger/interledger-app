package platforms

import (
	"context"
	"fmt"
	"gitlab.com/fynbos/cli/identities"
)

type Platform interface {
	FetchPublicProof(ctx context.Context, url string) (*PublicProof, error)
}

func Get(platform identities.Platform) (Platform, error) {
	switch platform {
	case identities.PlatformTwitter:
		return newTwitterPlatform(), nil
	}

	return nil, fmt.Errorf("unknown platform: %s", platform)
}
