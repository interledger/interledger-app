package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"gitlab.com/fynbos/backend/providers/astra/external"
)

func main() {

	_ = os.Setenv("ASTRA_CLIENT_ID", "29b899344bfb462d98fa4dff08ca1fe8")
	_ = os.Setenv("ASTRA_CLIENT_SECRET", "4c1c61166fe04bb686ae4dc8267d2e4c")

	time.Sleep(time.Second * 1)

	cl := external.New(nil)
	/*intentID, err := cl.CreateIntent(context.Background(), external.CreateIntentReq{
		Email:          "barnard+cmdline+3@fynbos.dev",
		Phone:          "+13073819218",
		FirstName:      "Henry",
		LastName:       "Du Toit",
		Address1:       "52 Derry Street",
		Address2:       "",
		City:           "San Antonio",
		State:          "TX",
		PostalCode:     "78202",
		DateOfBirth:    "1989-05-04",
		SocialSecurity: "311401232",
		IPAddress:      "41.71.7.83",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("InentID", intentID)*/

	//intentID := "bbd25c56fca34942981c094d9a573f8a" //  "barnard+cmdline@fynbos.dev"
	//walletID := "71b285aa-15c5-4523-9cfa-3f6ead094fa6" // "barnard+cmdline@fynbos.dev"

	//intentID := "5b886fb0a5b14a10baffc10c1a55ee62" // barnard+cmdline+2@fynbos.dev

	//?authorization_code=I38KSFyRBmNkXxGxbRuBzd5dOmGn7gCfYgV0rWmfo2cRuIzQ&code=I38KSFyRBmNkXxGxbRuBzd5dOmGn7gCfYgV0rWmfo2cRuIzQ

	intentID := "cae1ef7805a944d49824e8b8b4cec74b"
	intent, err := cl.GetIntent(context.Background(), intentID)
	if err != nil {
		panic(err)
	}
	fmt.Println("intent", intent)
	/*
		at, err := cl.CreateAccessToken(context.Background(), intentID, walletID)
		if err != nil {
			panic(err)
		}
		fmt.Println("access token", at)*/

	accessToken := "yB9fjJCHaPxPHoN71gcOVDFenbNFtSC5vzKXPZoJTz"
	refreshToken := "L0BfpN8WlkBXJGAAY9zpm66yV4YUdx1qOxnxfgfV7eq3Th8z"

	at, err := cl.RefreshAccessToken(context.Background(), refreshToken)
	if err != nil {
		panic(err)
	}

	fmt.Println(at)
	accessToken = at.AccessToken

	/**
	Visa Debit Card Example

	Type: Visa
	Card BIN: 400005
	Card #: 4000056655665556
	CVC: Any 3 digits
	Exp. Date: Any future date, e.g. 02/25
	*/
	/*ccr, err := cl.AddCard(context.Background(), accessToken, external.CreateCardArgs{
		CardNumber:       "4000056655665556",
		CardSecurityCode: "131",
		ExpirationDate:   "02/25",
		FirstName:        "Henry",
		LastName:         "Du Toit",
		StreetLine1:      "52 Derry Street",
		StreetLine2:      "",
		City:             "SSan Antonio",
		State:            "TX",
		ZipCode:          "78202",
		AddedByUser:      true,
	})
	if err != nil {
		panic(err)
	}

	crrStr, _ := json.MarshalIndent(ccr, "", "    ")
	fmt.Println(string(crrStr))

	cardID := "882244ff-575f-4739-8fb0-4c5d75da9d2d"
	card, err := cl.LookupCard(context.Background(), accessToken, cardID)
	if err != nil {
		panic(err)
	}

	crrStr, _ := json.MarshalIndent(card, "", "    ")
	fmt.Println(string(crrStr))*/

	/*acc, err := cl.AddAccount(context.Background(), accessToken, external.CreateAccountArgs{
		BankAccountType: external.AccountTypeChecking,
		Name:            "Universal",
		AccountNumber:   "123234558124",
		RoutingNumber:   "125386482",
	})
	if err != nil {
		panic(err)
	}

	accStr, _ := json.MarshalIndent(acc, "", "    ")
	fmt.Println(string(accStr))
	*/
	cardID := "882244ff-575f-4739-8fb0-4c5d75da9d2d"
	accountID := "astra_generic_1ecb81ad52d74259b60ee51603478563"
	/*accountID := "astra_generic_1ecb81ad52d74259b60ee51603478563"
	acc, err := cl.LookupAccount(context.Background(), accessToken, accountID)
	if err != nil {
		panic(err)
	}

	accStr, _ := json.MarshalIndent(acc, "", "    ")
	fmt.Println(string(accStr))*/

	c2a, err := cl.CardToAccount(context.Background(), accessToken, external.CardToAccountArgs{
		Name:                "Test",
		Amount:              10.1,
		ClientCorrelationID: "3d77fb66",
		DebitFeePercent:     0,
		Card: external.Source{
			ID: cardID,
		},
		Account: external.Destination{
			ID:     accountID,
			UserID: intentID,
		},
	})
	if err != nil {
		panic(err)
	}

	c2aStr, _ := json.MarshalIndent(c2a, "", "    ")
	fmt.Println(string(c2aStr))

}
