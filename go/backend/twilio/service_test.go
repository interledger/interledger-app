package twilio

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func TestCreateVerification(t *testing.T) {
	mockMxServer := NewMockServer()
	tws, err := NewService(&ServiceArgs{
		AccountSid:   "testAccountSid",
		AccountToken: "testAccountToken",
		ServiceSid:   "testServiceSid",
		ApiBaseUrl:   mockMxServer.URL,
		Enabled:      true,
	})
	if err != nil {
		t.Fatal(err)
	}

	phoneNumber := "+90555555555"
	res, err := tws.SendVerificationCode(context.Background(), phoneNumber)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, phoneNumber, res.PhoneNumber)
	assert.Equal(t, "pending", res.Status)
}

func TestCheckVerification(t *testing.T) {
	mockMxServer := NewMockServer()
	tws, err := NewService(&ServiceArgs{
		AccountSid:   "testAccountSid",
		AccountToken: "testAccountToken",
		ServiceSid:   "testServiceSid",
		ApiBaseUrl:   mockMxServer.URL,
		Enabled:      true,
	})
	if err != nil {
		t.Fatal(err)
	}

	phoneNumber := "+90555555555"
	res, err := tws.CheckVerificationCode(context.Background(), &CheckVerificationCodeArgs{
		PhoneNumber: phoneNumber,
		Code:        "123456",
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, phoneNumber, res.PhoneNumber)
	assert.Equal(t, "approved", res.Status)
}

func TestServiceDisabled(t *testing.T) {
	tws, err := NewService(&ServiceArgs{
		Enabled: false,
	})
	if err != nil {
		t.Fatal(err)
	}

	phoneNumber := "+90555555555"

	sendRes, err := tws.SendVerificationCode(context.Background(), phoneNumber)
	assert.NoError(t, err)
	assert.Equal(t, phoneNumber, sendRes.PhoneNumber)
	assert.Equal(t, "pending", sendRes.Status)

	checkRes, err := tws.CheckVerificationCode(context.Background(), &CheckVerificationCodeArgs{
		PhoneNumber: phoneNumber,
		Code:        "123456",
	})
	assert.NoError(t, err)
	assert.Equal(t, phoneNumber, checkRes.PhoneNumber)
	assert.True(t, checkRes.IsValid())

	listRes, err := tws.ListSuccessfulVerificationAttempts(context.Background(), ListSuccessfulVerificationAttemptsArgs{
		To: phoneNumber,
	})
	assert.NoError(t, err)
	assert.Len(t, listRes, 1)
	assert.Equal(t, phoneNumber, listRes[0].PhoneNumber)
	assert.True(t, listRes[0].IsValid())
}

func TestNewServiceFailsFastOnInvalidVerifyService(t *testing.T) {
	invalidServiceServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"code":20404,"message":"The requested resource /Services/VA_invalid was not found","status":404}`))
	}))
	defer invalidServiceServer.Close()

	_, err := NewService(&ServiceArgs{
		AccountSid:   "testAccountSid",
		AccountToken: "testAccountToken",
		ServiceSid:   "VA_invalid",
		ApiBaseUrl:   invalidServiceServer.URL,
		Enabled:      true,
	})

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidArgument))
}

func TestNewServiceConfiguresRequestTimeoutOnClonedHTTPClient(t *testing.T) {
	originalTimeout := otelhttp.DefaultClient.Timeout
	t.Cleanup(func() {
		otelhttp.DefaultClient.Timeout = originalTimeout
	})
	otelhttp.DefaultClient.Timeout = 0

	mockMxServer := NewMockServer()
	tws, err := NewService(&ServiceArgs{
		AccountSid:   "testAccountSid",
		AccountToken: "testAccountToken",
		ServiceSid:   "testServiceSid",
		ApiBaseUrl:   mockMxServer.URL,
		Enabled:      true,
	})
	if err != nil {
		t.Fatal(err)
	}

	svc, ok := tws.(*service)
	if !ok {
		t.Fatal("expected concrete *service")
	}

	customClient, ok := svc.twilioClient.RequestHandler.Client.(*CustomClient)
	if !ok {
		t.Fatal("expected concrete *CustomClient")
	}

	assert.Equal(t, time.Duration(0), otelhttp.DefaultClient.Timeout)
	assert.NotSame(t, otelhttp.DefaultClient, customClient.HTTPClient)
	assert.Equal(t, twilioRequestTimeout, customClient.HTTPClient.Timeout)
}

func TestNewServiceTimesOutAgainstBlockingLocalhostServer(t *testing.T) {
	originalTimeout := twilioRequestTimeout
	t.Cleanup(func() {
		twilioRequestTimeout = originalTimeout
	})
	twilioRequestTimeout = 50 * time.Millisecond

	blocked := make(chan struct{})
	blockingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-blocked
	}))
	t.Cleanup(blockingServer.Close)

	start := time.Now()
	_, err := NewService(&ServiceArgs{
		AccountSid:   "testAccountSid",
		AccountToken: "testAccountToken",
		ServiceSid:   "testServiceSid",
		ApiBaseUrl:   blockingServer.URL,
		Enabled:      true,
	})
	duration := time.Since(start)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInternal)
	assert.Less(t, duration, time.Second)
	close(blocked)
}
