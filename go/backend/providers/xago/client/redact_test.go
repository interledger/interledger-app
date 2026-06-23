package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedact(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		output string
	}{
		{
			name: "create sub account req",
			input: `{
    "firstName": "Maynard T",
    "lastName": "Keenan",
    "email": "test@interledger.test",
    "mobileNumber": "+27644637031",
    "country": "ZA",
    "nationality": "ZA",
    "identificationDocumentType": "ID Smart Card",
    "identificationNumber": "9811075039085",
    "address": "44 Culver Street",
    "city":"Pretoria",
    "district":"Garsfontein",
    "postalCode":"0081",
    "addressDocumentType":"Bank account statement",
    "dateOfBirth": 19981107
}`,
			output: `{"address":"44 Culver Street","addressDocumentType":"Bank account statement","city":"Pretoria","country":"ZA","dateOfBirth":19981107,"district":"Garsfontein","email":"*****","firstName":"Maynard T","identificationDocumentType":"ID Smart Card","identificationNumber":"*****","lastName":"Keenan","mobileNumber":"*****","nationality":"ZA","postalCode":"0081"}`,
		},
		{
			name: "create sub account resp",
			input: `{
    "accountId": "f5881d30-9e1a-11ed-8865-873a5c1928a3",
    "depositAddress": "rF2132313cCDsdcCaasdaeed",
    "depositTag": "91273812",
    "beneficiaries": [
        {
            "beneficiaryId": "57ecee1c-a9b6-4835-9410-ccc9b009a03e",
            "currencyId": "ZAR",
            "bankName": "Absa",
            "accountNumber": "123456789",
            "accountName": "Xago PTY LTD",
            "depositReference": "AShcd123cd",
            "beneficiaryAction": "rollup"
        }]
}`,
			output: `{"accountId":"f5881d30-9e1a-11ed-8865-873a5c1928a3","beneficiaries":[{"accountName":"Xago PTY LTD","accountNumber":"*****","bankName":"Absa","beneficiaryAction":"rollup","beneficiaryId":"57ecee1c-a9b6-4835-9410-ccc9b009a03e","currencyId":"ZAR","depositReference":"AShcd123cd"}],"depositAddress":"rF2132313cCDsdcCaasdaeed","depositTag":"91273812"}`,
		},
		{
			name: "access token req",
			input: `{
    "policyId": "5e2585a474b0e90012ce8ff1",
    "fields": [
        {
            "fieldName": "apiPublicKey",
            "fieldValue": "8e6a130b-f28c-4b13-a5ba-713146b11475"
        },
        {
            "fieldName": "apiSecretKey",
            "fieldValue": "c173002c-1155-4eb8-81c8-8616baad8b49"
        }
    ],
    "multiFactor": false
}`,
			output: `{"fields":[{"fieldName":"apiPublicKey","fieldValue":"*****"},{"fieldName":"apiSecretKey","fieldValue":"*****"}],"multiFactor":false,"policyId":"5e2585a474b0e90012ce8ff1"}`,
		},
		{
			name:   "access token resp",
			input:  `{"tokenValue":"fluffy_bunny_records","tokenType":"jwt","identityInfo":{"identityType":"Business","activeStatus":true,"adminVerified":true,"aiqEnabled":false,"statusLocked":false,"referralCode":"leon","parentAdmin":true,"parentId":"65379d2f129d2f14a796314e","superParentId":"xago"}}`,
			output: `{"identityInfo":{"activeStatus":true,"adminVerified":true,"aiqEnabled":false,"identityType":"Business","parentAdmin":true,"parentId":"65379d2f129d2f14a796314e","referralCode":"leon","statusLocked":false,"superParentId":"xago"},"tokenType":"jwt","tokenValue":"*****"}`,
		},
		{
			name: "create beneficiary req",
			input: `{
    "name": "Leon Kowalski I",
    "scope": "bank",
    "currencyCode": "ZAR",
    "accountNumber": "121045537",
    "branchCode": "1244556",
    "bankName": "ABSA",
    "bankCountry": "ZA",
    "accountName": "Leon",
    "bankBeneficiaryType": "IBAN",
    "reference": "Leon A",
    "IBAN": "12345",
    "BIC": "ABC123",
    "beneficiaryPhysicalAddress": "16 Breda Street",
    "beneficiaryDistrict": "Gardens",
    "beneficiaryCity": "Cape Town",
    "beneficiaryCountry": "ZA",
    "beneficiaryPostalCode": "8000",
    "beneficiaryAddress": "16 Breda Street",
    "accountType": "typeAccountNumber"
}`,
			output: `{"BIC":"*****","IBAN":"*****","accountName":"Leon","accountNumber":"*****","accountType":"typeAccountNumber","bankBeneficiaryType":"IBAN","bankCountry":"ZA","bankName":"ABSA","beneficiaryAddress":"16 Breda Street","beneficiaryCity":"Cape Town","beneficiaryCountry":"ZA","beneficiaryDistrict":"Gardens","beneficiaryPhysicalAddress":"16 Breda Street","beneficiaryPostalCode":"8000","branchCode":"1244556","currencyCode":"ZAR","name":"Leon Kowalski I","reference":"Leon A","scope":"bank"}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := Redact(context.Background(), []byte(tc.input))
			require.NoError(t, err)
			assert.Equal(t, tc.output, string(out))
		})
	}
}
