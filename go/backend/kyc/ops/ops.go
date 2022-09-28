package ops

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"gitlab.com/fynbos/backend/kyc"
)

func mergeIdentities(old dbUserDetails, new kyc.UserDetails) (dbUserDetails, bool, error) {
	var merged dbUserDetails
	noop := true
	merged.UserID = new.UserID

	merged.FirstName = old.FirstName
	if new.FirstName != "" && new.FirstName != old.FirstName.String {
		noop = false
		merged.FirstName = sql.NullString{
			String: new.FirstName,
			Valid:  true,
		}
	}

	merged.LastName = old.LastName
	if new.LastName != "" && new.LastName != old.LastName.String {
		noop = false
		merged.LastName = sql.NullString{
			String: new.LastName,
			Valid:  true,
		}
	}

	merged.CountryCode = old.CountryCode
	if new.CountryCode != "" && new.CountryCode != old.CountryCode.String {
		noop = false
		merged.CountryCode = sql.NullString{
			String: new.CountryCode,
			Valid:  true,
		}
	}

	merged.Gender = old.Gender
	if new.Gender != kyc.GenderUnknown && new.Gender != kyc.Gender(old.Gender.Int32) {
		noop = false
		merged.Gender = sql.NullInt32{
			Int32: int32(new.Gender),
			Valid: true,
		}
	}

	merged.DateOfBirth = old.DateOfBirth
	if !new.DateOfBirth.IsZero() && !new.DateOfBirth.Equal(old.DateOfBirth.Time) {
		noop = false
		merged.DateOfBirth = sql.NullTime{
			Time:  new.DateOfBirth,
			Valid: true,
		}
	}

	merged.Address = old.Address
	if new.Address != nil {
		noop = false
		addressJson, err := json.Marshal(new.Address)
		if err != nil {
			return merged, noop, err
		}
		merged.Address = sql.NullString{
			String: string(addressJson),
			Valid:  true,
		}
	}

	merged.Revision = old.Revision + 1

	return merged, noop, nil
}

type dbUserDetails struct {
	UserID      string         `db:"user_id"`
	Revision    int            `db:"revision"`
	FirstName   sql.NullString `db:"first_name"`
	LastName    sql.NullString `db:"last_name"`
	CountryCode sql.NullString `db:"country_code"`
	Gender      sql.NullInt32  `db:"gender"`
	DateOfBirth sql.NullTime   `db:"date_of_birth"`
	Address     sql.NullString `db:"address"`
}

func getUserDetails(ctx context.Context, b Backends, userID string) (*dbUserDetails, error) {
	var id dbUserDetails
	err := b.DB().GetContext(ctx, &id,
		"SELECT user_id, revision, country_code, first_name, last_name, gender, date_of_birth, address FROM user_kyc_details WHERE user_id=$1 ORDER BY revision DESC LIMIT 1",
		userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, kyc.ErrNoKYCInfo
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", kyc.ErrInternal, err)
	}

	return &id, nil
}

func GetUserDetails(ctx context.Context, b Backends, userID string) (*kyc.UserDetails, error) {
	db, err := getUserDetails(ctx, b, userID)
	if err != nil {
		return nil, err
	}

	return convertDBDetails(*db)
}

func UpdateUserDetails(ctx context.Context, b Backends, ident kyc.UserDetails) (*kyc.UserDetails, error) {
	err := b.Validator().Struct(ident)
	if err != nil {
		return nil, fmt.Errorf("%w %s", kyc.ErrInternal, err)
	}

	old, err := getUserDetails(ctx, b, ident.UserID)
	if err != nil && !errors.Is(err, kyc.ErrNoKYCInfo) {
		return nil, err
	}

	// If this is the first insert we can compare with blank identity
	if old == nil {
		old = &dbUserDetails{}
	}
	merged, noop, err := mergeIdentities(*old, ident)
	if err != nil {
		return nil, fmt.Errorf("%w %s", kyc.ErrInternal, err)
	}

	if noop {
		//  Do nothing, just lookup the existing and off we go
		return convertDBDetails(merged)
	}

	_, err = b.DB().ExecContext(ctx, "INSERT INTO user_kyc_details (revision, user_id, country_code, first_name, last_name, gender, date_of_birth, address)"+
		" VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		merged.Revision, merged.UserID, merged.CountryCode, merged.FirstName, merged.LastName, merged.Gender, merged.DateOfBirth, merged.Address)
	if err != nil {
		return nil, fmt.Errorf("%w %s", kyc.ErrInternal, err)
	}

	return convertDBDetails(merged)
}

func convertDBDetails(details dbUserDetails) (*kyc.UserDetails, error) {
	resp := &kyc.UserDetails{
		UserID:      details.UserID,
		FirstName:   details.FirstName.String,
		LastName:    details.LastName.String,
		CountryCode: details.CountryCode.String,
		Gender:      kyc.Gender(details.Gender.Int32),
		DateOfBirth: details.DateOfBirth.Time,
	}

	if !details.Address.Valid {
		return resp, nil
	}

	var address kyc.Address
	err := json.Unmarshal([]byte(details.Address.String), &address)
	if err != nil {
		return nil, fmt.Errorf("%w %s", kyc.ErrInternal, err)
	}

	resp.Address = &address

	return resp, nil
}
