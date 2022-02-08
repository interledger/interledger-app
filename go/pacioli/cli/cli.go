package cli

import (
	"embed"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	ledger "gitlab.com/fynbos/pacioli/ledger"
	"gitlab.com/fynbos/pacioli/migrations"
	"gitlab.com/fynbos/tigerbeetle_go"
)

type InitArgs struct {
	BackendLedgerCode  uint16
	DbConnectionString string
	TbClusterID        uint32
	TbUrls             []string
	Fs                 *embed.FS
}

// We choose to specify the ledger's for the consuming services here so that
// it can be configured from deploy config.
func ParseInitArgs() (*InitArgs, error) {
	baseDbUrl := os.Getenv("DB_URL")
	if baseDbUrl == "" {
		baseDbUrl = "cockroach://pacioli@cockroachdb-public:26257/pacioli?sslmode=verify-full&max_conns=20&max_idle_conns=4"
	}
	connString, err := migrations.InlineSslCreds(
		strings.Replace(baseDbUrl, "cockroach", "postgres", 1), // replace cockroach protocol with postgres so that we can use pq driver.
		"/cockroach-certs/client.pacioli.key",
		"/cockroach-certs/client.pacioli.crt",
		"/cockroach-certs/ca.crt",
	)
	if err != nil {
		return nil, err
	}

	tbUrl := os.Getenv("TB_URL")
	if tbUrl == "" {
		tbUrl = "tigerbeetle-0.tigerbeetle.default.svc.cluster.local"
	}
	// checking if tbUrl is a host name and converting it to an ip address.
	// here till this is supported by the TB client.
	tbUrls, err := parseTburl(tbUrl)
	if err != nil {
		return nil, err
	}

	tbClusterID := os.Getenv("TB_CLUSTER_ID")
	if tbClusterID == "" {
		return nil, errors.New("TigerBeetle cluster ID not specified.")
	}
	parsedTbClusterID, err := strconv.ParseUint(tbClusterID, 10, 32)
	if err != nil {
		return nil, err
	}

	backendLedgerCodeString := os.Getenv("BACKEND_LEDGER_CODE")
	if backendLedgerCodeString == "" {
		return nil, errors.New("BACKEND_LEDGER_CODE is required.")
	}
	backendLedgerCodeUInt, err := strconv.ParseUint(backendLedgerCodeString, 10, 16)
	if err != nil {
		return nil, errors.New("BACKEND_LEDGER_CODE must be a uint16.")
	}

	return &InitArgs{
		DbConnectionString: connString,
		BackendLedgerCode:  uint16(backendLedgerCodeUInt), // ParseUint will check the max size.
		TbUrls:             tbUrls,
		TbClusterID:        uint32(parsedTbClusterID),
	}, nil
}

func Init(args *InitArgs) error {
	// run migrations
	err := migrations.Migrate(args.Fs)
	if err != nil {
		return err
	}

	// insert ledgers for consuming services
	db, err := sqlx.Connect("postgres", args.DbConnectionString)
	defer db.Close()

	tbClient, err := tigerbeetle_go.NewClient(args.TbClusterID, args.TbUrls)
	if err != nil {
		log.Fatalln(err)
	}
	defer tbClient.Deinit()

	// drive the TB client.
	// TODO: update to official client when it lands.
	go func() {
		tick := time.Tick(20 * time.Millisecond)
		for range tick {
			tbClient.Tick()
		}
	}()

	ls, err := ledger.NewLedgerService(db, tbClient)
	if err != nil {
		log.Fatalln(err)
	}

	// TODO: rework configuration when grpc auth is introduced.
	_, err = ls.CreateLedger("backend-usd", args.BackendLedgerCode)
	if err != nil {
		switch err.(type) {
		case ledger.ErrDuplicate:
			// do nothing.
		default:
			return err
		}
	}

	return nil
}

type StartArgs struct {
	Port               string
	DbConnectionString string
	TbUrls             []string
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
		tbUrl = "tigerbeetle-0.tigerbeetle.default.svc.cluster.local"
	}

	// checking if tbUrl is a host name and converting it to an ip address.
	// here till this is supported by the TB client.
	tbUrls, err := parseTburl(tbUrl)
	if err != nil {
		return nil, err
	}

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
		TbUrls:             tbUrls,
		TbClusterID:        uint32(parsedTbClusterID),
	}, nil
}

func parseTburl(url string) ([]string, error) {
	if url == "" {
		return nil, errors.New("Tb url must be specified.")
	}
	split := strings.Split(url, ":")
	host := split[0]
	port := ":80"
	if len(split) > 1 {
		port = split[1]
	}

	tbUrls := []string{}
	tbIp := net.ParseIP(host)
	if tbIp == nil { // not an ip address
		ips, err := net.LookupIP(host)
		if err != nil {
			return nil, err
		}

		for _, ip := range ips {
			tbUrls = append(tbUrls, ip.String()+port)
		}
	} else {
		tbUrls = append(tbUrls, url)
	}

	for i, url := range tbUrls {
		fmt.Println("tigerbeetle-", i, " url:", url)
	}

	return tbUrls, nil
}
