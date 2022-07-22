package graph

import (
	"context"
	"testing"

	_ "github.com/lib/pq"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"

	"gitlab.com/fynbos/backend/graph/generated"
)

func TestCountries(s *testing.T) {
	s.Skip("being deprecated")
	ctx := context.Background()
	container, err := NewTestContainer(ctx, s)
	if err != nil {
		s.Fatal(err)
	}

	s.Cleanup(func() {
		err = container.Cleanup(ctx)
		if err != nil {
			s.Fatal(err)
		}
	})

	/*
		Scenario: user needs to fetch all countries
		Should return a list of all countries sorted alphabetically by name
	*/
	s.Run("user gets all countries sorted by name", func(t *testing.T) {
		response, err := getAllCountries(container)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 249, len(response))
	})

}

func getAllCountries(container *TestContainer) ([]*generated.Country, error) {
	req := graphql.NewRequest(`
			    query GetCountries {
						countries {
							id
							name
						}
			    }
			`)
	var data map[string][]*generated.Country
	if err := container.Client.Run(container.Ctx, req, &data); err != nil {
		return nil, err
	}
	ret := data["countries"]

	return ret, nil
}
