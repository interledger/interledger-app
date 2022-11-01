package ops

type dbWallet struct {
	ID         string `db:"id"`
	SendUserID string `db:"send_user_id"`
	Nickname   string `db:"nickname"`
	CreatedAt  string `db:"created_at"`
	UpdatedAt  string `db:"updated_at"`
}
