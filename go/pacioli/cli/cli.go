package cli

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type InitArgs struct {
	DbConnectionString string
	TbClusterID        uint32
	TbUrls             []string
	TbSeedFile         string
}

// We choose to specify the ledger's for the consuming services here so that
// it can be configured from deploy config.
func ParseInitArgs() (*InitArgs, error) {
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		dbUrl = "cockroach://backend@cockroachdb-public:26257/pacioli?sslmode=verify-full&sslrootcert=/cockroach-certs/ca.crt&sslkey=/cockroach-certs/client.backend.key&sslcert=/cockroach-certs/client.backend.crt&max_conns=20&max_idle_conns=4"
	}

	tbUrl := os.Getenv("TB_URL")
	if tbUrl == "" {
		tbUrl = "tigerbeetle-0.tigerbeetle.default.svc.cluster.local"
	}
	// checking if tbUrl is a host name and converting it to an ip address.
	// here till this is supported by the TB client.
	tbUrls, err := ParseTburl(tbUrl)
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

	tbSeedFile := os.Getenv("TB_SEED_FILE")
	if tbSeedFile == "" {
		return nil, errors.New("tigerbeetle seed file not specified")
	}

	return &InitArgs{
		DbConnectionString: dbUrl,
		TbUrls:             tbUrls,
		TbClusterID:        uint32(parsedTbClusterID),
		TbSeedFile:         tbSeedFile,
	}, nil
}

type StartArgs struct {
	Port               string
	DbConnectionString string
	TbUrls             []string
	TbClusterID        uint32
	LogLevel           string
	LogOutputPath      string
}

func ParseStartArgs() (*StartArgs, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "443" // default port for grpc
	}
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		dbUrl = "cockroach://backend@cockroachdb-public:26257/pacioli?sslmode=verify-full&sslrootcert=/cockroach-certs/ca.crt&sslkey=/cockroach-certs/client.backend.key&sslcert=/cockroach-certs/client.backend.crt&max_conns=20&max_idle_conns=4"
	}
	tbUrl := os.Getenv("TB_URL")
	if tbUrl == "" {
		tbUrl = "tigerbeetle-0.tigerbeetle.default.svc.cluster.local"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	logOutputPath := os.Getenv("LOG_OUTPUT_PATH")
	if logOutputPath == "" {
		logOutputPath = "stderr"
	}

	// checking if tbUrl is a host name and converting it to an ip address.
	// here till this is supported by the TB client.
	tbUrls, err := ParseTburl(tbUrl)
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

	return &StartArgs{
		Port:               port,
		DbConnectionString: dbUrl,
		TbUrls:             tbUrls,
		TbClusterID:        uint32(parsedTbClusterID),
		LogLevel:           logLevel,
		LogOutputPath:      logOutputPath,
	}, nil
}

func ParseTburl(url string) ([]string, error) {
	fmt.Println("printed url: ", url)
	if url == "" {
		return nil, errors.New("Tb url must be specified.")
	}
	split := strings.Split(url, ":")
	host := split[0]
	port := "8080"
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
			if ip.To4() != nil {
				tbUrls = append(tbUrls, ip.String()+":"+port)
			}
		}
	} else {
		tbUrls = append(tbUrls, url)
	}

	for i, url := range tbUrls {
		fmt.Println("tigerbeetle-", i, " url:", url)
	}

	return tbUrls, nil
}
