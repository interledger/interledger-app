package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"gitlab.com/fynbos/pacioli/db/utils"
	"gitlab.com/fynbos/pacioli/pacioli"
	"gitlab.com/fynbos/pacioli/rpc"
	"gitlab.com/fynbos/tigerbeetle_go"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	baseDbUrl := os.Getenv("DB_URL")
	if baseDbUrl == "" {
		baseDbUrl = "cockroach://pacioli@cockroachdb-public:26257/pacioli?sslmode=verify-full&max_conns=20&max_idle_conns=4"
	}
	tbUrl := os.Getenv("TB_URL")
	if tbUrl == "" {
		tbUrl = "tigerbeetle-0.tigerbeetle.default.svc.cluster.local:80"
	}
	tbClusterID := os.Getenv("TB_CLUSTER_ID")
	if tbClusterID == "" {
		log.Fatalln("TigerBeetle cluster ID not specified.")
	}

	connString, err := utils.InlineSslCreds(
		strings.Replace(baseDbUrl, "cockroach", "postgres", 1), // replace cockroach protocol with postgres so that we can use pq driver.
		"/cockroach-certs/client.pacioli.key",
		"/cockroach-certs/client.pacioli.crt",
		"/cockroach-certs/ca.crt",
	)
	if err != nil {
		log.Fatalln(err)
	}
	db, err := sqlx.Connect("postgres", connString)
	defer db.Close()

	if err != nil {
		log.Fatalln(err)
	}

	clusterID, err := (strconv.ParseUint(tbClusterID, 10, 32))
	tbClient, err := tigerbeetle_go.NewClient(uint32(clusterID), []string{tbUrl})
	if err != nil {
		log.Fatalln(err)
	}
	defer tbClient.Deinit()

	// drive the TB client.
	go func() {
		tick := time.Tick(20 * time.Millisecond)
		for range tick {
			tbClient.Tick()
		}
	}()

	ps, err := pacioli.NewPacioliService(db, tbClient)
	if err != nil {
		log.Fatalln(err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", defaultPort))
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("grpc server: 0.0.0.0:%s", defaultPort)
	server := rpc.NewServer(ps)
	err = server.Serve(listener)
	if err != nil {
		log.Fatalln(err)
	}
}
