package test_utils

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

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
		Image:        "823058932981.dkr.ecr.eu-west-1.amazonaws.com/cockroach:latest-v21.1",
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
	Crdb           testcontainers.Container
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

	err = c.Crdb.Terminate(ctx)
	if err != nil {
		return err
	}

	err = c.PacioliNetwork.Remove(ctx)
	if err != nil {
		return err
	}

	return nil
}

func SetupPacioli(ctx context.Context) (*PacioliContainer, error) {
	fmt.Println("Starting pacioli test container.")
	containerNetwork := "pacioli-" + uuid.NewString()
	network, err := testcontainers.GenericNetwork(ctx, testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{
			Name:           containerNetwork,
			CheckDuplicate: true,
		},
	})
	if err != nil {
		return nil, err
	}

	_, moduleDir, _, ok := runtime.Caller(0)
	if !ok {
		return nil, errors.New("Could not get directory path for utils/testing.")
	}

	fmt.Println("Starting pacioli crdb.")
	crdb, err := pacioli_utils.SetupTestCockroachDB(ctx, containerNetwork)
	if err != nil {
		return nil, err
	}

	fmt.Println("Starting pacioli tb.")
	tb, err := pacioli_utils.SetupTigerBeetle(ctx, 0, containerNetwork)
	if err != nil {
		return nil, err
	}

	fmt.Println("Starting pacioli.")
	port, err := getFreePort()
	if err != nil {
		return nil, err
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
		fmt.Sprintf("DB_URL=%s", crdb.URI),
		fmt.Sprintf("TB_URL=%s", tb.URI),
		"TB_CLUSTER_ID=0",
	)
	if err = pacioli.Start(); err != nil {
		return nil, err
	}

	return &PacioliContainer{
		Crdb:           crdb,
		Tb:             tb,
		Pacioli:        pacioli,
		PacioliUrl:     fmt.Sprintf("%s:%d", hostIP, port),
		PacioliNetwork: network,
	}, nil
}

func getFreePort() (int, error) {
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
