package main

import (
	"embed"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"gitlab.com/fynbos/pacioli/cli"
	"gitlab.com/fynbos/pacioli/healthcheck"
	ledger "gitlab.com/fynbos/pacioli/ledger"
	"gitlab.com/fynbos/pacioli/rpc"
	"gitlab.com/fynbos/tigerbeetle_go"
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

		err = cli.Init(args)
		if err != nil {
			log.Fatalln(err)
		}
	default:
		log.Fatalln("Unknown command: ", command)
	}
}

func start(args *cli.StartArgs) {
	db, err := sqlx.Connect("postgres", args.DbConnectionString)
	defer db.Close()

	if err != nil {
		log.Fatalln(err)
	}

	tbClient, err := tigerbeetle_go.NewClient(args.TbClusterID, args.TbUrls)
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

	ls, err := ledger.NewLedgerService(db, tbClient)
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
