package external

type (
	User struct {
		Guid       string `json:"guid"`
		ID         string `json:"id"`
		IsDisabled bool   `json:"is_disabled"`
	}

	GetWidgetURLArgs struct {
		UserGuid                 string `json:"-"`
		CurrentMemberGuid        string `json:"current_member_guid,omitempty"`
		DisableInstitutionSearch bool   `json:"disable_institution_search,omitempty"`
		IncludeTransactions      bool   `json:"include_transactions,omitempty"`
		IncludeIdentity          bool   `json:"include_identity,omitempty"`
		Mode                     string `json:"mode"`
		WidgetType               string `json:"widget_type"`
	}

	WidgetURL struct {
		WidgetType string `json:"type"`
		URL        string `json:"url"`
		UserID     string `json:"user_id"`
	}

	GetWidgerURLResponse struct {
		WidgetURL WidgetURL `json:"widget_url"`
	}

	Pagination struct {
		CurrentPage  int `json:"current_page"`
		PerPage      int `json:"per_page"`
		TotalEntries int `json:"total_entries"`
		TotalPages   int `json:"total_pages"`
	}

	ListUsersResponse struct {
		Users      []User     `json:"users"`
		Pagination Pagination `json:"pagination"`
	}

	AccountOwners struct {
		Guid        string
		UserGuid    string `json:"user_guid"`
		MemberGuid  string `json:"member_guid"`
		AccountGuid string `json:"account_guid"`
		Address     string
		City        string
		Country     string
		Email       string
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		OwnerName   string `json:"owner_name"`
		Phone       string `json:"phone"`
		PostalCode  string `json:"postal_code"`
		State       string `json:"state"`
	}

	ListAccountOwnersResponse struct {
		AccountOwners []AccountOwners `json:"account_owners"`
		Pagination    Pagination      `json:"pagination"`
	}

	AccountNumbers struct {
		Guid              string
		MemberGuid        string `json:"member_guid"`
		UserGuid          string `json:"user_guid"`
		AccountGuid       string `json:"account_guid"`
		AccountNumber     string `json:"account_number"`
		InstitutionNumber string `json:"institution_number"`
		RoutingNumber     string `json:"routing_number"`
		TransitNumber     string `json:"transit_number"`
		PassedValidation  bool   `json:"passed_validation"`
	}

	ListAccountNumbersResponse struct {
		AccountNumbers []AccountNumbers `json:"account_numbers"`
		Pagination     Pagination       `json:"pagination"`
	}

	Account struct {
		AccountNumber         string `json:"account_number,omitempty"`
		Apr                   string `json:"apr,omitempty"`
		Apy                   string `json:"apy,omitempty"`
		AvailableBalance      string `json:"available_balance,omitempty"`
		AvailableCredit       string `json:"available_credit,omitempty"`
		Balance               string `json:"balance,omitempty"`
		CashBalance           string `json:"cash_balance,omitempty"`
		CashSurrenderValue    string `json:"cash_surrender_value,omitempty"`
		CreatedAt             string `json:"created_at,omitempty"`
		CreditLimit           string `json:"credit_limit,omitempty"`
		CurrencyCode          string `json:"currency_code,omitempty"`
		DayPaymentIsDue       string `json:"day_payment_is_due,omitempty"`
		DeathBenefit          string `json:"death_benefit,omitempty"`
		GUID                  string `json:"guid,omitempty"`
		HoldingsValue         string `json:"holdings_value,omitempty"`
		ID                    string `json:"id,omitempty"`
		ImportedAt            string `json:"imported_at,omitempty"`
		InstitutionCode       string `json:"institution_code,omitempty"`
		InsuredName           string `json:"insured_name,omitempty"`
		InterestRate          string `json:"interest_rate,omitempty"`
		IsClosed              string `json:"is_closed,omitempty"`
		IsHidden              string `json:"is_hidden,omitempty"`
		LastPayment           string `json:"last_payment,omitempty"`
		LastPaymentAt         string `json:"last_payment_at,omitempty"`
		LoanAmount            string `json:"loan_amount,omitempty"`
		MaturesOn             string `json:"matures_on,omitempty"`
		MemberGUID            string `json:"member_guid,omitempty"`
		MemberID              string `json:"member_id,omitempty"`
		MemberIsManagedByUser string `json:"member_is_managed_by_user,omitempty"`
		Metadata              string `json:"metadata,omitempty"`
		MinimumBalance        string `json:"minimum_balance,omitempty"`
		MinimumPayment        string `json:"minimum_payment,omitempty"`
		Name                  string `json:"name,omitempty"`
		Nickname              string `json:"nickname,omitempty"`
		OriginalBalance       string `json:"original_balance,omitempty"`
		PayOutAmount          string `json:"pay_out_amount,omitempty"`
		PaymentDueAt          string `json:"payment_due_at,omitempty"`
		PayoffBalance         string `json:"payoff_balance,omitempty"`
		PremiumAmount         string `json:"premium_amount,omitempty"`
		RoutingNumber         string `json:"routing_number,omitempty"`
		StartedOn             string `json:"started_on,omitempty"`
		Subtype               string `json:"subtype,omitempty"`
		TotalAccountValue     string `json:"total_account_value,omitempty"`
		Type                  string `json:"type,omitempty"`
		UpdatedAt             string `json:"updated_at,omitempty"`
		UserGUID              string `json:"user_guid,omitempty"`
		UserID                string `json:"user_id,omitempty"`
	}

	ListAccountsResponse struct {
		Accounts   []Account
		Pagination Pagination `json:"pagination"`
	}
)
