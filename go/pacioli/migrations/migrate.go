package migrations

import (
	"embed"
	"net/url"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/cockroachdb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed **.*.sql
var fs embed.FS

func Migrate(dbConnectionString string) error {
	// make sure protocol is cockroach
	connectionString := strings.Replace(dbConnectionString, "postgres", "cockroach", -1)
	d, err := iofs.New(fs, ".")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, connectionString)
	if err != nil {
		return err
	}

	// The expected behaviour is for it to return ErrNoChange
	if err := m.Up(); err != migrate.ErrNoChange {
		return err
	}

	return nil
}

// We read the ssl certs into memory and add them inline to the connection string.
// This side steps the file permission issues (postgres spec requires
// private key to have 0600) when mounting secrets into the pod and running
// as non-root user.
// https://github.com/hashicorp/vault/issues/10925
// https://github.com/lib/pq/commit/b9bb726ebf154627a21b50f9ffa4b28c6ed3f4d8
func InlineSslCreds(baseUrl string, pathToPrivateKey string, pathToClientCert string, pathToRootCert string) (string, error) {
	connString := baseUrl + "&sslinline=true"
	sslkeyBytes, err := os.ReadFile(pathToPrivateKey)
	if err != nil {
		return "", err
	}

	sslcertBytes, err := os.ReadFile(pathToClientCert)
	if err != nil {
		return "", err
	}

	sslrootcertBytes, err := os.ReadFile(pathToRootCert)
	if err != nil {
		return "", err
	}

	connString += "&sslinline=true"
	connString += "&sslkey=" + url.QueryEscape(string(sslkeyBytes))
	connString += "&sslcert=" + url.QueryEscape(string(sslcertBytes))
	connString += "&sslrootcert=" + url.QueryEscape(string(sslrootcertBytes))

	return connString, nil
}
