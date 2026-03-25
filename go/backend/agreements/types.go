package agreements

type Signature struct {
	ID                      string  `db:"id"`
	AgreementID             string  `db:"agreement_id"`
	UserID                  string  `db:"user_id"`
	IPAddress               string  `db:"ip_address"`
	CreatedAt               string  `db:"created_at"`
	UpdatedAt               string  `db:"updated_at"`
	LastNotifiedAgreementID *string `db:"last_notified_agreement_id"`
}

type Agreement struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	Version     string `db:"version"`
	Content     string `db:"content"`
	CreatedAt   string `db:"created_at"`
	UpdatedAt   string `db:"updated_at"`
	GitFilePath string `db:"git_file_path"`
	Notified    bool   `db:"notified"`
}

type SignArgs struct {
	AgreementIDs []string `validate:"required"`
	UserID       string   `validate:"required"`
	IPAddress    string   `validate:"required,ip_addr"`
}

// AgreementChange identifies a changed agreement and the new version ID to exclude from old-signer queries.
type AgreementChange struct {
	Name     string
	ExceptID string
}
