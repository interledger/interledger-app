package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"gitlab.com/fynbos/backend/agreements"

	"github.com/lib/pq"
)

func Get(ctx context.Context, b Backends, id string) (*agreements.Agreement, error) {
	var agreement agreements.Agreement

	err := b.DB().GetContext(ctx, &agreement, "SELECT * FROM agreements WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w %s", agreements.ErrNotFound, err.Error())
		}
		return nil, fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
	}

	return &agreement, nil
}

func Sign(ctx context.Context, b Backends, args *agreements.SignArgs) error {
	if err := b.Validator().Struct(args); err != nil {
		return fmt.Errorf("%w %s", agreements.ErrInvalidArgument, err)
	}

	tx, err := b.DB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Rollback()
	}()
	if err != nil {
		return fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
	}

	txStmt, err := tx.PrepareContext(ctx, "INSERT INTO agreement_signatures (agreement_id, user_id, ip_address) VALUES ($1, $2, $3)")
	if err != nil {
		return fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
	}
	defer txStmt.Close()

	for _, id := range args.AgreementIDs {
		_, err := txStmt.ExecContext(ctx, id, args.UserID, args.IPAddress)
		if err != nil {
			if pgErr, isPGErr := err.(*pq.Error); isPGErr {
				if pgErr.Code != "23503" {
					return fmt.Errorf("%w %s", agreements.ErrNotFound, err.Error())
				}
			}
			return fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
	}

	return nil
}

func GetSignatures(ctx context.Context, b Backends, userID string) ([]agreements.Signature, error) {
	var agreementSigns []agreements.Signature

	err := b.DB().SelectContext(ctx, &agreementSigns, "SELECT * FROM agreement_signatures WHERE user_id = $1", userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, agreements.ErrNotFound
		}
		return nil, fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
	}

	var signatures []agreements.Signature
	for _, sign := range agreementSigns {
		signatures = append(signatures, agreements.Signature{
			ID:                      sign.ID,
			AgreementID:             sign.AgreementID,
			UserID:                  sign.UserID,
			IPAddress:               sign.IPAddress,
			CreatedAt:               sign.CreatedAt,
			UpdatedAt:               sign.UpdatedAt,
			LastNotifiedAgreementID: sign.LastNotifiedAgreementID,
		})
	}

	return signatures, nil
}

func buildUserPlaceholders(userIDs []string) (placeholders []string, args []interface{}) {
	placeholders = make([]string, len(userIDs))
	args = make([]interface{}, len(userIDs))
	for i, u := range userIDs {
		args[i] = u
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	return
}

// buildChangePlaceholders returns WHERE conditions and args for changes, with placeholders starting at $baseOffset.
func buildChangePlaceholders(changes []agreements.AgreementChange, baseOffset int) (placeholders []string, args []interface{}) {
	for i, c := range changes {
		args = append(args, c.Name, c.ExceptID)
		base := baseOffset + 2*i
		placeholders = append(placeholders, fmt.Sprintf("(a.name = $%d AND a.id != $%d)", base, base+1))
	}
	return
}

// ListAffectedUserIDsPaginated returns user IDs who signed an older version of any of the given agreements.
func ListAffectedUserIDsPaginated(ctx context.Context, b Backends, changes []agreements.AgreementChange, limit, offset int) ([]string, error) {
	if len(changes) == 0 {
		return nil, nil
	}

	placeholders, args := buildChangePlaceholders(changes, 1)
	limitParam := len(args) + 1
	offsetParam := len(args) + 2
	args = append(args, limit, offset)

	query := fmt.Sprintf(`SELECT DISTINCT as_.user_id FROM agreement_signatures as_
		JOIN agreements a ON a.id = as_.agreement_id
		WHERE %s
		ORDER BY as_.user_id
		LIMIT $%d OFFSET $%d`, strings.Join(placeholders, " OR "), limitParam, offsetParam)

	var userIDs []string
	if err := b.DB().SelectContext(ctx, &userIDs, query, args...); err != nil {
		return nil, fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
	}

	return userIDs, nil
}

func MarkUsersNotified(ctx context.Context, b Backends, userIDs []string, changes []agreements.AgreementChange) error {
	if len(userIDs) == 0 || len(changes) == 0 {
		return nil
	}

	userPlaceholders, userArgs := buildUserPlaceholders(userIDs)
	nameParam := len(userIDs) + 1
	exceptIDParam := len(userIDs) + 2

	query := fmt.Sprintf(`UPDATE agreement_signatures SET last_notified_agreement_id = $%d
		WHERE user_id IN (%s)
		AND agreement_id IN (SELECT id FROM agreements WHERE name = $%d AND id != $%d)`,
		exceptIDParam,
		strings.Join(userPlaceholders, ","),
		nameParam,
		exceptIDParam,
	)

	for _, c := range changes {
		args := make([]interface{}, len(userIDs), len(userIDs)+2)
		copy(args, userArgs)
		args = append(args, c.Name, c.ExceptID)

		if _, err := b.DB().ExecContext(ctx, query, args...); err != nil {
			return fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
		}
	}
	return nil
}

// GetAgreementNamesSignedByUsersFromSet returns, keyed by user ID, the old agreement names each user has signed.
func GetAgreementNamesSignedByUsersFromSet(ctx context.Context, b Backends, userIDs []string, changes []agreements.AgreementChange) (map[string][]string, error) {
	if len(userIDs) == 0 || len(changes) == 0 {
		return map[string][]string{}, nil
	}

	userPlaceholders, userArgs := buildUserPlaceholders(userIDs)
	changePlaceholders, changeArgs := buildChangePlaceholders(changes, len(userIDs)+1)
	args := append(userArgs, changeArgs...)

	query := `SELECT DISTINCT ON (as_.user_id, a.name) as_.user_id, a.name FROM agreement_signatures as_
		JOIN agreements a ON a.id = as_.agreement_id
		WHERE as_.user_id IN (` + strings.Join(userPlaceholders, ",") + `) AND (` + strings.Join(changePlaceholders, " OR ") + `)
		ORDER BY as_.user_id, a.name`

	rows, err := b.DB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
	}
	defer rows.Close()

	result := make(map[string][]string)
	for rows.Next() {
		var userID, name string
		if err := rows.Scan(&userID, &name); err != nil {
			return nil, fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
		}
		result[userID] = append(result[userID], name)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
	}
	return result, nil
}
