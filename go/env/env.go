package env

import (
	"os"
)

var allowedEnvs = []string{
	"prod",    // Live production environment
	"sandbox", // Semi public testing env
	"dev",     // Internal testing environment
	"local",   // For local development
}

func GetEnv() string {
	fynbosEnv := os.Getenv("FYNBOS_ENV")
	if fynbosEnv == "" {
		fynbosEnv = "prod"
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
