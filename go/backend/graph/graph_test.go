package graph

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"

	// "github.com/stretchr/testify/assert"

	"gitlab.com/fynbos/backend/db/utils"
	"gitlab.com/fynbos/backend/graph/generated"
	"gitlab.com/fynbos/backend/services"
)

func TestGraphql(t *testing.T) {
	ctx := context.Background()
	crdb, err := utils.SetupTestCockroachDB(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer crdb.Container.Terminate(ctx)

	// the tests are run in serial. We use a global connection for
	// each of the tests.
	db, err := sqlx.Connect("postgres", crdb.URI)
	defer db.Close()

	organisations, err := services.NewOrganisationsService(db)
	if err != nil {
		t.Fatal(err)
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &Resolver{
		Organisations: organisations,
	}}))
	http.Handle("/graphql", srv)
	server := httptest.NewServer(srv)
	defer server.Close()

	client := graphql.NewClient(server.URL + "/graphql")

	t.Run("can create an organisation", func(tt *testing.T) {
		req := graphql.NewRequest(`
		    mutation ($name: String!) {
		        createOrganisation (name: $name) {
		            id
		            name
		        }
		    }
		`)
		req.Var("name", "My first graphql organisation.")

		var respData map[string]map[string]string
		if err := client.Run(ctx, req, &respData); err != nil {
			tt.Fatal(err)
		}

		org, err := organisations.Get(respData["createOrganisation"]["id"])
		if err != nil {
			tt.Fatal(err)
		}
		assert.Equal(tt, respData["createOrganisation"]["name"], "My first graphql organisation.")
		assert.Equal(tt, org.Name, "My first graphql organisation.")
	})

	t.Run("can get an organisation", func(tt *testing.T) {
		org, err := organisations.Create("My second graphql organisation.")
		if err != nil {
			tt.Fatal(err)
		}
		req := graphql.NewRequest(`
		    query ($id: String!) {
		        organisation (id: $id) {
		            id
		            name
		        }
		    }
		`)
		req.Var("id", org.ID)

		var respData map[string]map[string]string
		if err := client.Run(ctx, req, &respData); err != nil {
			tt.Fatal(err)
		}
		assert.Equal(tt, respData["organisation"]["name"], "My second graphql organisation.")
		assert.Equal(tt, respData["organisation"]["id"], org.ID)
	})
}
