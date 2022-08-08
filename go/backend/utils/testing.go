package test_utils

import (
	"context"
	"errors"
	"fmt"
	tigerbeetle_go "github.com/coilhq/tigerbeetle-go"
	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/pacioli"
	pacioli_client "gitlab.com/fynbos/pacioli/client"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	_ "github.com/golang-migrate/migrate/v4/database/cockroachdb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gitlab.com/fynbos/backend/migrations"
	pacioli_utils "gitlab.com/fynbos/pacioli/utils"
)

const testingCrdbConnectionString = "postgres://root@0.0.0.0:26257/%s?sslmode=disable"

// Assumes that CRDB is running locally on port 26257.
func MigrateCockroachDB(t *testing.T, ctx context.Context) (db *sqlx.DB) {
	_, moduleDir, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Could not get directory path for utils/testing.")
	}

	dbName := "backend_test_" + strings.Replace(uuid.NewString(), "-", "", 4)
	connString := os.Getenv("DB_URL")
	if connString == "" {
		connString = testingCrdbConnectionString
	}
	connString = fmt.Sprintf(connString, dbName)
	db, err := sqlx.Connect("postgres", connString)
	if err != nil {
		t.Fatal(err)
	}

	query := fmt.Sprintf("CREATE DATABASE %s;", dbName)
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		t.Fatal(err)
	}

	migrationsPath := filepath.Join(filepath.Dir(moduleDir), "../migrations")
	err = migrations.MigrateFromDir(connString, migrationsPath)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		cleanupQuery := fmt.Sprintf("DROP DATABASE %s;", dbName)
		_, err := db.ExecContext(ctx, cleanupQuery)
		if err != nil {
			t.Fatal(err)
		}

		if err = db.Close(); err != nil {
			t.Fatal(err)
		}
	})

	return db
}

type KratosContainer struct {
	testcontainers.Container
	URI string
}

func SetupKratos() (*KratosContainer, error) {
	fmt.Println("Creating kratos.")

	ctx := context.Background()

	_, moduleDir, _, ok := runtime.Caller(0)
	if !ok {
		return nil, errors.New("Could not get directory path for utils/testing.")
	}
	configPath := filepath.Join(filepath.Dir(moduleDir), "/kratos")

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    configPath,
			Dockerfile: "Dockerfile",
		},
		ExposedPorts: []string{"4433/tcp", "4434/tcp"},
		Env: map[string]string{
			"DSN": "sqlite:///tmp/some-db.sqlite?_fk=true",
		},
		Entrypoint: []string{
			"/bin/sh",
		},
		WaitingFor: wait.ForHTTP("/health/ready").WithPort("4433"),
		Cmd:        []string{"-c", "kratos migrate sql -c /etc/config/kratos/kratos.yml -e -y; kratos serve -c /etc/config/kratos/kratos.yml --dev --watch-courier"},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "4433")
	if err != nil {
		return nil, err
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	connString := fmt.Sprintf("http://%s:%s", hostIP, mappedPort.Port())

	return &KratosContainer{
		container,
		connString,
	}, nil
}

func TruncateDb(ctx context.Context, db *sqlx.DB) error {
	fmt.Println("Truncating all tables.")
	// TODO disabled as query is actually failing
	//const query = `SELECT 'TRUNCATE TABLE ' + Table_Schema + '.' + Table_Name from INFORMATION_SCHEMA.tables where table_type = 'base table'`
	//_, err := db.ExecContext(ctx, "SELECT")
	return nil
}

type pacioliBackends struct {
	db  *sqlx.DB
	tbc tigerbeetle_go.Client
	val *validator.Validate
}

func (t pacioliBackends) DB() *sqlx.DB {
	return t.db
}

func (t pacioliBackends) TigerBeetle() tigerbeetle_go.Client {
	return t.tbc
}

func (t pacioliBackends) Validator() *validator.Validate {
	return t.val
}

func SetupPacioli(t *testing.T, ctx context.Context) pacioli.Client {
	fmt.Println("setup pacioli")

	_, moduleDir, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Could not get directory path for utils/testing.")
	}

	connString, db := pacioli_utils.MigrateCockroachDB(t, ctx)
	err := pacioli_utils.SeedTigerbeetle(t, moduleDir, "0.0.0.0:3000", connString)
	if err != nil {
		t.Fatal(err)
	}

	tbClient, err := tigerbeetle_go.NewClient(0, []string{"0.0.0.0:3000"}, 10)
	if err != nil {
		t.Fatal(err)
	}

	backends := pacioliBackends{
		db:  db,
		tbc: tbClient,
		val: validator.New(),
	}

	t.Cleanup(func() {
		tbClient.Close()
	})

	return pacioli_client.NewLocal(backends)
}

func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port, nil
}

func SetupUnitMockServer(ctx context.Context) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/application-forms":
			fmt.Println("HERE", r.Header.Get("Content-Type"), r.Header.Get("Authorization"))
			if r.Header.Get("Content-Type") != "application/vnd.api+json" {
				w.WriteHeader(400)
				_, err := w.Write([]byte(fmt.Sprintf("Expected 'Content-Type: application/vnd.api+json' header, got: %s", r.Header.Get("Content-Type"))))
				if err != nil {
					w.WriteHeader(400)
				}
			}
			if r.Header.Get("Authorization") != "Bearer test token" {
				w.WriteHeader(400)
				_, err := w.Write([]byte(fmt.Sprintf("Expected 'Authorization: Bearer test token' header, got: %s", r.Header.Get("Authorization"))))
				if err != nil {
					w.WriteHeader(400)
				}
			}
			if r.Method == "POST" {
				_, err := w.Write([]byte(`{
					"data": {
							"type": "applicationForm",
							"id": "411479",
							"attributes": {
									"tags": {
											"userId": "9fe19d6a-ce2e-4401-85f5-442dec6bf242"
									},
									"url": "https://application-form.sh/LJ45W6SSGO6VFFNKMLR5WPOSLH6KMSXQZPGXIPG64SLXHD5TCV4GSYXWZVUSNUEIW2KP5SZOI4RMP6IJRKLF5TTDJTU4TCLU3LQX2XFDIQAMG7TKSXHCQY3KGZ3RFEBYEQCB3GGYUGIUWBXT2ZEIOVNBG72GGNNJKMFJ6",
									"stage": "EnterIndividualInformation",
									"applicantDetails": {
											"applicationType": "Individual",
											"nationality": "US",
											"email": "peter@oscorp.com"
									},
									"allowedApplicationTypes": [
											"Individual"
									]
							}
					}
			}`))
				if err != nil {
					w.WriteHeader(400)
				}
			} else if r.Method == "GET" {
				_, err := w.Write([]byte(`{
					"data": [
							{
									"type": "applicationForm",
									"id": "411479",
									"attributes": {
											"tags": {
													"userId": "9fe19d6a-ce2e-4401-85f5-442dec6bf242"
											},
											"url": "https://application-form.sh/DXB4GXQMBGY377CD5KQ3OWX4XJEF4Z3DQPKTMDGF77CFQM7M55WOQR5C2C3D5N2NYP52AOCSVZX6JLLGSHRLI3DXZ45R43QPDIBWUAI7KL6I7ESUPTB7C7EFURQKMZZSINKSXYQ2N63L7TFPCQVQIW6TVQQLXUYJQP6FY",
											"stage": "EnterIndividualInformation",
											"applicantDetails": {
													"applicationType": "Individual",
													"nationality": "US",
													"email": "peter@oscorp.com"
											},
											"allowedApplicationTypes": [
													"Individual"
											]
									}
							}
					]
			}`))
				if err != nil {
					w.WriteHeader(400)
				}
			}
		case "/applications":
			fmt.Println("HERE", r.Header.Get("Content-Type"), r.Header.Get("Authorization"))
			if r.Header.Get("Content-Type") != "application/vnd.api+json" {
				w.WriteHeader(400)
				_, err := w.Write([]byte(fmt.Sprintf("Expected 'Content-Type: application/vnd.api+json' header, got: %s", r.Header.Get("Content-Type"))))
				if err != nil {
					w.WriteHeader(400)
				}
			}
			if r.Header.Get("Authorization") != "Bearer test token" {
				w.WriteHeader(400)
				_, err := w.Write([]byte(fmt.Sprintf("Expected 'Authorization: Bearer test token' header, got: %s", r.Header.Get("Authorization"))))
				if err != nil {
					w.WriteHeader(400)
				}
			}
			if r.Method == "POST" {
				_, err := w.Write([]byte(`{
					"data": {
						"type": "individualApplication",
						"id": "53",
						"attributes": {
							"createdAt": "2020-01-14T14:05:04.718Z",
							"fullName": {
								"first": "Peter",
								"last": "Parker"
							},
							"ssn": "721074426",
							"address": {
								"street": "20 Ingram St",
								"street2": null,
								"city": "Forest Hills",
								"state": "NY",
								"postalCode": "11375",
								"country": "US"
							},
							"dateOfBirth": "2001-08-10",
							"email": "peter@oscorp.com",
							"phone": {
								"countryCode": "1",
								"number": "5555555555"
							},
							"status": "AwaitingDocuments",
							"ip": "127.0.0.1",
							"soleProprietorship": true,
							"ein": "123456789",
							"dba": "Piedpiper Inc",
							"tags": {
								"fynbosUserId": "106a75e9-de77-4e25-9561-faffe59d7814"
							},
							"archived": false
						},
						"relationships": {
							"org": {
								"data": {
									"type": "org",
									"id": "1"
								}
							},
							"documents": {
								"data": [
									{
										"type": "document",
										"id": "1"
									},
									{
										"type": "document",
										"id": "2"
									}
								]
							}
						}
					},
					"included": [
						{
							"type": "document",
							"id": "1",
							"attributes": {
								"documentType": "AddressVerification",
								"status": "Required",
								"name": "Peter Parker",
								"description": "Please provide a document to verify your address. Document may be a utility bill, bank statement, lease agreement or current pay stub.",
								"address": {
									"street": "20 Ingram St",
									"street2": null,
									"city": "Forest Hills",
									"state": "NY",
									"postalCode": "11375",
									"country": "US"
								}
							}
						},
						{
							"type": "document",
							"id": "2",
							"attributes": {
								"documentType": "IdDocument",
								"status": "Required",
								"name": "Peter Parker",
								"description": "Please provide a copy of your unexpired government issued photo ID which would include Drivers License or State ID.",
								"dateOfBirth": "2001-08-10"
							}
						}
					]
				}`))
				if err != nil {
					w.WriteHeader(400)
				}
			}
		case "/counterparties":
			if r.Method != "POST" {
				http.Error(w, "Method not implemented", 501)
				return
			}

			_, err := w.Write([]byte(fmt.Sprintf(`{
				"data": {
					"type": "achCounterparty",
					"id": "%d"
				}
			}`, rand.Intn(100))))
			if err != nil {
				w.WriteHeader(500)
			}
		}

	}))
}
