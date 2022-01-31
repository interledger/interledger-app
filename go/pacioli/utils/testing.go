// TODO: this can be extracted to be re-used for all services
package test_utils

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
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
)

type CockroachDBContainer struct {
	testcontainers.Container
	URI string
}

// This will start the crdb test container and run the migrations.
func SetupTestCockroachDB(ctx context.Context) (*CockroachDBContainer, error) {
	fmt.Println("Starting CRDB test container.")

	_, moduleDir, _, ok := runtime.Caller(0)
	if !ok {
		return nil, errors.New("Could not get directory path for utils/testing.")
	}

	req := testcontainers.ContainerRequest{
		Image:        "cockroachdb/cockroach:latest-v21.1",
		ExposedPorts: []string{"26257/tcp", "8080/tcp"},
		WaitingFor:   wait.ForHTTP("/health").WithPort("8080"),
		Cmd:          []string{"start-single-node", "--insecure"},
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
	URI     string
	DataDir string
}

func SetupTigerBeetle(ctx context.Context, clusterID uint32) (*TigerBeetleContainer, error) {
	const (
		TIGERBEETLE_PORT = "3000"
		TIGERBEETLE_DIR  = "/var/lib/tigerbeetle"
	)

	dataDir, err := ioutil.TempDir("", "tb")
	if err != nil {
		return nil, err
	}

	fmt.Println("Initializing TigerBeetle temporary data dir:" + dataDir)
	initReq := testcontainers.ContainerRequest{
		Image:        "tigerbeetle", // TODO: host image
		ExposedPorts: []string{TIGERBEETLE_PORT},
		WaitingFor:   wait.ForLog("info: initialized data file").WithPollInterval(1 * time.Second),
		BindMounts: map[string]string{
			TIGERBEETLE_DIR: dataDir,
		},
		Cmd: []string{
			"init",
			fmt.Sprintf("--cluster=%d", clusterID),
			"--replica=0",
			"--directory=" + TIGERBEETLE_DIR,
		},
	}
	_, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: initReq,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	fmt.Println("Starting TigerBeetle test container.")
	req := testcontainers.ContainerRequest{
		Image:        "tigerbeetle", // TODO: host image
		ExposedPorts: []string{TIGERBEETLE_PORT},
		BindMounts: map[string]string{
			TIGERBEETLE_DIR: dataDir,
		},
		WaitingFor: wait.ForLog(fmt.Sprintf("info: cluster=%d replica=0: listening on 0.0.0.0:3000", clusterID)).WithPollInterval(1 * time.Second),
		Cmd: []string{
			"start",
			fmt.Sprintf("--cluster=%d", clusterID),
			"--replica=0",
			"--addresses=0.0.0.0:" + TIGERBEETLE_PORT,
			"--directory=" + TIGERBEETLE_DIR,
		},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, TIGERBEETLE_PORT)
	if err != nil {
		return nil, err
	}

	return &TigerBeetleContainer{Container: container, URI: fmt.Sprintf("0.0.0.0:%s", mappedPort.Port()), DataDir: dataDir}, nil
}
