// TODO: this can be extracted to be re-used for all services
package test_utils

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/cockroachdb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	cli "gitlab.com/fynbos/pacioli/cli"
)

type CockroachDBContainer struct {
	testcontainers.Container
	URI string
}

// This will start the crdb test container and run the migrations.
func SetupTestCockroachDB(ctx context.Context, network string) (*CockroachDBContainer, error) {
	fmt.Println("Starting CRDB test container.")
	containerNetwork := network
	if containerNetwork == "" {
		containerNetwork = "pacioli-test"
	}

	_, moduleDir, _, ok := runtime.Caller(0)
	if !ok {
		return nil, errors.New("Could not get directory path for utils/testing.")
	}

	req := testcontainers.ContainerRequest{
		Image:          "823058932981.dkr.ecr.eu-west-1.amazonaws.com/cockroach:latest-v21.1",
		Networks:       []string{containerNetwork},
		NetworkAliases: map[string][]string{containerNetwork: {"pacioli-crdb"}},
		ExposedPorts:   []string{"26257/tcp", "8080/tcp"},
		WaitingFor:     wait.ForHTTP("/health").WithPort("8080"),
		Cmd:            []string{"start-single-node", "--insecure"},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "26257")
	if err != nil {
		return nil, err
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	connString := fmt.Sprintf("postgres://root@%s:%s/pacioli?sslmode=disable", hostIP, mappedPort.Port())
	db, err := sqlx.Connect("postgres", connString)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	fmt.Println("Creating pacioli database")
	const query = `CREATE DATABASE pacioli;`
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return nil, err
	}

	migrationsPath := filepath.Join(filepath.Dir(moduleDir), "../migrations")
	fmt.Println("Applying migrations from file://" + migrationsPath)

	m, err := migrate.New(
		"file://"+migrationsPath,
		fmt.Sprintf("cockroach://root@%s:%s/pacioli?sslmode=disable", hostIP, mappedPort.Port()))
	if err != nil {
		return nil, err
	}

	// The expected behaviour is for it to return ErrNoChange
	if err := m.Up(); err != nil {
		return nil, err
	}

	return &CockroachDBContainer{Container: container, URI: connString}, nil
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

	initTbCommand := fmt.Sprintf("./tigerbeetle init --cluster=%d --replica=0 --directory=%s;", clusterID, TIGERBEETLE_DIR)
	startTbCommand := fmt.Sprintf("./tigerbeetle start --cluster=%d --replica=0 --addresses=0.0.0.0:%s --directory=%s;", clusterID, TIGERBEETLE_PORT, TIGERBEETLE_DIR)

	fmt.Println("Starting TigerBeetle test container.")
	req := testcontainers.ContainerRequest{
		Image:        "823058932981.dkr.ecr.eu-west-1.amazonaws.com/tigerbeetle:patch-1", // TODO: host image
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
			fmt.Sprintf("%s %s", initTbCommand, startTbCommand),
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
