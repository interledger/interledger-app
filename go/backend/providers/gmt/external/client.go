package external

import (
	"net/http"

	"github.com/hooklift/gowsdl/soap"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func New() Service {
	// TODO: Check URL to use in different environments
	cl := soap.NewClient("http://35.166.119.115/gmtpay/Service1.svc",
		soap.WithHTTPClient(&http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}))

	return &iService1{
		client: cl,
	}
}
