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
	_country "gitlab.com/fynbos/backend/country"
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

	cs := _country.NewService()
	is, err := identity.NewService(cs)
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
	as, err := accounts.NewService(is, cs, ledgerCode, pClient)
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
			req := createIdentityRequest(generateIdentityInput())
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
			input := generateIdentityInput()
			req := createIdentityRequest(input)
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
			assert.Equal(tt, user.ID, response.Identity.ID)
			assert.Equal(tt, input.FirstName, response.Identity.FirstName)
			assert.Equal(tt, input.LastName, response.Identity.LastName)
			assert.Equal(tt, input.MobileNumber, response.Identity.MobileNumber)
			assert.Equal(tt, user.Email, response.Identity.Email)
			assert.Equal(tt, input.Country, response.Identity.Country)
			assert.Equal(tt, "", response.Identity.DateOfBirth)
			assert.Equal(tt, []string{}, response.Identity.Address)
			assert.Equal(tt, "", response.Identity.City)
			assert.Equal(tt, "", response.Identity.State)
			assert.Equal(tt, "", response.Identity.PostalCode)
			assert.Equal(tt, "", response.Identity.TaxIDNumber)

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
			assert.Equal(tt, user.ID, userIdentity.ID)
			assert.Equal(tt, input.FirstName, userIdentity.FirstName)
			assert.Equal(tt, input.LastName, userIdentity.LastName)
			assert.Equal(tt, input.MobileNumber, userIdentity.MobileNumber)
			assert.Equal(tt, user.Email, userIdentity.Email)
			assert.Equal(tt, input.Country, userIdentity.Country)
			assert.Equal(tt, "", userIdentity.DateOfBirth)
			assert.Equal(tt, []string{}, userIdentity.Address)
			assert.Equal(tt, "", userIdentity.City)
			assert.Equal(tt, "", userIdentity.State)
			assert.Equal(tt, "", userIdentity.PostalCode)
			assert.Equal(tt, "", userIdentity.TaxIDNumber)

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
			input := generateIdentityInput()
			req := createIdentityRequest(input)
			_user.ActingAs(req, user)
			var respData map[string]generated.CreateIdentityMutationResponse
			if err := client.Run(ctx, req, &respData); err != nil {
				tt.Fatal(err)
			}
			response := respData["createIdentity"]
			assert.Equal(tt, true, response.Success)

			additionalInput := generateIdentityInput(withCountry("ZA"))
			req.Var("input", additionalInput)
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
				Email: faker.Email(),
			}
			input := generateIdentityInput()
			req := createIdentityRequest(input)
			_user.ActingAs(req, user)
			ledgerAccountID := uuid.NewString()
			pClient.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).Return(&pacioliv1.Account{
				Id: ledgerAccountID,
			}, nil).Times(1)

			var respData map[string]generated.CreateIdentityMutationResponse
			if err := client.Run(ctx, req, &respData); err != nil {
				tt.Fatal(err)
			}

			getReq := getIdentityRequest()
			_user.ActingAs(getReq, user)

			var getResp map[string]identity.Identity
			if err := client.Run(ctx, getReq, &getResp); err != nil {
				tt.Fatal(err)
			}

			response := getResp["identity"]
			assert.Equal(tt, user.ID, response.ID)
			assert.Equal(tt, input.FirstName, response.FirstName)
			assert.Equal(tt, input.LastName, response.LastName)
			assert.Equal(tt, input.MobileNumber, response.MobileNumber)
			assert.Equal(tt, user.Email, response.Email)
			assert.Equal(tt, input.Country, response.Country)
			assert.Equal(tt, "", response.DateOfBirth)
			assert.Equal(tt, []string{}, response.Address)
			assert.Equal(tt, "", response.City)
			assert.Equal(tt, "", response.State)
			assert.Equal(tt, "", response.PostalCode)
			assert.Equal(tt, "", response.TaxIDNumber)
		})
	})
}

func createIdentityRequest(input *generated.CreateIdentityInput) *graphql.Request {
	req := graphql.NewRequest(`
			    mutation ($input: CreateIdentityInput!) {
			        createIdentity (input: $input) {
			            code
			            success
			            message
			            identity {
			            	id
			            	firstName
			            	lastName
			            	mobileNumber
			            	email
			            	dateOfBirth
			            	address
			            	city
			            	state
			            	postalCode
			            	country
			            	taxIdNumber
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
		            	firstName
		            	lastName
		            	mobileNumber
		            	email
		            	dateOfBirth
		            	address
		            	city
		            	state
		            	postalCode
		            	country
		            	taxIdNumber
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

// TODO: auto generate helpers.
// Factory to generate create args
func generateIdentityInput(opts ...func(*generated.CreateIdentityInput)) *generated.CreateIdentityInput {
	args := &generated.CreateIdentityInput{
		FirstName:    faker.Name(),
		LastName:     faker.LastName(),
		MobileNumber: faker.Phonenumber(),
		Country:      "US",
	}

	for _, opt := range opts {
		opt(args)
	}

	return args
}

func withFirstName(name string) func(*generated.CreateIdentityInput) {
	return func(args *generated.CreateIdentityInput) {
		args.FirstName = name
	}
}

func withLastName(name string) func(*generated.CreateIdentityInput) {
	return func(args *generated.CreateIdentityInput) {
		args.LastName = name
	}
}

func withMobileNumber(number string) func(*generated.CreateIdentityInput) {
	return func(args *generated.CreateIdentityInput) {
		args.MobileNumber = number
	}
}

func withCountry(country string) func(*generated.CreateIdentityInput) {
	return func(args *generated.CreateIdentityInput) {
		args.Country = country
	}
}
