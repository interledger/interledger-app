package env

import (
	"flag"
	"testing"
)

var applicationURL = "https://interledger.test"

func SetApplicationURL(url string) {
	applicationURL = url
}

func GetUrl() string {
	return applicationURL
}

var fynbosEnv = "prod"
var blockedRegions = []string{}
var allowedWalletIds = []string{}

var allowedEnvs = []string{
	"prod",    // Live production environment
	"sandbox", // Semi public testing env
	"dev",     // Internal testing environment
	"local",   // For local development
	"test",    // Go testing env
}

func SetEnv(t *testing.T, env string) {
	orig := GetEnv()
	fynbosEnv = env
	t.Cleanup(func() {
		fynbosEnv = orig
	})
}

func SetFynbosEnv(v string) {
	fynbosEnv = v
}

func SetAllowedWalletIDs(ids []string) {
	allowedWalletIds = ids
}

func SetBlockedRegions(regions []string) {
	blockedRegions = regions
}

func GetAllowedWalletIds() []string {
	return allowedWalletIds
}

func GetBlockedRegions() []string {
	return blockedRegions
}

func GetEnv() string {
	if fynbosEnv == "" {
		panic("FYNBOS_ENV environment variable is not set")
	}
	var contains bool
	for _, env := range allowedEnvs {
		if env == fynbosEnv {
			contains = true
			break
		}
	}

	if !contains {
		panic("Invalid env=" + fynbosEnv)
	}

	return fynbosEnv
}

// IsLocal returns true if the environment is set up for local development
func IsLocal() bool {
	return GetEnv() == "local"
}

// IsDev returns true if the environment is set up for the internal testing environment
func IsDev() bool {
	return GetEnv() == "dev"
}

// IsSandbox returns true if the environment is set up for the public sandbox environment
func IsSandbox() bool {
	return GetEnv() == "sandbox"
}

// IsProd returns true if the environment is set up for the full public production environment
func IsProd() bool {
	return GetEnv() == "prod"
}

func IsTest() bool {
	return GetEnv() == "test"
}

// IsTestExecution returns true when running under `go test` or when env is explicitly set to test.
func IsTestExecution() bool {
	if IsTest() {
		return true
	}
	return flag.Lookup("test.v") != nil
}

var openPaymentsURL string

func SetOpenPaymentsURL(url string) {
	openPaymentsURL = url
}

func OpenPaymentsURL() string {
	if openPaymentsURL == "" {
		panic("Open Payments URL is not set")
	}
	return openPaymentsURL
}

var authURL string

func SetAuthURL(url string) {
	authURL = url
}

func AuthURL() string {
	if authURL == "" {
		panic("Auth URL is not set")
	}
	return authURL
}

var adminURL string

func SetAdminURL(url string) {
	adminURL = url
}

var personaDashboardURL string

func SetPersonaDashboardURL(url string) {
	personaDashboardURL = url
}

func PersonaDashboardURL() string {
	if personaDashboardURL == "" {
		panic("Persona dashboard URL is not set")
	}
	return personaDashboardURL
}

func AdminURL() string {
	if adminURL == "" {
		panic("Admin URL is not set")
	}
	return adminURL
}
