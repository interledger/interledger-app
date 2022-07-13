package temporal

import (
	"go.temporal.io/sdk/client"
)

func NewTemporalClient(temporalUrl string) (client.Client, error) {
	c, err := client.NewClient(client.Options{
		HostPort: temporalUrl,
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}
