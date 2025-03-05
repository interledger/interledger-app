package env

import (
	"os"
	"sync"
	"testing"
)

const (
	prodUrl  = "https://interledger.app"
	devUrl   = "https://eu1.fynbos.dev"
	localUrl = "https://interledger.test"
)

var fynbosEnv = "prod"
var once = sync.Once{}

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

func GetEnv() string {
	once.Do(func() {
		fynbosEnv = os.Getenv("FYNBOS_ENV")
		if fynbosEnv == "" {
			fynbosEnv = "prod"
		}
	})
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

func GetUrl() string {
	if IsLocal() {
		return localUrl
	}

	if IsSandbox() || IsDev() {
		return devUrl
	}

	return prodUrl
}

var openPaymentsURL string
var openPaymentsSync sync.Once

func OpenPaymentsURL() string {
	openPaymentsSync.Do(func() {
		openPaymentsURL = os.Getenv("OPEN_PAYMENTS_BASE_URL")
		if openPaymentsURL == "" {
			if IsProd() {
				openPaymentsURL = "https://fynbos.me"
			} else if IsDev() {
				openPaymentsURL = "https://eu1.fynbos.me"
			} else if IsLocal() || IsTest() {
				openPaymentsURL = "https://local.fynbos.me"
			} else {
				openPaymentsURL = "https://eu1.fynbos.me"
			}
		}
	})

	return openPaymentsURL
}

var authURL string
var authURLSync sync.Once

func AuthURL() string {
	authURLSync.Do(func() {
		authURL = os.Getenv("AUTH_BASE_URL")
		if authURL == "" {
			if IsProd() {
				authURL = "https://auth.fynbos.me"
			} else if IsDev() {
				authURL = "https://auth.eu1.fynbos.dev"
			} else if IsLocal() || IsTest() {
				authURL = "https://auth.interledger.test"
			} else {
				authURL = "https://auth.eu1.fynbos.dev"
			}
		}
	})

	return authURL
}

var adminURL string
var adminURLSync sync.Once

func AdminURL() string {
	adminURLSync.Do(func() {
		adminURL = os.Getenv("ADMIN_BASE_URL")
		if adminURL == "" {
			if IsProd() {
				adminURL = "https://admin.mgnt.fynbos.dev"
			} else if IsDev() {
				adminURL = "https://admin-dev.mgnt.fynbos.dev"
			} else if IsLocal() || IsTest() {
				adminURL = "https://admin.interledger.test"
			} else {
				adminURL = "https://admin-dev.mgnt.fynbos.dev"
			}
		}
	})

	return adminURL
}
