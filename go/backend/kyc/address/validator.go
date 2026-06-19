package address

import (
	"context"
	"strings"

	"github.com/interledger/interledger-app/go/log"
	"go.uber.org/zap"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/interledger/interledger-app/go/backend/kyc"
	street "github.com/smartystreets/smartystreets-go-sdk/us-street-api"
	"github.com/smartystreets/smartystreets-go-sdk/wireup"
)

type Validator interface {
	USPSAddress(ctx context.Context, address kyc.Address) (bool, error)
}

type client struct {
	extenal *street.Client
}

func New(authID, authToken string) Validator {
	ex := wireup.BuildUSStreetAPIClient(
		wireup.SecretKeyCredential(authID, authToken),
		// The appropriate license values to be used for your subscriptions
		// can be found on the Subscriptions page the account dashboard.
		// https://www.smartystreets.com/docs/cloud/licensing
		wireup.WithLicenses("us-core-cloud"), // TODO get this from Matt
		wireup.WithHTTPClient(otelhttp.DefaultClient),
	)

	return &client{ex}
}

func (c *client) USPSAddress(ctx context.Context, address kyc.Address) (bool, error) {
	// we store state in iso3166-2 format, smarty assumes a US address so we can strip the prefix.
	state := address.State
	stateParts := strings.Split(state, "-")
	if len(stateParts) > 1 {
		state = strings.TrimSpace(stateParts[1])
	}

	lookup := &street.Lookup{
		Street:        address.Line1,
		Street2:       address.Line2,
		Secondary:     strings.TrimSpace(address.Apartment + " " + address.Building),
		City:          address.City,
		State:         state,
		ZIPCode:       address.ZipCode,
		MatchStrategy: street.MatchStrict,
	}

	batch := street.NewBatch()
	batch.Append(lookup)

	err := c.extenal.SendBatchWithContext(ctx, batch)
	if err != nil {
		log.Error("error query address", zap.Any("address", address))
		return true, nil
	}

	for _, input := range batch.Records() {
		if len(input.Results) > 0 {
			return true, nil
		}
	}

	log.Error("usps query did not return any results", zap.Any("address", address))

	return true, nil
}
