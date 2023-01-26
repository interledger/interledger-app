package ilp

import "context"

type Client interface {
	CreateStreamCredentials(ctx context.Context, args CreateStreamCredentialsArgs) (*StreamCredentials, error)
	ClearIncomingPackets(ctx context.Context, packets []IncomingPacket) error
}
