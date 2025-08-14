package pacioli

type ConfigureLedgerArgs struct {
	ID    uint32
	Name  string `validate:"required"`
	Asset string `validate:"required"`
	Scale uint8  `validate:"gt=0"`
}

type Ledger struct {
	ID        uint32 `json:"id"`
	Name      string `json:"name"`
	Asset     string `json:"string"`
	Scale     uint8  `json:"scale"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

type ConfigureAccountArgs struct {
	ID                         string `validate:"required,uuid4"`
	LedgerID                   uint32 `validate:"required"`
	Code                       uint16 `validate:"required"`
	DebitsMustNotExceedCredits bool
	CreditsMustNotExceedDebits bool
}

type Account struct {
	ID                         string `db:"id"`
	LedgerID                   uint32 `db:"ledger_id"`
	Code                       uint16 `db:"code"`
	DebitsPending              uint64 `db:"debits_pending"`
	DebitsPosted               uint64 `db:"debits_posted"`
	CreditsPending             uint64 `db:"credits_pending"`
	CreditsPosted              uint64 `db:"credits_posted"`
	DebitsMustNotExceedCredits bool   `db:"debits_must_not_exceed_credits"`
	CreditsMustNotExceedDebits bool   `db:"credits_must_not_exceed_debits"`
}

//go:generate stringer -type=TransferState -trimprefix=Transfer
type TransferState int

const (
	TransferStateUnknown  TransferState = 0
	TransferStatePending  TransferState = 1
	TransferStateVoided   TransferState = 2
	TransferStateTimeout  TransferState = 3
	TransferStatePosted   TransferState = 4
	TransferStateSentinel TransferState = 5 // Sanity check value
)

func (ts TransferState) IsValid() bool {
	return ts > TransferStateUnknown && ts < TransferStateSentinel
}

type Transfer struct {
	ID              string        `db:"id"`
	LedgerID        uint32        `db:"ledger_id"` // this field is coming soon to a TigerBeetle near you.
	DebitAccountID  string        `db:"debit_account_id"`
	CreditAccountID string        `db:"credit_account_id"`
	Amount          uint64        `db:"amount"`
	ProviderFee     uint64        `db:"provider_fee"`
	Code            uint16        `db:"code"`
	State           TransferState `db:"state"`
	Timeout         uint64
}

type CreateTransferArgs struct {
	ID              string `validate:"required,uuid4"`
	Amount          uint64 `validate:"gt=0"`
	ProviderFee     uint64
	DebitAccountID  string `validate:"required,uuid4"`
	CreditAccountID string `validate:"required,uuid4"`
	Pending         bool
	Code            uint16 `validate:"required"`
	Timeout         uint64
	Ledger          uint32 `validate:"required"`
}

type LedgerResult struct {
	Index uint32
	Code  LedgerResultCode
}

//go:generate stringer -type=LedgerResultCode -trimprefix=Ledger
type LedgerResultCode uint32

const (
	LedgerOK                       LedgerResultCode = 0
	LedgerExistsWithDifferentName  LedgerResultCode = 1
	LedgerExistsWithDifferentAsset LedgerResultCode = 2
	LedgerExistsWithDifferentScale LedgerResultCode = 3
)

type AccountResult struct {
	Index uint32
	Code  AccountResultCode
}

//go:generate stringer -type=AccountResultCode -trimprefix=Account
type AccountResultCode uint32

const (
	AccountOK                                AccountResultCode = 0
	AccountLinkedEventFailed                 AccountResultCode = 1
	AccountReservedFlag                      AccountResultCode = 2
	AccountReservedField                     AccountResultCode = 3
	AccountIdMustNotBeZero                   AccountResultCode = 4
	AccountIdMustNotBeMaxInt                 AccountResultCode = 5
	AccountLedgerMustNotBeZero               AccountResultCode = 6
	AccountCodeMustNotBeZero                 AccountResultCode = 7
	AccountMutuallyExclusiveFlags            AccountResultCode = 8
	AccountOverflowsDebits                   AccountResultCode = 9
	AccountOverflowsCredits                  AccountResultCode = 10
	AccountExceedsCredits                    AccountResultCode = 11
	AccountExceedsDebits                     AccountResultCode = 12
	AccountExistsWithDifferentFlags          AccountResultCode = 13
	AccountExistsWithDifferentUserData       AccountResultCode = 14
	AccountExistsWithDifferentLedger         AccountResultCode = 15
	AccountExistsWithDifferentCode           AccountResultCode = 16
	AccountExistsWithDifferentDebitsPending  AccountResultCode = 17
	AccountExistsWithDifferentDebitsPosted   AccountResultCode = 18
	AccountExistsWithDifferentCreditsPending AccountResultCode = 19
	AccountExistsWithDifferentCreditsPosted  AccountResultCode = 20
	AccountExists                            AccountResultCode = 21
	AccountLedgerDoesNotExist                AccountResultCode = 404
)

type TransferResult struct {
	Index uint32
	Code  TransferResultCode
}

//go:generate stringer -type=TransferResultCode -trimprefix=Transfer
type TransferResultCode uint32

const (
	TransferLinkedEventFailed                          TransferResultCode = 1
	TransferReservedFlag                               TransferResultCode = 2
	TransferReservedField                              TransferResultCode = 3
	TransferIdMustNotBeZero                            TransferResultCode = 4
	TransferIdMustNotBeMaxInt                          TransferResultCode = 5
	TransferDebitAccountIdMustNotBeZero                TransferResultCode = 6
	TransferDebitAccountIdMustNotBeMaxInt              TransferResultCode = 7
	TransferCreditAccountIdMustNotBeZero               TransferResultCode = 8
	TransferCreditAccountIdMustNotBeMaxInt             TransferResultCode = 9
	TransferAccountsMustBeDifferent                    TransferResultCode = 10
	TransferPendingIdMustBeZero                        TransferResultCode = 11
	TransferPendingTransferMustTimeout                 TransferResultCode = 12
	TransferLedgerMustNotBeZero                        TransferResultCode = 13
	TransferCodeMustNotBeZero                          TransferResultCode = 14
	TransferAmountMustNotBeZero                        TransferResultCode = 15
	TransferDebitAccountNotFound                       TransferResultCode = 16
	TransferCreditAccountNotFound                      TransferResultCode = 17
	TransferAccountsMustHaveTheSameLedger              TransferResultCode = 18
	TransferTransferMustHaveTheSameLedgerAsAccounts    TransferResultCode = 19
	TransferExistsWithDifferentFlags                   TransferResultCode = 20
	TransferExistsWithDifferentDebitAccountId          TransferResultCode = 21
	TransferExistsWithDifferentCreditAccountId         TransferResultCode = 22
	TransferExistsWithDifferentUserData                TransferResultCode = 23
	TransferExistsWithDifferentPendingId               TransferResultCode = 24
	TransferExistsWithDifferentTimeout                 TransferResultCode = 25
	TransferExistsWithDifferentCode                    TransferResultCode = 26
	TransferExistsWithDifferentAmount                  TransferResultCode = 27
	TransferExists                                     TransferResultCode = 28
	TransferOverflowsDebitsPending                     TransferResultCode = 29
	TransferOverflowsCreditsPending                    TransferResultCode = 30
	TransferOverflowsDebitsPosted                      TransferResultCode = 31
	TransferOverflowsCreditsPosted                     TransferResultCode = 32
	TransferOverflowsDebits                            TransferResultCode = 33
	TransferOverflowsCredits                           TransferResultCode = 34
	TransferExceedsCredits                             TransferResultCode = 35
	TransferExceedsDebits                              TransferResultCode = 36
	TransferCannotPostAndVoidPendingTransfer           TransferResultCode = 37
	TransferPendingTransferCannotPostOrVoidAnother     TransferResultCode = 38
	TransferTimeoutReservedForPendingTransfer          TransferResultCode = 39
	TransferPendingIdMustNotBeZero                     TransferResultCode = 40
	TransferPendingIdMustNotBeMaxInt                   TransferResultCode = 41
	TransferPendingIdMustBeDifferent                   TransferResultCode = 42
	TransferPendingTransferNotFound                    TransferResultCode = 43
	TransferPendingTransferNotPending                  TransferResultCode = 44
	TransferPendingTransferHasDifferentDebitAccountId  TransferResultCode = 45
	TransferPendingTransferHasDifferentCreditAccountId TransferResultCode = 46
	TransferPendingTransferHasDifferentLedger          TransferResultCode = 47
	TransferPendingTransferHasDifferentCode            TransferResultCode = 48
	TransferExceedsPendingTransferAmount               TransferResultCode = 49
	TransferPendingTransferHasDifferentAmount          TransferResultCode = 50
	TransferPendingTransferAlreadyPosted               TransferResultCode = 51
	TransferPendingTransferAlreadyVoided               TransferResultCode = 52
	TransferPendingTransferExpired                     TransferResultCode = 53
)
