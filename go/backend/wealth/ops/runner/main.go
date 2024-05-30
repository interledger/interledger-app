package main

import (
	"context"
	"fmt"

	"gitlab.com/fynbos/backend/wealth/ops"
)

func main() {
	/*var username, password string
	fmt.Println("Username: ")
	fmt.Scanln(&username)
	fmt.Println("Password:")
	fmt.Scanln(&password)
	valid, err := ops.Login(context.Background(), username, password)
	fmt.Println("is valid", valid, "err", err)
	if err != nil {
		return
	}
	_, err = ops.GetTFSATransactions(context.Background(), valid)
	fmt.Println("err", err)
	if err != nil {
		return
	}*/
	fmt.Println(ops.ParseTXHistory(context.Background(), "/home/barnard/Downloads/transactions.xlsx"))
}
