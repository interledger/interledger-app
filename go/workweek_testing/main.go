package main

import (
	"context"
	"fmt"

	"gitlab.com/fynbos/proto/backend/v1"
	geov1 "gitlab.com/fynbos/proto/geo/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:8443", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	client := backend.NewBackendServiceClient(conn)

	in := &backend.CreatePaymentV2Request{
		SenderWalletId:  "ecf1a947-8d96-4768-832e-9ddebec4f082",
		SenderAccountId: "e1d681dd-ea17-4a24-a84e-5e0ffa6d6ca8",

		ReceiverWalletAddress: "https://local.ilp.link/radusa2",

		SenderCurrency: &geov1.Currency{
			Amount:      "100",
			CountryCode: "ZAR",
			Asset: &geov1.Asset{
				Code:    "ZAR",
				Scale:   2,
				Numeric: "710",
			},
		},
	}

	res, err := client.CreatePaymentV2(context.TODO(), in)
	if err != nil {
		panic(err)
	}

	fmt.Println(res.Id)

}
