package env

import (
	"os"
)

var allowedEnvs = []string{
	"prod",
	"sandbox",
	"dev",
	"testing",
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

func IsTesting() bool {
	return GetEnv() == "testing"
}

func IsDev() bool {
	return GetEnv() == "dev"
}

func IsSandbox() bool {
	return GetEnv() == "sandbox"
}

func IsProd() bool {
	return GetEnv() == "prod"
}
