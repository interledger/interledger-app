package main

import (
	"context"
	"fmt"

	"gitlab.com/fynbos/backend/providers/gmt/external"
)

func main() {
	cl := external.NewClient()
	resp, err := cl.OfacVerification(context.Background(), external.OfacVerification{
		Alias:     "FYN001",
		User:      "Fynbos_api",
		Pass:      "VUJ6bnkxN2dQVXkwMjZaOA==",
		LastName:  "Jerry",
		FirstName: "Smith",
	})
	if err != nil {
		fmt.Println("ERROR", err)
		return
	}

	fmt.Println(resp)
	fmt.Println(resp.Valid)

	nresp, err := cl.GetNotifications(context.Background(), external.GetNotifications{
		Alias: "FYN001",
		User:  "Fynbos_api",
		Pass:  "VUJ6bnkxN2dQVXkwMjZaOA==",
	})

	if err != nil {
		fmt.Println("ERROR", err)
		return
	}

	fmt.Println(nresp)
}
