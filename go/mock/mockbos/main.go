package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/interledger/interledger-app/go/mock/mockbos/db"
	"github.com/interledger/interledger-app/go/mock/mockbos/pti"
	"github.com/interledger/interledger-app/go/mock/mockbos/xago"
	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()
	fmt.Println("starting mockbos")

	dbConnStr := os.Getenv("DB_URL")
	if dbConnStr == "" {
		dbConnStr = "postgres://postgres:password@localhost:5432?sslmode=disable"
	}

	conn, err := pgx.Connect(context.Background(), dbConnStr)
	if err != nil {
		fmt.Println("Unable to connection to database:", err)
		return
	}
	defer conn.Close(ctx)
	fmt.Println("connected to database")

	err = db.Migrate(ctx, dbConnStr, conn)
	if err != nil {
		fmt.Println("Unable to migrate database:", err)
		return
	}

	sDb := db.New(conn)
	xs := xago.New(sDb)
	ps := pti.New(conn)

	// setup a new http server using a chi router
	router := chi.NewRouter()
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	router.Post("/xago/v1/company/accounts", xs.CreateSubAccount())
	router.Post("/xago/v1/company/accounts/testdeposit", xs.CreateDeposit())
	router.Get("/xago/v1/currencies", xs.GetBankAccounts())
	router.Post("/xago/v1/beneficiaries", xs.AddBeneficiary())
	router.Post("/xago/v1/transfers", xs.CreateTransaction())
	router.Get("/xago/v1/transactions", xs.GetTransaction())
	router.Get("/xago/v1/company/transactions", xs.ListDepositTransactions())
	router.Post("/xago/v1/login", xs.CreateLogin())

	router.Route("/pti", ps.Register)

	// start the server
	err = http.ListenAndServe(":8080", router)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("listening on :8080")

	// safely shutdown the server
}
