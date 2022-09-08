package mx

type (
	Account struct {
		Guid            string //from mx
		UserGuid        string `db:"user_guid"`   // from mx
		MemberGuid      string `db:"member_guid"` // from mx
		AccountID       string `db:"account_id"`  // Fynbos account id
		FundingsourceID string `db:"fundingsource_id"`
		CreatedAt       string `db:"created_at"`
		UpdatedAt       string `db:"updated_at"`
	}

	Member struct {
		Guid                     string
		UserGuid                 string
		AggregatedAt             string `json:"aggregated_at"`
		IsBeingAggregated        bool   `json:"is_being_aggregated"`
		SuccessfullyAggregatedAt string `json:"successfully_aggregated_at"`
		ConnectionStatus         string `json:"connection_status"`
	}

	AccountOwner struct {
		AccountGuid string
		OwnerName   string
	}

	AccountDetails struct {
		Guid            string
		UserGuid        string `json:"user_guid"`
		MemberGuid      string `json:"member_guid"`
		AccountNumber   string `json:"account_number"`
		InstitutionCode string `json:"institution_code"`
		RoutingNumber   string `json:"routing_number"`
		TransitNumber   string `json:"transit_number"`
		CurrencyCode    string `json:"currency_code"`
		Type            string
	}

	AccountBalance struct {
		AssetCode  string
		AssetScale uint8
		Value      int64
	}

	CreateAccountArgs struct {
		Guid            string `validate:"required"` // from mx
		UserGuid        string `validate:"required"` // from mx
		MemberGuid      string `validate:"required"` // from mx
		AccountID       string `validate:"uuid4"`
		FundingsourceID string `validate:"uuid4"`
	}

	GetAccountOwnerArgs struct {
		MxUserGuid    string
		MxMemberGuid  string
		MxAccountGuid string
	}

	VerifyOwnershipArgs struct {
		AccountID     string
		MxUserGuid    string
		MxMemberGuid  string
		MxAccountGuid string
	}

	InitiateCreateAccountArgs struct {
		UserGuid   string `validate:"required"` // from mx
		MemberGuid string `validate:"required"` // from mx
		AccountID  string `validate:"uuid4"`
		IdentityID string `validate:"uuid4"`
	}

	InitiateCreateFundingsourceArgs struct {
		AccountID     string `validate:"required"`
		Otp           string `validate:"required"`
		Name          string `validate:"required"`
		MxAccountGuid string `validate:"uuid4"`
	}
)
