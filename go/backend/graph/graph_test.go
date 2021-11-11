package graph

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"

	"gitlab.com/fynbos/backend/authorization"
	"gitlab.com/fynbos/backend/graph/generated"
	"gitlab.com/fynbos/backend/organisation"
	userLib "gitlab.com/fynbos/backend/user"
	test_utils "gitlab.com/fynbos/backend/utils"
)

func TestGraphql(s *testing.T) {
	ctx := context.Background()
	crdb, err := test_utils.SetupTestCockroachDB(ctx)
	if err != nil {
		s.Fatal(err)
	}
	defer crdb.Container.Terminate(ctx)

	// the tests are run in serial. We use a global connection for
	// each of the tests.
	db, err := sqlx.Connect("postgres", crdb.URI)
	defer db.Close()

	authz, err := authorization.NewService()
	if err != nil {
		s.Fatal(err)
	}

	org, err := organisation.NewService(db, authz)
	if err != nil {
		s.Fatal(err)
	}

	users := userLib.NewMockService()

	router := chi.NewRouter()
	router.Use(userLib.MakeMiddleware(users))
	router.Handle("/graphql", MakeHandler(org, users))
	server := httptest.NewServer(router)
	defer server.Close()

	client := graphql.NewClient(server.URL + "/graphql")

	s.Run("authenticated user can create an organisation", func(t *testing.T) {
		t.Cleanup(func() {
			test_utils.TruncateDb(ctx, db)
		})
		user := userLib.User{
			ID: uuid.New().String(),
		}
		req := graphql.NewRequest(`
		    mutation ($name: String!) {
		        createOrganisation (name: $name) {
		            code
		            success
		            message
		            organisation {
		            	id
		            	name
		            }
		        }
		    }
		`)
		req.Var("name", "My first graphql organisation.")
		userLib.ActingAs(req, &user)

		var respData map[string]generated.OrganisationMutationResponse
		if err := client.Run(ctx, req, &respData); err != nil {
			t.Fatal(err)
		}

		org, err := org.Get(respData["createOrganisation"].Organisation.ID, user)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "200", respData["createOrganisation"].Code)
		assert.Equal(t, "Created organisation.", respData["createOrganisation"].Message)
		assert.Equal(t, true, respData["createOrganisation"].Success)
		assert.Equal(t, respData["createOrganisation"].Organisation.Name, "My first graphql organisation.")
		assert.Equal(t, org.Name, "My first graphql organisation.")
	})

	s.Run("unauthenticated user can't create an organisation", func(t *testing.T) {
		t.Cleanup(func() {
			test_utils.TruncateDb(ctx, db)
		})
		req := graphql.NewRequest(`
		    mutation ($name: String!) {
		        createOrganisation (name: $name) {
		            code
		            success
		            message
		            organisation {
		            	id
		            	name
		            }
		        }
		    }
		`)
		req.Var("name", "My first graphql organisation.")
		userLib.ActingAs(req, nil)

		var respData map[string]generated.OrganisationMutationResponse
		err := client.Run(ctx, req, &respData)

		assert.Error(t, err)
	})

	s.Run("user can get their organisation", func(t *testing.T) {
		t.Cleanup(func() {
			test_utils.TruncateDb(ctx, db)
		})
		user := userLib.User{
			ID: uuid.New().String(),
		}
		org, err := org.Create("My second graphql organisation.", user)
		if err != nil {
			t.Fatal(err)
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
		userLib.ActingAs(req, &user)

		var respData map[string]organisation.Organisation
		if err := client.Run(ctx, req, &respData); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, respData["organisation"].Name, "My second graphql organisation.")
		assert.Equal(t, respData["organisation"].ID, org.ID)
	})

	s.Run("user can only get their own organisation", func(t *testing.T) {
		t.Cleanup(func() {
			test_utils.TruncateDb(ctx, db)
		})
		me := userLib.User{
			ID: uuid.New().String(),
		}
		otherUser := userLib.User{
			ID: uuid.New().String(),
		}
		notMyOrg, err := org.Create("Not my organisation.", otherUser)
		if err != nil {
			t.Fatal(err)
		}
		req := graphql.NewRequest(`
		    query ($id: String!) {
		        organisation (id: $id) {
		            id
		            name
		        }
		    }
		`)
		req.Var("id", notMyOrg.ID)
		userLib.ActingAs(req, &me)

		var respData map[string]organisation.Organisation
		err = client.Run(ctx, req, &respData)

		// oso recommends returning a not found error.
		assert.EqualError(t, err, "graphql: Not found.")
	})
}
