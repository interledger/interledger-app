package external_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/clbanning/mxj"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/providers/gmt/external"
)

const exampleRequest = `<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
   <s:Body>
      <InsertTransaction xmlns="http://tempuri.org/">
         <alias>{{Alias}}</alias>
         <user>{{User}}</user>
         <pass>{{Password}}</pass>
         <sender xmlns:a="http://schemas.datacontract.org/2004/07/gmtpay" xmlns:i="http://www.w3.org/2001/XMLSchema-instance">
            <a:SenderAddress>123 TEST ADDRESS</a:SenderAddress>
            <a:SenderBirthDate>2004-07-01T23:59:59.00</a:SenderBirthDate>
            <a:SenderCity>SAN DIEGO</a:SenderCity>
            <a:SenderCountryCode>US</a:SenderCountryCode>
            <a:SenderEmail>TEST@TEST.COM</a:SenderEmail>
            <a:SenderIP>127.0.0.1</a:SenderIP>
            <a:SenderDocExpiration>07/28/2022</a:SenderDocExpiration>
            <a:SenderIdIssuer>CALIFORNIA</a:SenderIdIssuer>
            <a:SenderIdNumber>23424234</a:SenderIdNumber>
            <a:SenderIdType>DRIVER LICENSE</a:SenderIdType>
            <a:SenderLastName>SENDER</a:SenderLastName>
            <a:SenderName>TEST</a:SenderName>
            <a:SenderPhone>619852147</a:SenderPhone>
            <a:SenderResidenceCity>SAN DIEGO</a:SenderResidenceCity>
            <a:SenderResidenceCountryCode>US</a:SenderResidenceCountryCode>
            <a:SenderResidenceState>CA</a:SenderResidenceState>
            <a:SenderState>CA</a:SenderState>
            <a:SenderTrackingNumber>API-0000-0001</a:SenderTrackingNumber>
            <a:SenderZip>91909</a:SenderZip>
         </sender>
        <receiver xmlns:a="http://schemas.datacontract.org/2004/07/gmtpay" xmlns:i="http://www.w3.org/2001/XMLSchema-instance">
                <a:ReceiverActive>true</a:ReceiverActive>
                <a:ReceiverAddress>123 TEST ST</a:ReceiverAddress>
                <a:ReceiverCity>SANTA CLARA</a:ReceiverCity>
                <a:ReceiverCountry>US</a:ReceiverCountry>
                <a:ReceiverLastName>TEST</a:ReceiverLastName>
                <a:ReceiverName>RECEIVER</a:ReceiverName>
                <a:ReceiverPhone></a:ReceiverPhone>
                <a:ReceiverState>CALIFORNIA</a:ReceiverState>
                <a:ReceiverZip></a:ReceiverZip>
            </receiver>
            <transfer xmlns:a="http://schemas.datacontract.org/2004/07/gmtpay" xmlns:i="http://www.w3.org/2001/XMLSchema-instance">
                <a:AmountToReceive>10</a:AmountToReceive>
                <a:BancosNombre></a:BancosNombre>
                <a:BankAccount>12345678901</a:BankAccount>
                <a:BankCode>WFBI</a:BankCode>
                <a:CorrespondentCode>GACH</a:CorrespondentCode>
                <a:DestinationCurrency>USD</a:DestinationCurrency>
                <a:ExchangeRate>1</a:ExchangeRate>
                <a:Fee>0</a:Fee>
                <a:MTSID>1</a:MTSID>
                <a:NetAmount>10.00</a:NetAmount>
                <a:OfficeCode>0</a:OfficeCode>
                <a:OriginalCurrency>USD</a:OriginalCurrency>
                <a:OriginalPaymentMethod>DEBIT</a:OriginalPaymentMethod>
                <a:ServicioCodigo>BD</a:ServicioCodigo>
                <a:SucursalBanco>121042882</a:SucursalBanco>
                <a:ThirdPartyReceipt>F000001</a:ThirdPartyReceipt>
                <a:TipoCuentaCodigo>CHK</a:TipoCuentaCodigo>
            </transfer>
      </InsertTransaction>
   </s:Body>
</s:Envelope>`

func TestRedact(t *testing.T) {
	t.Parallel()

	redactedReq, err := external.Redact(context.Background(), []byte(exampleRequest))
	require.NoError(t, err)
	fmt.Println(string(redactedReq))

	// Check req
	redactedXML, err := mxj.NewMapXml(redactedReq)
	require.NoError(t, err)

	redactedFields, err := redactedXML.ValuesForPath("Envelope.Body.InsertTransaction.sender.SenderIdNumber")
	require.NoError(t, err)
	require.Len(t, redactedFields, 1)
	assert.Equal(t, "*****", redactedFields[0])

	redactedFields, err = redactedXML.ValuesForPath("Envelope.Body.InsertTransaction.pass")
	require.NoError(t, err)
	require.Len(t, redactedFields, 1)
	assert.Equal(t, "*****", redactedFields[0])
}
