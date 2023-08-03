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
	return fmt.Sprintf("%s%s", sub, envStr)
}

func (t TemplateID) String() string {
	return string(t)
}

const (
	ReceiptTemplateID           TemplateID = `d-9b905a8243894b298c4dc2eab502e7d5`
	ReceivedReceiptTemplateID   TemplateID = `d-57a7764bedd8488bb0a65d8f2c79df10`
	DepositSuccessTemplateID    TemplateID = `d-c945aa33f30a4f238a46498c1dba837a`
	WithdrawalInitiated         TemplateID = `d-17430f64081a423c86afa15396df9ee4`
	StatementTemplateID         TemplateID = `d-d1acf02459324ee9aefbeae818354ad5`
	FailedTransactionTemplateID TemplateID = `d-a68f1a97b6e94ec3bd687aee89c942c9`
	ApplicationDenied           TemplateID = "applicationDenied"
	ApplicationApproved         TemplateID = "applicationApproved"
)

var templateSubjects = map[TemplateID]string{
	ReceiptTemplateID:           "Fynbos payment receipt",
	ReceivedReceiptTemplateID:   "You received a payment",
	DepositSuccessTemplateID:    "Deposit received",
	WithdrawalInitiated:         "You've initiated a withdrawal",
	StatementTemplateID:         "Your monthly statement",
	FailedTransactionTemplateID: "Your recent %s was unsuccessful",
	ApplicationDenied:           "Application denied",
	ApplicationApproved:         "Your wallet has been created",
}

type Attachment struct {
	Content     []byte
	ContentType string
	Name        string
}
