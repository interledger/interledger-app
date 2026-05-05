package agreements

type Signature struct {
	ID          string `db:"id"`
	AgreementID string `db:"agreement_id"`
	UserID      string `db:"user_id"`
	CreatedAt   string `db:"created_at"`
	UpdatedAt   string `db:"updated_at"`
}

type Agreement struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	Version     string `db:"version"`
	Content     string `db:"content"`
	CreatedAt   string `db:"created_at"`
	UpdatedAt   string `db:"updated_at"`
	GitFilePath string `db:"git_file_path"`
}

type SignArgs struct {
	AgreementIDs []string `validate:"required"`
	UserID       string   `validate:"required"`
}
