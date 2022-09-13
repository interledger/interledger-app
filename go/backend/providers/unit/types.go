package unit

import "github.com/jmoiron/sqlx/types"

type (
	StatementPDF struct {
		ID  string
		PDF []byte
	}

	GetStatementPDFArgs struct {
		StatementID string `validate:"required"`
		CustomerID  string `validate:"required"`
	}

	Statement struct {
		ID        string
		Period    string
		AccountID string
	}

	ApplicationForm struct {
		ID  string
		URL string
	}

	CreateApplicationFormArgs struct {
		ID      string `validate:"required"`
		Email   string `validate:"required"`
		Country string `validate:"required"`
	}

	CreateApplicationArgs struct {
		Ssn                string   `validate:"required"`
		DateOfBirth        string   `validate:"required"`
		Street             string   `validate:"required"`
		Street2            string   `validate:"required"`
		City               string   `validate:"required"`
		State              string   `validate:"required"`
		PostalCode         string   `validate:"required"`
		IpAddress          string   `validate:"required"`
		UserID             string   `validate:"required"`
		DeviceFingerprints []string `validate:"required"`
	}

	Application struct {
		Type         string
		ID           string
		Status       string
		FynbosUserId string
		Archived     bool
		CustomerID   string
	}

	// Maps the unit customer to the Fynbos account. ID is the external ID.
	Customer struct {
		ID         string `db:"id"`
		IdentityID string `db:"identity_id"`
		Type       string `db:"type"`
		CreatedAt  string `db:"created_at"`
		UpdatedAt  string `db:"updated_at"`
	}

	CreateCustomerArgs struct {
		ID         string
		IdentityID string
		Type       string
	}

	CreateCounterPartyArgs struct {
		FundingsourceID string `validate:"uuid4"`
		Name            string `validate:"required"`
		UnitCustomerID  string `validate:"required"`
		RoutingNumber   string `validate:"required"`
		AccountNumber   string `validate:"required"`
		AccountType     string `validate:"required"`
		Type            string `validate:"required"`

		// This idempotency key is valid for 48 hours on Unit's api.
		IdempotencyKey string `validate:"lte=255"`
	}

	// Maps the unit counterparty to the Fynbos fundingsource. ID is the external ID.
	CounterParty struct {
		ID              string
		FundingsourceID string `db:"fundingsource_id"`
		CreatedAt       string `db:"created_at"`
		UpdatedAt       string `db:"updated_at"`
	}

	// Maps the unit deposit account and unit customerID. ID is the external ID.
	DepositAccount struct {
		ID         string `db:"id"`
		CustomerID string `db:"customer_id"`
		CreatedAt  string `db:"created_at"`
		UpdatedAt  string `db:"updated_at"`
	}

	UserAchDeposit struct {
		ID               string `db:"id"`
		DepositAccountID string `db:"deposit_account_id"`
		DepositID        string `db:"deposit_id"`
		CounterPartyID   string `db:"counterparty_id"`
		Amount           uint64
		CreatedAt        string `db:"created_at"`
		UpdatedAt        string `db:"updated_at"`
	}

	InitiateUserDepositArgs struct {
		DepositID       string `validate:"uuid4"`
		AccountID       string `validate:"uuid4"`
		FundingsourceID string `validate:"uuid4"`
		Amount          uint64

		// This will show up on the statement of the counterparty account.
		Description string
	}
)

type DbEvent struct {
	ID        string         `db:"id"`
	Type      string         `db:"type"`
	RawEvent  types.JSONText `db:"raw_event"`
	CreatedAt string         `db:"created_at"`
	UpdatedAt string         `db:"updated_at"`
}

const (
	OnboardingCustomerCreatedChannel   = "onboard-unit-customer-created"
	OnboardingApplicationDeniedChannel = "onboard-unit-application-denied"
	OnboardingWorkflowName             = "unit_onboarding_"

	AchDepositChannel   = "unit-user-ach-deposit"
	DepositWorkflowName = "deposit_"
)
