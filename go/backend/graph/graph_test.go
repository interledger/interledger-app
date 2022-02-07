package graph

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbsqlx"
	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"gitlab.com/fynbos/backend/accounts"
	_account "gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/graph/generated"
	"gitlab.com/fynbos/backend/identity"
	_user "gitlab.com/fynbos/backend/user"
	test_utils "gitlab.com/fynbos/backend/utils"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
	mockPacioliV1 "gitlab.com/fynbos/proto/pacioli/v1/mock"
)

func TestGraphql(s *testing.T) {
	ctx := context.Background()
	crdb, err := test_utils.SetupTestCockroachDB(ctx)
	if err != nil {
		s.Fatal(err)
	}
	defer crdb.Container.Terminate(ctx)

	logger, err := zap.NewDevelopment()
	if err != nil {
		s.Fatal(err)
	}
	defer logger.Sync()

	// the tests are run in serial. We use a global connection for
	// each of the tests.
	db, err := sqlx.Connect("postgres", crdb.URI)
	defer db.Close()

	is, err := identity.NewService()
	if err != nil {
		s.Fatal(err)
	}
	is = identity.NewLoggingService(is, logger)

	ctrl := gomock.NewController(s)
	defer ctrl.Finish()
	pacioliLedgerID := uuid.NewString()
	ledgerCode := uint16(1)
	pClient := mockPacioliV1.NewMockPacioliServiceClient(ctrl)
	pClient.EXPECT().GetLedgerByCode(gomock.Any(), gomock.Any()).Return(&pacioliv1.Ledger{
		Id:   pacioliLedgerID,
		Code: uint32(ledgerCode),
	}, nil).Times(1)
	as, err := accounts.NewService(is, ledgerCode, pClient)
	if err != nil {
		s.Fatal(err)
	}
	as = accounts.NewLoggingService(as, logger)

	users := _user.NewMockService()
	users = _user.NewLoggingService(users, logger)

	graph, err := NewService(GraphqlOpts{
		Db:       db,
		Identity: is,
		User:     users,
		Account:  as,
	})
	graph = NewLoggingService(graph, logger)

	router := chi.NewRouter()
	router.Use(_user.MakeMiddleware(users))
	router.Handle("/graphql", MakeHandler(graph, GraphqlHttpHandlerOpts{}))
	server := httptest.NewServer(router)
	defer server.Close()

	client := graphql.NewClient(server.URL + "/graphql")

	s.Run("create identity", func(t *testing.T) {
		t.Run("requires authenticated user", func(tt *testing.T) {
			req := createIdentityRequest(generated.IdentityInput{
				LegalName: faker.Name(),
				Country:   "USA",
			})
			_user.ActingAs(req, nil)

			var respData map[string]generated.CreateIdentityMutationResponse
			err = client.Run(ctx, req, &respData)

			assert.Error(tt, err)
		})

		t.Run("creates identity and account", func(tt *testing.T) {
			tt.Cleanup(func() {
				test_utils.TruncateDb(ctx, db)
			})
			user := &_user.User{
				ID:    uuid.New().String(),
				Email: faker.Email(),
			}
			name := faker.Name()
			req := createIdentityRequest(generated.IdentityInput{
				LegalName: name,
				Country:   "USA",
			})
			_user.ActingAs(req, user)
			ledgerAccountID := uuid.NewString()
			pClient.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).Return(&pacioliv1.Account{
				Id: ledgerAccountID,
			}, nil).Times(1)

			var respData map[string]generated.CreateIdentityMutationResponse
			if err := client.Run(ctx, req, &respData); err != nil {
				tt.Fatal(err)
			}

			response := respData["createIdentity"]
			assert.Equal(tt, "200", response.Code)
			assert.Equal(tt, "Created account holder.", response.Message)
			assert.Equal(tt, true, response.Success)
			assert.Equal(tt, response.Identity.LegalName, name)
			assert.Equal(tt, response.Identity.Country, "USA")
			assert.Equal(tt, response.Identity.Email, user.Email)
			assert.Equal(tt, response.Identity.ID, user.ID)

			var userIdentity *identity.Identity
			err = crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
				_identity, err := is.Get(ctx, tx, user.ID)
				if err != nil {
					return err
				}

				userIdentity = _identity
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(tt, userIdentity.LegalName, name)
			assert.Equal(tt, userIdentity.Country, "USA")
			assert.Equal(tt, userIdentity.Email, user.Email)
			assert.Equal(tt, userIdentity.ID, user.ID)

			pClient.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Return(&pacioliv1.Account{
				Id:              ledgerAccountID,
				DebitsReserved:  1, // return non-zero to make sure default values aren't used.
				DebitsAccepted:  2,
				CreditsAccepted: 3,
				CreditsReserved: 4,
			}, nil).Times(1)
			var account *_account.Account
			err = crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
				_acc, err := as.GetByIdentityID(ctx, tx, user.ID)
				if err != nil {
					return err
				}

				account = _acc
				return nil
			})
			if err != nil {
				tt.Fatal(err)
			}
			assert.Equal(tt, userIdentity.ID, account.IdentityID)
			assert.Equal(tt, ledgerAccountID, account.LedgerAccountID)
			assert.Equal(tt, uint64(1), account.DebitsReserved)
			assert.Equal(tt, uint64(2), account.DebitsAccepted)
			assert.Equal(tt, uint64(3), account.CreditsAccepted)
			assert.Equal(tt, uint64(4), account.CreditsReserved)
		})

		t.Run("user can only create 1 identity", func(tt *testing.T) {
			tt.Cleanup(func() {
				test_utils.TruncateDb(ctx, db)
			})
			pClient.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).Return(&pacioliv1.Account{
				Id: uuid.NewString(),
			}, nil).Times(1)
			user := &_user.User{
				ID:    uuid.NewString(),
				Email: faker.Email(),
			}
			name := faker.Name()
			req := createIdentityRequest(generated.IdentityInput{
				LegalName: name,
				Country:   "USA",
			})
			_user.ActingAs(req, user)
			var respData map[string]generated.CreateIdentityMutationResponse
			if err := client.Run(ctx, req, &respData); err != nil {
				tt.Fatal(err)
			}
			response := respData["createIdentity"]
			assert.Equal(tt, true, response.Success)

			req.Var("input", generated.IdentityInput{
				LegalName: name,
				Country:   "ZAR",
			})
			err = client.Run(ctx, req, &respData)
			assert.EqualError(tt, err, "graphql: Unable to process request.")
		})
	})

	s.Run("get identity", func(t *testing.T) {
		t.Run("requires authenticated user", func(tt *testing.T) {
			req := getIdentityRequest()
			_user.ActingAs(req, nil)

			var respData map[string]identity.Identity
			err := client.Run(ctx, req, &respData)

			assert.Error(tt, err)
		})

		t.Run("returns not found if there is no identity", func(tt *testing.T) {
			tt.Cleanup(func() {
				test_utils.TruncateDb(ctx, db)
			})
			user := &_user.User{
				ID:    uuid.New().String(),
				Email: faker.Name(),
			}
			req := getIdentityRequest()
			_user.ActingAs(req, user)

			var respData map[string]identity.Identity
			err := client.Run(ctx, req, &respData)

			assert.EqualError(tt, err, "graphql: Not found.")
		})

		t.Run("user can get their identity", func(tt *testing.T) {
			tt.Cleanup(func() {
				test_utils.TruncateDb(ctx, db)
			})
			user := &_user.User{
				ID:    uuid.New().String(),
				Email: faker.Name(),
			}
			var id *identity.Identity
			err = crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
				_id, err := is.Create(ctx, tx, identity.CreateArgs{
					Country:   "USA",
					LegalName: faker.Name(),
					User:      user,
				})
				if err != nil {
					return err
				}

				id = _id
				return nil
			})
			if err != nil {
				tt.Fatal(err)
			}
			req := getIdentityRequest()
			_user.ActingAs(req, user)

			var respData map[string]identity.Identity
			if err := client.Run(ctx, req, &respData); err != nil {
				tt.Fatal(err)
			}

			response := respData["identity"]
			assert.Equal(tt, response.LegalName, id.LegalName)
			assert.Equal(tt, response.Country, id.Country)
			assert.Equal(tt, response.Email, id.Email)
			assert.Equal(tt, response.ID, id.ID)
		})
	})
}

func createIdentityRequest(input generated.IdentityInput) *graphql.Request {
	req := graphql.NewRequest(`
			    mutation ($input: IdentityInput!) {
			        createIdentity (input: $input) {
			            code
			            success
			            message
			            identity {
			            	id
			            	email
			            	legalName
			            	country
			            }
			        }
			    }
			`)
	req.Var("input", input)

	return req
}

func getIdentityRequest() *graphql.Request {
	return graphql.NewRequest(`
			    query {
			        identity {
			            id
			            legalName
			            email
			            country
			        }
			    }
			`)
}

func getAccountRequest() *graphql.Request {
	return graphql.NewRequest(`
			    query {
			        account {
			            id
			            balance
			            asset
			        }
			    }
			`)
}
