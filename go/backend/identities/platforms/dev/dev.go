package dev

import (
	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/identities"
)

type dev struct {
	platform identities.Platform
}

func New(platform identities.Platform) *dev {
	return &dev{platform: platform}
}

func (d *dev) VerifyWorkflow() interface{} {
	//TODO implement me
	panic("implement me")
}

func (d *dev) NewVerifyCode() string {
	return uuid.NewString()
}

func (d *dev) VerifyInstructions() string {
	return `In this environment all you need to do to verify is to request it. Enjoy in a NON Production environment.`
}
