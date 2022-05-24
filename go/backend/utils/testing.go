package test_utils

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
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
func MigrateCockroachDB(t *testing.T, ctx context.Context) (db *sqlx.DB, cleanup func()) {
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

	cleanup = func() {
		cleanupQuery := fmt.Sprintf("DROP DATABASE %s;", dbName)
		_, err := db.ExecContext(ctx, cleanupQuery)
		if err != nil {
			t.Fatal(err)
		}

		if err = db.Close(); err != nil {
			t.Fatal(err)
		}
	}

	return db, cleanup
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
		//WaitingFor:   wait.ForHTTP("/health/ready").WithPort("4434"),
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

type PacioliContainer struct {
	DbCleanup      func()
	Tb             testcontainers.Container
	URI            string
	Pacioli        *exec.Cmd
	PacioliUrl     string
	PacioliNetwork testcontainers.Network
}

func (c *PacioliContainer) Terminate(ctx context.Context) error {
	err := c.Pacioli.Process.Kill()
	if err != nil {
		return err
	}

	err = c.Tb.Terminate(ctx)
	if err != nil {
		return err
	}

	c.DbCleanup()

	err = c.PacioliNetwork.Remove(ctx)
	if err != nil {
		return err
	}

	return nil
}

func SetupPacioli(t *testing.T, ctx context.Context) *PacioliContainer {
	fmt.Println("Starting pacioli test container.")
	containerNetwork := "pacioli-" + uuid.NewString()
	network, err := testcontainers.GenericNetwork(ctx, testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{
			Name:           containerNetwork,
			CheckDuplicate: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	_, moduleDir, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Could not get directory path for utils/testing.")
	}

	connString, _, dbCleanup := pacioli_utils.MigrateCockroachDB(t, ctx)

	tb, err := pacioli_utils.SetupTigerBeetle(ctx, 0, containerNetwork)
	if err != nil {
		t.Fatal(err)
	}

	port, err := GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	hostIP := "127.0.0.1"
	pacioli := exec.Command(
		"go",
		"run",
		filepath.Join(filepath.Dir(moduleDir), "../../pacioli/main.go"),
		"start",
	)
	pacioli.Env = append(
		os.Environ(),
		"ENV=testing",
		fmt.Sprintf("PORT=%d", port),
		fmt.Sprintf("DB_URL=%s", connString),
		fmt.Sprintf("TB_URL=%s", tb.URI),
		"TB_CLUSTER_ID=0",
	)
	if err = pacioli.Start(); err != nil {
		t.Fatal(err)
	}

	return &PacioliContainer{
		DbCleanup:      dbCleanup,
		Tb:             tb,
		Pacioli:        pacioli,
		PacioliUrl:     fmt.Sprintf("%s:%d", hostIP, port),
		PacioliNetwork: network,
	}
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
		}
	}))
}
