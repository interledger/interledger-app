package env

import (
	"fmt"
)

var allowedEnvs = []string{
	"prod",
	"sandbox",
	"dev",
	"testing",
}

type FynbosEnv struct {
	env string
}

func NewFynbosEnv(env string) (*FynbosEnv, error) {
	fynbosEnv := env
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
		return nil, fmt.Errorf("Invalid env=%s", fynbosEnv)
	}

	return &FynbosEnv{fynbosEnv}, nil
}

func (e *FynbosEnv) IsSandbox() bool {
	return e.env == "sandbox"
}

func (e *FynbosEnv) IsDev() bool {
	return e.env == "dev"
}

func (e *FynbosEnv) IsTesting() bool {
	return e.env == "testing"
}

func (e *FynbosEnv) IsProd() bool {
	return e.env == "prod"
}
