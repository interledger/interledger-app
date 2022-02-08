package test_utils

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	_ "github.com/golang-migrate/migrate/v4/database/cockroachdb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gitlab.com/fynbos/backend/migrations"
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

	connString := fmt.Sprintf("postgres://root@%s:%s/backend?sslmode=disable", hostIP, mappedPort.Port())
	db, err := sqlx.Connect("postgres", connString)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	fmt.Println("Creating backend database")
	const query = `CREATE DATABASE backend;`
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return nil, err
	}

	migrationsPath := filepath.Join(filepath.Dir(moduleDir), "../migrations")
	fmt.Println("Applying migrations from file://" + migrationsPath)
	err = migrations.MigrateFromDir(connString, migrationsPath)
	if err != nil {
		return nil, err
	}

	return &CockroachDBContainer{Container: container, URI: connString}, nil
}

// We use docker compose as the migrations need to be run using another container.
func SetupKratos() (string, error) {
	fmt.Println("Creating kratos.")

	_, moduleDir, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("Could not get directory path for utils/testing.")
	}

	composeFilePaths := []string{
		filepath.Join(filepath.Dir(moduleDir), "../../../services/kratos/docker-compose-dev.yaml"),
	}
	identifier := strings.ToLower(uuid.New().String())

	compose := testcontainers.NewLocalDockerCompose(composeFilePaths, identifier)
	execError := compose.
		WithCommand([]string{"up", "-d"}).
		Invoke()
	err := execError.Error
	if err != nil {
		return "", fmt.Errorf("Could not run compose file: %v - %v", composeFilePaths, err)
	}

	return identifier, nil
}

func TeardownKratos(identifier string) error {
	fmt.Println("Tearing down kratos.")
	_, moduleDir, _, ok := runtime.Caller(0)
	if !ok {
		return errors.New("Could not get directory path for utils/testing.")
	}

	composeFilePaths := []string{
		filepath.Join(filepath.Dir(moduleDir), "../../../services/kratos/docker-compose-dev.yaml"),
	}

	compose := testcontainers.NewLocalDockerCompose(composeFilePaths, identifier)
	execError := compose.Down()
	err := execError.Error
	if err != nil {
		return fmt.Errorf("Could not run compose file: %v - %v", composeFilePaths, err)
	}

	return nil
}

func TruncateDb(ctx context.Context, db *sqlx.DB) error {
	fmt.Println("Truncating all tables.")
	const query = `SELECT 'TRUNCATE TABLE ' + Table_Schema + '.' + Table_Name from INFORMATION_SCHEMA.tables where table_type = 'base table'`
	_, err := db.ExecContext(ctx, query)
	return err
}
