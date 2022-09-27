package test_utils

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	tigerbeetle_go "github.com/coilhq/tigerbeetle-go"
	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/pacioli"
	pacioli_client "gitlab.com/fynbos/pacioli/client"

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

	/*t.Cleanup(func() {
		cleanupQuery := fmt.Sprintf("DROP DATABASE %s;", dbName)
		_, err := db.ExecContext(ctx, cleanupQuery)
		if err != nil {
			t.Fatal(err)
		}

		if err = db.Close(); err != nil {
			t.Fatal(err)
		}
	})*/

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
