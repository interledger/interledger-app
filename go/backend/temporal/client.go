package temporal

import (
	"go.temporal.io/sdk/client"
)

func NewTemporalClient() (client.Client, error) {
	c, err := client.NewClient(client.Options{
		HostPort: "temporal:7233",
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}
