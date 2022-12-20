// TODO: this can be extracted to be re-used for all services
package test_utils

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	tigerbeetle_go "github.com/coilhq/tigerbeetle-go"
	_ "github.com/golang-migrate/migrate/v4/database/cockroachdb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	cli "gitlab.com/fynbos/pacioli/cli"
	"gitlab.com/fynbos/pacioli/seed"
)

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
		Image: "823058932981.dkr.ecr.eu-west-1.amazonaws.com/tigerbeetle@sha256:b1fe98356a0db183b56b555eac17c5a43f4b61305f5ac711ea741d5085a2f977", // TODO: host image
		//Image:        "localhost:5005/tigerbeetle:debug", // TODO: host image
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

func SeedTigerbeetle(t *testing.T, moduleDir, tbURI, dbConn string) error {
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

	err = seed.TigerBeetle(NewBackends(t, db, tbClient), filepath.Join(filepath.Dir(moduleDir), "../../pacioli/utils/test_seed.yml"))
	return err
}
