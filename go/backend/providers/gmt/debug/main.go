package main

import (
	"context"
	"encoding/json"
	"fmt"
	external_client "gitlab.com/fynbos/backend/providers/gmt/external"
	"os"
	"time"
)

//func main() {
//	err := os.Setenv("GMT_URL", "https://goldenmoneytransfer.com/gmtpay/Service1.svc")
//	err = os.Setenv("GMT_ALIAS", "CA5320")
//	err = os.Setenv("GMT_USER", "Fynbos_api")
//	err = os.Setenv("GMT_PASSWORD", "V29jZzZzZWp1Y21KU3E5MWo1MUg=")
//	if err != nil {
//		panic(err)
//	}
//
//	time.Sleep(1 * time.Second)
//	ex := external_client.NewClient(nil)
//
//	res, err := ex.InsertTransaction(context.Background(), external_client.InsertTransaction{
//		Sender: &external_client.WsSender{
//			SenderAchAccount:            "325055382270",
//			SenderAchRouting:            "121000358",
//			SenderAchType:               "CHK",
//			SenderAddress:               "785 Market St, San Francisco, CA 94103, USA",
//			SenderAddressStreet:         "1300",
//			SenderBirthDate:             external_client.GMTDate(time.Date(1982, 10, 2, 0, 0, 0, 0, time.UTC)),
//			SenderCity:                  "San Francisco County",
//			SenderCountryCode:           "US",
//			SenderCurrencyCode:          "USD",
//			SenderEmail:                 "adrian@fynbos.dev",
//			SenderGender:                "Male",
//			SenderIP:                    "41.71.7.172",
//			SenderLastName:              "Hope-Bailie",
//			SenderMobile:                "+13073819218",
//			SenderName:                  "Adrian",
//			SenderResidenceAddress:      "785 Market St, San Francisco, CA 94103, USA<",
//			SenderResidenceAddressExtra: "1300",
//			SenderResidenceCity:         "San Francisco County",
//			SenderResidenceCountryCode:  "US",
//			SenderResidenceState:        "US-CA",
//			SenderResidenceZip:          "94103",
//			SenderState:                 "US-CA",
//			SenderTrackingNumber:        "6e652c09-5e89-434a-90d5-3930665f18f7",
//			SenderZip:                   "94103",
//		},
//		Receiver: &external_client.WsReceiver{
//			ReceiverAddress:   "785 Market St, San Francisco, CA 94103, USA",
//			ReceiverBirthDate: external_client.GMTDate(time.Date(1991, 8, 21, 0, 0, 0, 0, time.UTC)),
//			ReceiverCity:      "San Francisco County",
//			ReceiverCountry:   "US",
//			ReceiverCurrency:  "USD",
//			ReceiverEmail:     "matt@fynbos.dev",
//			ReceiverGender:    "Male",
//			ReceiverLastName:  "de Haast",
//			ReceiverMobile:    "+27827096667",
//			ReceiverName:      "Matthew",
//			ReceiverState:     "US-CA",
//			ReceiverZip:       "94103",
//		},
//		Transfer: &external_client.WsTransferInfo{
//			AmountToReceive:       1,
//			CorrespondentCode:     "USCD",
//			BankAccount:           "1234123412341234",
//			DestinationCurrency:   "USD",
//			ExchangeRate:          1,
//			Fee:                   0,
//			OfficeCode:            "0",
//			MTSID:                 1,
//			NetAmount:             1,
//			OriginalCurrency:      "USD",
//			OriginalPaymentMethod: "ACH", // ACH | CHECK | WALLET | CASH | DEBIT | WIRE
//			ReceiverCity:          "San Francisco County",
//			ReceiverState:         "US-CA",
//			SenderID:              0,
//			ServicioCodigo:        "BD",
//		},
//	})
//	if err != nil {
//		panic(err)
//	}
//
//	respJson, err := json.Marshal(res)
//	if err != nil {
//		panic(err)
//	}
//
//	fmt.Println(string(respJson))
//}

// Card 2 Card
func main() {
	err := os.Setenv("GMT_URL", "https://goldenmoneytransfer.com/gmtpay/Service1.svc")
	err = os.Setenv("GMT_ALIAS", "CA5320")
	err = os.Setenv("GMT_USER", "Fynbos_api")
	err = os.Setenv("GMT_PASSWORD", "V29jZzZzZWp1Y21KU3E5MWo1MUg=")

	err = os.Setenv("GMT_TX_URL", "https://goldenmoneytransfer.com/gmtupd/Service1.svc")
	err = os.Setenv("GMT_TX_USER", "Fynbos_payer")
	err = os.Setenv("GMT_TX_PASSWORD", "SzVJOXNTM0k0RXdRbGJoamI=")
	err = os.Setenv("GMT_TX_PARTNER", "87")
	if err != nil {
		panic(err)
	}

	time.Sleep(1 * time.Second)
	ex := external_client.NewClient(nil)

	//res, err := ex.InsertTransaction(context.Background(), external_client.InsertTransaction{
	//	Sender: &external_client.WsSender{
	//		SenderAddress:               "785 Market St, San Francisco, CA 94103, USA",
	//		SenderAddressStreet:         "1300",
	//		SenderBirthDate:             external_client.GMTDate(time.Date(1982, 10, 2, 0, 0, 0, 0, time.UTC)),
	//		SenderCity:                  "San Francisco County",
	//		SenderCountryCode:           "US",
	//		SenderCurrencyCode:          "USD",
	//		SenderEmail:                 "adrian@fynbos.dev",
	//		SenderGender:                "Male",
	//		SenderIP:                    "41.71.7.172",
	//		SenderLastName:              "Hope-Bailie",
	//		SenderMobile:                "+13073819218",
	//		SenderName:                  "Adrian",
	//		SenderResidenceAddress:      "785 Market St, San Francisco, CA 94103, USA<",
	//		SenderResidenceAddressExtra: "1300",
	//		SenderResidenceCity:         "San Francisco County",
	//		SenderResidenceCountryCode:  "US",
	//		SenderResidenceState:        "US-CA",
	//		SenderResidenceZip:          "94103",
	//		SenderState:                 "US-CA",
	//		SenderTrackingNumber:        "0e481e94-5890-4ca9-886e-4097987fdc0e",
	//		SenderZip:                   "94103",
	//	},
	//	Receiver: &external_client.WsReceiver{
	//		ReceiverAddress:   "785 Market St, San Francisco, CA 94103, USA",
	//		ReceiverBirthDate: external_client.GMTDate(time.Date(1991, 8, 21, 0, 0, 0, 0, time.UTC)),
	//		ReceiverCity:      "San Francisco County",
	//		ReceiverCountry:   "US",
	//		ReceiverCurrency:  "USD",
	//		ReceiverEmail:     "matt@fynbos.dev",
	//		ReceiverGender:    "Male",
	//		ReceiverLastName:  "de Haast",
	//		ReceiverMobile:    "+27827096667",
	//		ReceiverName:      "Matthew",
	//		ReceiverState:     "US-CA",
	//		ReceiverZip:       "94103",
	//	},
	//	Transfer: &external_client.WsTransferInfo{
	//		AmountToReceive:       1,
	//		CorrespondentCode:     "USCD",
	//		BankAccount:           "1234123412341234",
	//		DestinationCurrency:   "USD",
	//		ExchangeRate:          1,
	//		Fee:                   0,
	//		OfficeCode:            "0",
	//		MTSID:                 1,
	//		NetAmount:             1,
	//		OriginalCurrency:      "USD",
	//		OriginalPaymentMethod: "DEBIT", // ACH | CHECK | WALLET | CASH | DEBIT | WIRE
	//		ReceiverCity:          "San Francisco County",
	//		ReceiverState:         "US-CA",
	//		SenderID:              0,
	//		ServicioCodigo:        "BD",
	//	},
	//})

	//Update trx
	res, err := ex.UpdateTransactionStatus(context.Background(), external_client.UpdateTransactionStatus{
		Reference:  "GMT000794060548",
		Statuscode: "0",
		Date:       external_client.GMTDate(time.Now()),
	})

	if err != nil {
		panic(err)
	}

	respJson, err := json.Marshal(res)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(respJson))
}
