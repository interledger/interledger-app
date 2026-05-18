package env

import (
	"flag"
	"strings"
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
	if openPaymentsURL != "" {
		return openPaymentsURL
	}
	if IsProd() {
		return "https://ilp.link"
	} else if IsDev() {
		return "https://sandbox.ilp.link"
	} else if IsLocal() || IsTest() {
		return "https://local.ilp.link"
	}
	return "https://sandbox.ilp.link"
}

var authURL string

func SetAuthURL(url string) {
	authURL = url
}

// TODO -  is this used?
func AuthURL() string {
	if authURL != "" {
		return authURL
	}
	if IsProd() {
		return "https://auth.ilp.link"
	} else if IsDev() {
		return "https://auth.sandbox.ilp.link"
	} else if IsLocal() || IsTest() {
		return "https://auth.local.ilp.link"
	}
	return "https://auth.ilp.link"
}

var adminURL string

func SetAdminURL(url string) {
	adminURL = url
}

func AdminURL() string {
	if adminURL != "" {
		return adminURL
	}
	if IsProd() {
		return "https://admin.interledger.tech"
	} else if IsDev() {
		return "https://admin.sandbox.interledger.tech"
	} else if IsLocal() || IsTest() {
		return "https://admin.interledger.test"
	}
	return "https://admin.sandbox.interledger.tech"
}

func parseList(input string) []string {
	input = strings.ReplaceAll(input, " ", "")
	return strings.Split(input, ",")
}
