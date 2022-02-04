package cli

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"

	"gitlab.com/fynbos/pacioli/migrations"
)

type StartArgs struct {
	Port               string
	DbConnectionString string
	TbUrl              string
	TbClusterID        uint32
}

func ParseStartArgs() (*StartArgs, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "443" // default port for grpc
	}
	baseDbUrl := os.Getenv("DB_URL")
	if baseDbUrl == "" {
		baseDbUrl = "cockroach://pacioli@cockroachdb-public:26257/pacioli?sslmode=verify-full&max_conns=20&max_idle_conns=4"
	}
	tbUrl := os.Getenv("TB_URL")
	if tbUrl == "" {
		tbUrl = "tigerbeetle-0.tigerbeetle.default.svc.cluster.local:80"
	}
	// TODO: look up TB_URL IP using net package.
	tbClusterID := os.Getenv("TB_CLUSTER_ID")
	if tbClusterID == "" {
		return nil, errors.New("TigerBeetle cluster ID not specified.")
	}
	parsedTbClusterID, err := strconv.ParseUint(tbClusterID, 10, 32)
	if err != nil {
		return nil, err
	}

	connString, err := migrations.InlineSslCreds(
		strings.Replace(baseDbUrl, "cockroach", "postgres", 1), // replace cockroach protocol with postgres so that we can use pq driver.
		"/cockroach-certs/client.pacioli.key",
		"/cockroach-certs/client.pacioli.crt",
		"/cockroach-certs/ca.crt",
	)
	if err != nil {
		log.Fatalln(err)
	}

	return &StartArgs{
		Port:               port,
		DbConnectionString: connString,
		TbUrl:              tbUrl,
		TbClusterID:        uint32(parsedTbClusterID),
	}, nil
}
