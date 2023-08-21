package client

import (
	"context"
	"testing"
	"time"

	"gitlab.com/fynbos/backend/providers/gmt/external"

	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/env"
)

func TestClient(t *testing.T) {
	t.Skip("skipping gmt external client integration test.")
	env.SetEnv(t, "local")
	client := NewClient(nil)

	dob, err := time.Parse("2006-01-2", "2004-07-01")
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.ComplianceCheck(context.Background(), external.ComplianceCheck{
		Sender: &external.WsSender{
			SenderAchAccount:           "5432154321",
			SenderAchRouting:           "000300258",
			SenderAchType:              "SV",
			SenderAddress:              "123 TEST ADDRESS",
			SenderBirthDate:            external.GMTDate(dob),
			SenderCity:                 "SAN DIEGO",
			SenderCountryCode:          "US",
			SenderCurrencyCode:         "USD",
			SenderEmail:                "TEST@TEST.COM",
			SenderIP:                   "127.0.0.1",
			SenderLastName:             "SENDER",
			SenderName:                 "TEST",
			SenderPhone:                "619852147",
			SenderResidenceCity:        "SAN DIEGO",
			SenderResidenceCountryCode: "US",
			SenderResidenceState:       "CA",
			SenderState:                "CA",
			SenderZip:                  "91909",
			SenderTrackingNumber:       "API-0000-0001",
		},
		Receiver: &external.WsReceiver{
			ReceiverActive:   true,
			ReceiverAddress:  "123 TEST ST",
			ReceiverCity:     "SANTA CLARA",
			ReceiverCountry:  "US",
			ReceiverLastName: "TEST",
			ReceiverName:     "RECEIVER",
			ReceiverCurrency: "USD",
			ReceiverPhone:    "",
			ReceiverState:    "CALIFORNIA",
			ReceiverZip:      "",
		},
		Transfer: &external.WsTransferInfo{
			AmountToReceive:       10,
			BankAccount:           "12345678901",
			BankCode:              "WFBI",
			CorrespondentCode:     "GACH",
			DestinationCurrency:   "USD",
			ExchangeRate:          1,
			Fee:                   0,
			MTSID:                 1,
			NetAmount:             10,
			OfficeCode:            "0",
			OriginalCurrency:      "USD",
			OriginalPaymentMethod: "ACH", // ACH | CHECK | WALLET | CASH | DEBIT | WIRE
			ServicioCodigo:        "BD",
			SucursalBanco:         "121042882", // Bank branch or routing number
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "OK", resp.Status)
	assert.True(t, resp.Valid)

	resp, err = client.InsertTransaction(context.Background(), external.InsertTransaction{
		Sender: &external.WsSender{
			SenderAchAccount:           "5432154321",
			SenderAchRouting:           "000300258",
			SenderAchType:              "SV",
			SenderAddress:              "123 TEST ADDRESS",
			SenderBirthDate:            external.GMTDate(dob),
			SenderCity:                 "SAN DIEGO",
			SenderCountryCode:          "US",
			SenderCurrencyCode:         "USD",
			SenderEmail:                "TEST@TEST.COM",
			SenderIP:                   "127.0.0.1",
			SenderLastName:             "SENDER",
			SenderName:                 "TEST",
			SenderPhone:                "619852147",
			SenderResidenceCity:        "SAN DIEGO",
			SenderResidenceCountryCode: "US",
			SenderResidenceState:       "CA",
			SenderState:                "CA",
			SenderZip:                  "91909",
			SenderTrackingNumber:       "API-0000-0001",
		},
		Receiver: &external.WsReceiver{
			ReceiverActive:   true,
			ReceiverAddress:  "123 TEST ST",
			ReceiverCity:     "SANTA CLARA",
			ReceiverCountry:  "US",
			ReceiverLastName: "TEST",
			ReceiverName:     "RECEIVER",
			ReceiverCurrency: "USD",
			ReceiverPhone:    "",
			ReceiverState:    "CALIFORNIA",
			ReceiverZip:      "",
		},
		Transfer: &external.WsTransferInfo{
			AmountToReceive:       10,
			BankAccount:           "12345678901",
			BankCode:              "WFBI",
			CorrespondentCode:     "GACH",
			DestinationCurrency:   "USD",
			ExchangeRate:          1,
			Fee:                   0,
			MTSID:                 1,
			NetAmount:             10,
			OfficeCode:            "0",
			OriginalCurrency:      "USD",
			OriginalPaymentMethod: "ACH", // ACH | CHECK | WALLET | CASH | DEBIT | WIRE
			ServicioCodigo:        "BD",
			SucursalBanco:         "121042882", // Bank branch or routing number
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "Created", resp.Status)
	assert.True(t, resp.Valid)
	assert.NotEmpty(t, resp.Password)
	assert.NotEmpty(t, resp.Receipt)
}
