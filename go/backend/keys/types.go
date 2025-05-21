package keys

import (
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"database/sql/driver"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"time"
)

type Key struct {
	ID        string
	Name      string
	KeyID     string `db:"key_id"`
	WalletID  string `db:"wallet_id"`
	Type      Type   `db:"key_type"`
	Location  string
	Reference string
	PublicKey string    `db:"public_key"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	DeletedAt time.Time `db:"deleted_at"`
}

func (key Key) Fingerprint() (string, error) {
	if key.PublicKey == "" {
		return "", fmt.Errorf("%w Key is empty.", ErrInternal)
	}
	keyBytes, err := base64.RawURLEncoding.DecodeString(key.PublicKey)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	switch key.Type {
	case NonCustodial:
		pkixBytes, err := x509.MarshalPKIXPublicKey(ed25519.PublicKey(keyBytes))
		if err != nil {
			return "", fmt.Errorf("%w %s", ErrInternal, err)
		}

		hashedPem := sha256.Sum256(pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: pkixBytes,
		}))

		fingerprint := fmt.Sprintf("SHA256:%s", hex.EncodeToString(hashedPem[:]))
		return fingerprint, nil
	case Custodial:
		pkixBytes, err := x509.MarshalPKCS8PrivateKey(ed25519.PrivateKey(keyBytes))
		if err != nil {
			return "", fmt.Errorf("%w %s", ErrInternal, err)
		}

		hashedPem := sha256.Sum256(pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: pkixBytes,
		}))

		fingerprint := fmt.Sprintf("SHA256:%s", hex.EncodeToString(hashedPem[:]))
		return fingerprint, nil
	default:
		return "", fmt.Errorf("%w Unknown key type=%s", ErrInternal, key.Type)
	}
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
