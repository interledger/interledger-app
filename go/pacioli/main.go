package main

import (
	"embed"
	"fmt"
	"log"
	"net"
	"os"

	tigerbeetle_go "github.com/coilhq/tigerbeetle-go"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"gitlab.com/fynbos/pacioli/cli"
	"gitlab.com/fynbos/pacioli/healthcheck"
	ledger "gitlab.com/fynbos/pacioli/ledger"
	"gitlab.com/fynbos/pacioli/migrations"
	"gitlab.com/fynbos/pacioli/rpc"
	"gitlab.com/fynbos/pacioli/seed"
)

//go:embed migrations/**.*.sql
var fs embed.FS

func main() {
	args := os.Args
	if len(args) < 2 {
		log.Fatalln("Expected 'start' or 'migrate'.")
	}

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
		args.Fs = &fs

		runInit(args)
	default:
		log.Fatalln("Unknown command: ", command)
	}
}

func runInit(args *cli.InitArgs) {
	// run migrations
	err := migrations.Migrate(args.Fs)
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

	ls, err := ledger.NewService(&ledger.ServiceArgs{
		Db: db,
		Tb: tbClient,
	})
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("tigerbeetle seeding starting")
	err = seed.TigerBeetle(ls, args.TbSeedFile)
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

	ls, err := ledger.NewService(&ledger.ServiceArgs{
		Db: db,
		Tb: tbClient,
	})
	if err != nil {
		log.Fatalln(err)
	}

	hs, err := healthcheck.NewService()
	if err != nil {
		log.Fatalln(err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", args.Port))
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("grpc server: 0.0.0.0:%s", args.Port)
	server := rpc.NewServer(ls, hs)
	err = server.Serve(listener)
	if err != nil {
		log.Fatalln(err)
	}
}
