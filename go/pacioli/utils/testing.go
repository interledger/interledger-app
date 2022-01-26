// TODO: this can be extracted to be re-used for all services
package test_utils

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"

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

	migrationsPath := filepath.Join(filepath.Dir(moduleDir), "../db/kodata")
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
