// TODO: this can be extracted to be re-used for all services
package test_utils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	tigerbeetle_go "github.com/coilhq/tigerbeetle-go"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/cockroachdb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	cli "gitlab.com/fynbos/pacioli/cli"
	"gitlab.com/fynbos/pacioli/ledger"
	"gitlab.com/fynbos/pacioli/seed"
)

const testingCrdbConnectionString = "postgres://root@0.0.0.0:26257/%s?sslmode=disable"

func MigrateCockroachDB(t *testing.T, ctx context.Context) (URI string, db *sqlx.DB) {
	_, moduleDir, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Could not get directory path for utils/testing.")
	}

	dbName := "pacioli_test_" + strings.Replace(uuid.NewString(), "-", "", 4)
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
	m, err := migrate.New(
		"file://"+migrationsPath,
		strings.Replace(connString, "postgres", "cockroach", 1),
	)
	if err != nil {
		t.Fatal(err)
	}

	// The expected behaviour is for it to return ErrNoChange
	if err := m.Up(); err != nil {
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

	return connString, db
}

func TruncateDb(ctx context.Context, db *sqlx.DB) error {
	fmt.Println("Truncating all tables.")
	const query = `SELECT 'TRUNCATE TABLE ' + Table_Schema + '.' + Table_Name from INFORMATION_SCHEMA.tables where table_type = 'base table'`
	_, err := db.ExecContext(ctx, query)
	return err
}

type TigerBeetleContainer struct {
	testcontainers.Container
	URI string
}

func SetupTigerBeetle(ctx context.Context, clusterID uint32, network string) (*TigerBeetleContainer, error) {
	const (
		TIGERBEETLE_PORT = "3000"
		TIGERBEETLE_DIR  = "/var/lib/tigerbeetle"
	)
	containerNetwork := network
	if containerNetwork == "" {
		containerNetwork = "pacioli-test"
	}

	mkdirTbCommand := fmt.Sprintf("mkdir -p %s;", TIGERBEETLE_DIR)
	initTbCommand := fmt.Sprintf("./tigerbeetle init --cluster=%d --replica=0 --directory=%s;", clusterID, TIGERBEETLE_DIR)
	startTbCommand := fmt.Sprintf("./tigerbeetle start --cluster=%d --replica=0 --addresses=0.0.0.0:%s --directory=%s;", clusterID, TIGERBEETLE_PORT, TIGERBEETLE_DIR)

	fmt.Println("Starting TigerBeetle test container.")
	req := testcontainers.ContainerRequest{
		Image:        "823058932981.dkr.ecr.eu-west-1.amazonaws.com/tigerbeetle@sha256:b1fe98356a0db183b56b555eac17c5a43f4b61305f5ac711ea741d5085a2f977", // TODO: host image
		ExposedPorts: []string{TIGERBEETLE_PORT},
		WaitingFor:   wait.ForLog("init").WithPollInterval(1 * time.Second),
		Entrypoint: []string{
			"/bin/bash",
		},
		Networks:       []string{containerNetwork},
		NetworkAliases: map[string][]string{containerNetwork: {"pacioli-tigerbeetle"}},
		Privileged:     true,
		Cmd: []string{
			"-c",
			fmt.Sprintf("%s %s %s", mkdirTbCommand, initTbCommand, startTbCommand),
		},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, TIGERBEETLE_PORT)
	if err != nil {
		return nil, err
	}

	connString, err := cli.ParseTburl(fmt.Sprintf("%s:%s", hostIP, mappedPort.Port()))
	if err != nil {
		return nil, err
	}

	return &TigerBeetleContainer{Container: container, URI: connString[0]}, nil
}

func SeedTigerbeetle(moduleDir, tbURI, dbConn string) error {
	db, err := sqlx.Connect("postgres", dbConn)
	if err != nil {
		return err
	}
	defer db.Close()

	if err != nil {
		return err
	}

	tbClient, err := tigerbeetle_go.NewClient(0, []string{tbURI}, 10)
	if err != nil {
		return err
	}
	defer tbClient.Close()

	ls, err := ledger.NewService(&ledger.ServiceArgs{
		Db: db,
		Tb: tbClient,
	})
	if err != nil {
		return err
	}

	err = seed.TigerBeetle(ls, filepath.Join(filepath.Dir(moduleDir), "../../pacioli/utils/test_seed.yml"))
	return err
}
