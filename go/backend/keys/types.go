package keys

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type Key struct {
	ID        string
	Name      string
	WalletID  string `db:"wallet_id"`
	Type      Type   `db:"key_type"`
	Location  string
	Reference string
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	DeletedAt time.Time `db:"deleted_at"`
}

type Type string

const (
	Custodial    Type = "custodial"
	NonCustodial Type = "noncustodial"
)

func (kt Type) String() string {
	switch kt {
	case Custodial:
		return "custodial"
	case NonCustodial:
		return "noncustodial"
	default:
		return ""
	}
}

func IsValidKeyType(s string) bool {
	return s == Custodial.String() || s == NonCustodial.String()
}

// Scan converts the database value to a Type value
func (kt *Type) Scan(value interface{}) error {
	if value == nil {
		*kt = NonCustodial
		return nil
	}
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("unsupported KeyType value: %v", value)
	}
	switch s {
	case Custodial.String():
		*kt = Custodial
	case NonCustodial.String():
		*kt = NonCustodial
	default:
		return fmt.Errorf("unsupported KeyType value: %v", s)
	}
	return nil
}

// Value converts the Type value to a database value
func (kt Type) Value() (driver.Value, error) {
	return kt.String(), nil
}
