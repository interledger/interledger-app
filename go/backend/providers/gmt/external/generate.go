package external

/* Usage instructions :
Do not RUN!!!

If you do need to run it....

* Replace all the soap.XSDDateTime with GMTDate to marshall in the expected format

*/
////go:generate gowsdl -o=generated.go  http://35.166.119.115/gmtpay/Service1.svc?wsdl
