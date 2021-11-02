package main
import (
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/cockroachdb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gitlab.com/fynbos/backend/db/utils"
)

func main() {
	baseConnString := os.Getenv("DB_URL")
	if baseConnString == "" {
		baseConnString = "cockroach://backend@cockroachdb-public:26257/backend?sslmode=verify-full&max_conns=20&max_idle_conns=4"
	}

	// We read the ssl certs into memory and add them inline to the connection string.
	// This side steps the file permission issues (postgres spec requires
    // private key to have 0600) when mounting secrets into the pod and running
    // as non-root user.
    // https://github.com/hashicorp/vault/issues/10925
    // https://github.com/lib/pq/commit/b9bb726ebf154627a21b50f9ffa4b28c6ed3f4d8
    connString, err := utils.InlineSslCreds(
    	baseConnString,
    	"/cockroach-certs/client.backend.key",
    	"/cockroach-certs/client.backend.crt",
    	"/cockroach-certs/ca.crt",
    )
    if err != nil {
        log.Fatal(err)
    }

	// ko will automatically set this environment 
	migrationsPath := os.Getenv("KO_DATA_PATH")
	if migrationsPath == "" {
		migrationsPath = "kodata"
	}

	m, err := migrate.New(
		"file://" + migrationsPath,
		connString)
	if err != nil {
		log.Fatal(err)
	}

	// The expected behaviour is for it to return ErrNoChange
	if err := m.Up(); err != migrate.ErrNoChange {
		log.Fatal(err)
	}
}
