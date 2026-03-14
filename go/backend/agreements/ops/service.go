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
			ID:          sign.ID,
			AgreementID: sign.AgreementID,
			UserID:      sign.UserID,
			IPAddress:   sign.IPAddress,
			CreatedAt:   sign.CreatedAt,
			UpdatedAt:   sign.UpdatedAt,
		})
	}

	return signatures, nil
}

// buildChangePlaceholders builds SQL WHERE clause conditions and collects args for a set of AgreementChange values.
// baseOffset is the 1-based index of the first placeholder ($baseOffset).
func buildChangePlaceholders(changes []agreements.AgreementChange, baseOffset int) (placeholders []string, args []interface{}) {
	for i, c := range changes {
		args = append(args, c.Name, c.ExceptID)
		base := baseOffset + 2*i
		placeholders = append(placeholders, fmt.Sprintf("(a.name = $%d AND a.id != $%d)", base, base+1))
	}
	return
}

// ListAffectedUserIDsPaginated returns a page of distinct user IDs who have signed at least one "old" version
// of any of the given agreements (old = same name but id != exceptID). Used to notify users when agreements change.
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

// GetAgreementNamesSignedByUsersFromSet returns for each user ID (in the given slice) the list of agreement
// names from the changes set that the user has signed an "old" version of. Used to personalize the agreement-change email.
func GetAgreementNamesSignedByUsersFromSet(ctx context.Context, b Backends, userIDs []string, changes []agreements.AgreementChange) (map[string][]string, error) {
	if len(userIDs) == 0 || len(changes) == 0 {
		return map[string][]string{}, nil
	}
	userArgs := make([]interface{}, len(userIDs))
	userPlaceholders := make([]string, len(userIDs))
	for i, u := range userIDs {
		userArgs[i] = u
		userPlaceholders[i] = fmt.Sprintf("$%d", i+1)
	}
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
