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
	PaymentSentTemplateID TemplateID = `d-aslkdfja0983u45pq0324r` // TODO: This is just an example
)

var templateSubjects = map[TemplateID]string{
	PaymentSentTemplateID: "Fynbos%s:You have received a payment on", // TODO: This is just an example
}
