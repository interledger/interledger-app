package migrations

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/agreements"
	"gitlab.com/fynbos/env"
)

//go:embed assets/*
var fs embed.FS

// gitFilePathForAsset returns the repo-relative path for an agreement file, same as MigrateFromEmbeddedMarkdowns.
// envName is the last path segment of the assets dir (e.g. "prod", "testing").
func gitFilePathForAsset(envName, filename string) string {
	return fmt.Sprintf("go/backend/agreements/migrations/assets/%s/%s", envName, filename)
}

func MigrateFromMarkdowns(ctx context.Context, db *sqlx.DB, dir string) error {
	agreementFiles, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("%w %s", agreements.ErrNotFound, err.Error())
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Rollback()
	}()
	if err != nil {
		return fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
	}

	regex, err := regexp.Compile(`^[a-zA-Z0-9_]+-[0-9]+\.[0-9]+\.[0-9]+\.md$`)
	if err != nil {
		return fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
	}

	txStmt, err := tx.PrepareContext(ctx, `INSERT INTO agreements (id, name, version, content, git_file_path) VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`)
	if err != nil {
		return fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
	}
	defer txStmt.Close()

	var agreementIDs []string
	err = tx.SelectContext(ctx, &agreementIDs, "SELECT id FROM agreements")
	if err != nil {
		return fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
	}

	// convert to map for faster lookup
	agreementsMap := make(map[string]bool)
	for _, agreement := range agreementIDs {
		agreementsMap[agreement] = true
	}

	envName := filepath.Base(dir)

	for _, agreementFile := range agreementFiles {
		if !regex.MatchString(agreementFile.Name()) {
			return fmt.Errorf("%w %s", agreements.ErrInternal, "invalid agreement file name format")
		}

		agreementID := agreementFile.Name()[:len(agreementFile.Name())-3]
		if _, ok := agreementsMap[agreementID]; ok {
			continue
		}

		agreementName := agreementID[:strings.Index(agreementID, "-")]
		agreementVersion := agreementID[strings.Index(agreementID, "-")+1:]

		agreementContent, err := os.ReadFile(fmt.Sprintf("%s/%s", dir, agreementFile.Name()))
		if err != nil {
			return fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
		}

		gitFilePath := gitFilePathForAsset(envName, agreementFile.Name())

		_, err = txStmt.Exec(agreementID, agreementName, agreementVersion, string(agreementContent), gitFilePath)
		if err != nil {
			return fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
	}

	return nil
}

func MigrateFromEmbeddedMarkdowns(ctx context.Context, db *sqlx.DB) error {
	envName := env.GetEnv()
	dir := fmt.Sprintf("assets/%s", envName)
	agreementFiles, err := fs.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("%w %s", agreements.ErrNotFound, err)
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Rollback()
	}()
	if err != nil {
		return fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
	}

	regex, err := regexp.Compile(`^[a-zA-Z0-9_]+-[0-9]+\.[0-9]+\.[0-9]+\.md$`)
	if err != nil {
		return fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
	}

	txStmt, err := tx.PrepareContext(ctx, `INSERT INTO agreements (id, name, version, content, git_file_path) VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`)
	if err != nil {
		return fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
	}
	defer txStmt.Close()

	var agreementIDs []string
	err = tx.SelectContext(ctx, &agreementIDs, "SELECT id FROM agreements")
	if err != nil {
		return fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
	}

	// convert to map for faster lookup
	agreementsMap := make(map[string]bool)
	for _, agreement := range agreementIDs {
		agreementsMap[agreement] = true
	}

	for _, agreementFile := range agreementFiles {
		if !regex.MatchString(agreementFile.Name()) {
			return fmt.Errorf("%w %s", agreements.ErrInternal, "invalid agreement file name format")
		}

		agreementID := agreementFile.Name()[:len(agreementFile.Name())-3]
		if _, ok := agreementsMap[agreementID]; ok {
			continue
		}

		agreementName := agreementID[:strings.Index(agreementID, "-")]
		agreementVersion := agreementID[strings.Index(agreementID, "-")+1:]

		agreementContent, err := fs.ReadFile(fmt.Sprintf("%s/%s", dir, agreementFile.Name()))
		if err != nil {
			return fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
		}

		gitFilePath := gitFilePathForAsset(envName, agreementFile.Name())

		_, err = txStmt.Exec(agreementID, agreementName, agreementVersion, string(agreementContent), gitFilePath)
		if err != nil {
			return fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
	}

	return nil
}
