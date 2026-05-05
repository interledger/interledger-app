package ops

import (
	"testing"
)

func TestResolvePTIWidgetURLs_ReturnsConfiguredValues(t *testing.T) {
	ConfigureWidgetURLs("https://mockpti.interledger.test/sdk/index.js", "https://mockpti.interledger.test/forms", "test-client-id")

	sdkURL, formsURL := ResolvePTIWidgetURLs()

	if sdkURL != "https://mockpti.interledger.test/sdk/index.js" {
		t.Fatalf("expected configured sdk url, got %q", sdkURL)
	}
	if formsURL != "https://mockpti.interledger.test/forms" {
		t.Fatalf("expected configured forms url, got %q", formsURL)
	}
}

func TestResolvePTIWidgetURLs_ReturnsEmptyValuesWhenUnset(t *testing.T) {
	ConfigureWidgetURLs("", "", "")

	sdkURL, formsURL := ResolvePTIWidgetURLs()

	if sdkURL != "" {
		t.Fatalf("expected empty sdk url, got %q", sdkURL)
	}
	if formsURL != "" {
		t.Fatalf("expected empty forms url, got %q", formsURL)
	}
}
