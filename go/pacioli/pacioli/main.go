package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/go-playground/validator/v10"

	"gitlab.com/fynbos/pacioli/rpcserver"

	tigerbeetle_go "github.com/coilhq/tigerbeetle-go"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"gitlab.com/fynbos/pacioli/cli"
	"gitlab.com/fynbos/pacioli/healthcheck"
	"gitlab.com/fynbos/pacioli/migrations"
	"gitlab.com/fynbos/pacioli/seed"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		log.Fatalln("Expected 'start' or 'migrate'.")
	}

	// Set the timezone globally
	time.Local = time.UTC

	command := args[1]
	switch command {
	case "start":
		args, err := cli.ParseStartArgs()
		if err != nil {
			log.Fatalln(err)
		}

		start(args)
	case "init":
		args, err := cli.ParseInitArgs()
		if err != nil {
			log.Fatalln(err)
		}

		runInit(args)
	default:
		log.Fatalln("Unknown command: ", command)
	}
}

func runInit(args *cli.InitArgs) {
	// run migrations
	err := migrations.Migrate()
	if err != nil {
		log.Fatalln(err)
	}

	db, err := sqlx.Connect("postgres", args.DbConnectionString)
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(db)

	if err != nil {
		log.Fatalln(err)
	}

	tbClient, err := tigerbeetle_go.NewClient(args.TbClusterID, args.TbUrls, 100)
	if err != nil {
		log.Fatalln(err)
	}
	defer tbClient.Close()

	b := NewBackends(db, tbClient)

	log.Println("tigerbeetle seeding starting")
	err = seed.TigerBeetle(b, args.TbSeedFile)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("tigerbeetle seeding complete")
}

func start(args *cli.StartArgs) {
	db, err := sqlx.Connect("postgres", args.DbConnectionString)
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(db)

	if err != nil {
		log.Fatalln(err)
	}

	tbClient, err := tigerbeetle_go.NewClient(args.TbClusterID, args.TbUrls, 1000)
	if err != nil {
		log.Fatalln(err)
	}
	defer tbClient.Close()

	b := NewBackends(db, tbClient)

	hs, err := healthcheck.NewService()
	if err != nil {
		log.Fatalln(err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", args.Port))
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("grpc server: 0.0.0.0:%s", args.Port)
	server := rpcserver.NewServer(b, hs)
	err = server.Serve(listener)
	if err != nil {
		log.Fatalln(err)
	}
}

type Backends interface {
	DB() *sqlx.DB
	TigerBeetle() tigerbeetle_go.Client
	Validator() *validator.Validate
}

var _ Backends = backends{}

type backends struct {
	db  *sqlx.DB
	tbc tigerbeetle_go.Client
	val *validator.Validate
}

func NewBackends(db *sqlx.DB, tbc tigerbeetle_go.Client) Backends {
	return &backends{
		db:  db,
		tbc: tbc,
		val: validator.New(),
	}
}

func (b backends) DB() *sqlx.DB {
	return b.db
}

func (b backends) TigerBeetle() tigerbeetle_go.Client {
	return b.tbc
}

func (b backends) Validator() *validator.Validate {
	return b.val
}
