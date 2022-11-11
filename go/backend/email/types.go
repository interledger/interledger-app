package email

import (
	"fmt"

	"gitlab.com/fynbos/env"
)

type TemplateID string

func (t TemplateID) IsValid() bool {
	_, ok := templateSubjects[t]
	return ok
}

func (t TemplateID) Subject() string {
	sub := templateSubjects[t]
	envStr := ""
	if !env.IsProd() {
		envStr = " [" + env.GetEnv() + "]"
	}
	return fmt.Sprintf(sub, envStr)
}

func (t TemplateID) String() string {
	return string(t)
}

const (
	ReceiptTemplateID TemplateID = `d-9b905a8243894b298c4dc2eab502e7d5`
)

var templateSubjects = map[TemplateID]string{
	ReceiptTemplateID: "Fynbos payment receipt.", // TODO: This is just an example
}
