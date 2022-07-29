package country

type Country struct {
	ID          string
	Name        string
	Alpha_2     string
	Alpha_3     string
	NumericCode uint16 `db:"numeric_code"`
	CreatedAt   string `db:"created_at"`
	UpdatedAt   string `db:"updated_at"`
}
