package main

import (
	"github.com/go-playground/validator/v10"
)

type Backends interface {
	Validator() *validator.Validate
}
