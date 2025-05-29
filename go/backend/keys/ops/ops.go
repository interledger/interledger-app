package ops

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"gitlab.com/fynbos/log"
	"go.uber.org/zap"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/backend/vault"
	"gitlab.com/fynbos/env"
)

type keyDB struct {
	keys.Key
	Reference sql.NullString
	PublicKey sql.NullString `db:"public_key"`
}

func GeneratePrivateKey(ctx context.Context, b Backends, walletID string) error {
	// Check if it exists yet?
	var id string
	err := b.DB().GetContext(ctx, &id,
		"select id from wallet_keys where wallet_id = $1 and key_type = $2", walletID, keys.Custodial.String())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w %s", keys.ErrInternal, err)
	}
	if id != "" {
		return nil
	}

	// Local env generate and store in DB
	if env.IsLocal() || env.IsTest() {
		publicKey, privateKey, err := ed25519.GenerateKey(nil)
		if err != nil {
			return fmt.Errorf("%w %s", keys.ErrInternal, err)
		}

		publicKeyBase64 := base64.StdEncoding.EncodeToString(publicKey)
		var id string
		err = b.DB().GetContext(ctx, &id,
			"INSERT INTO wallet_keys (wallet_id,key_type,location, reference, name, public_key, key_id) values ($1, $2, $3, $4, $5, $6, $7) returning id",
			walletID, keys.Custodial.String(), "database", base64.StdEncoding.EncodeToString(privateKey.Seed()), "Interledger Managed", publicKeyBase64, uuid.NewString())
		if err != nil {
			return fmt.Errorf("%w %s", keys.ErrInternal, err)
		}
		return nil
	}

	// Create key in vault.
	keyID := uuid.NewString()
	err = b.Vault().CreateKey(keyID)
	if err != nil {
		log.Error("unable to create key: %v", zap.Error(err))
		return err
	}

	publicKey, err := b.Vault().GetPublicKey(keyID)
	if err != nil {
		log.Error("unable to read key: %v", zap.Error(err))
		return err
	}

	err = b.DB().GetContext(ctx, &id,
		"INSERT INTO wallet_keys (wallet_id,key_type,location, reference, name, public_key, key_id) values ($1, $2, $3, $4, $5, $6, $7) returning id",
		walletID, keys.Custodial.String(), "vault", keyID, "Interledger Managed", publicKey, uuid.NewString())
	if err != nil {
		return fmt.Errorf("%w %s", keys.ErrInternal, err)
	}

	return nil
}

func ListKeys(ctx context.Context, b Backends, walletID string) ([]keys.Key, error) {
	sql := "SELECT id, wallet_id, key_type, location, reference, name, public_key, key_id, created_at FROM wallet_keys where wallet_id = $1 AND deleted_at IS NULL;"

	var ks []keyDB
	err := b.DB().SelectContext(ctx, &ks, sql, walletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", keys.ErrInternal, err)
	}

	var publicKeys []keys.Key
	for _, k := range ks {
		publicKeys = append(publicKeys, convertToKeyPublic(k))
	}

	return publicKeys, nil
}

func GetPublicKey(ctx context.Context, b Backends, id string, walletID string) (*keys.Key, error) {
	keyDb, err := getKey(ctx, b, id, walletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", keys.ErrInternal, err)
	}

	key := convertToKeyPublic(*keyDb)

	return &key, nil
}

func AddPublicKey(ctx context.Context, b Backends, walletID, publicKeyBase64, name, keyID string) (*keys.Key, error) {
	var id string
	err := b.DB().GetContext(ctx, &id,
		"select id from wallet_keys where wallet_id = $1 and key_type = $2 and public_key = $3 and deleted_at IS NULL", walletID, keys.NonCustodial.String(), publicKeyBase64)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w %s", keys.ErrInternal, err)
	}
	if id != "" {
		return nil, keys.ErrDuplicate
	}

	createdAt := time.Now()
	err = b.DB().GetContext(ctx, &id,
		"INSERT INTO wallet_keys (wallet_id,key_type,location, public_key, name, created_at, key_id) values ($1, $2, $3, $4, $5, $6, $7) returning id",
		walletID, keys.NonCustodial.String(), "database", publicKeyBase64, name, createdAt, keyID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", keys.ErrInternal, err)
	}

	return &keys.Key{
		ID:        id,
		Name:      name,
		KeyID:     keyID,
		WalletID:  walletID,
		Type:      keys.NonCustodial,
		Location:  "database",
		Reference: "",
		PublicKey: publicKeyBase64,
		CreatedAt: createdAt,
	}, nil
}

func DeletePublicKey(ctx context.Context, b Backends, id string) error {
	_, err := b.DB().ExecContext(ctx, "UPDATE wallet_keys SET deleted_at=now()::TIMESTAMP WHERE id=$1 AND key_type=$2;", id, keys.NonCustodial)
	if err != nil {
		return fmt.Errorf("%w %s", keys.ErrInternal, err)
	}

	return nil
}

func getKey(ctx context.Context, b Backends, keyID string, walletID string) (*keyDB, error) {
	sqlQuery := "SELECT id, wallet_id, key_type, location, reference, name, key_id, public_key, key_id FROM wallet_keys where id = $1 and wallet_id = $2"

	var k keyDB
	err := b.DB().GetContext(ctx, &k, sqlQuery, keyID, walletID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w %s", keys.ErrNotFound, err)
		}
		return nil, fmt.Errorf("%w %s", keys.ErrInternal, err)
	}

	return &k, nil
}

func Sign(ctx context.Context, b Backends, keyID string, walletID string, message []byte) ([]byte, error) {
	k, err := getKey(ctx, b, keyID, walletID)
	if err != nil {
		return nil, err
	}

	if k.Type != keys.Custodial {
		return nil, fmt.Errorf("%w can only sign with custodial keys", keys.ErrInternal)
	}

	if env.IsLocal() || env.IsTest() {
		refBytes, err := base64.StdEncoding.DecodeString(k.Reference.String)
		if err != nil {
			return nil, err
		}

		pk := ed25519.NewKeyFromSeed(refBytes)
		return ed25519.Sign(pk, message), nil
	}

	signedMessage, err := b.Vault().Sign(k.Reference.String, string(message))
	if err != nil {
		return nil, err
	}

	return signedMessage, nil
}

func Verify(ctx context.Context, b Backends, keyID string, walletID string, message, sig []byte) (bool, error) {
	k, err := getKey(ctx, b, keyID, walletID)
	if err != nil {
		return false, err
	}

	if k.Type == keys.NonCustodial {
		refBytes, err := base64.StdEncoding.DecodeString(k.PublicKey.String)
		if err != nil {
			return false, err
		}

		return ed25519.Verify(refBytes, message, sig), nil
	}

	// If local we need to pull the private key out of reference
	if env.IsLocal() || env.IsTest() {
		refBytes, err := base64.StdEncoding.DecodeString(k.Reference.String)
		if err != nil {
			return false, err
		}

		pk := ed25519.NewKeyFromSeed(refBytes)
		pubKey := pk.Public().(ed25519.PublicKey)
		return ed25519.Verify(pubKey, message, sig), nil
	}

	// Otherwise we can verify with vault.
	return b.Vault().Verify(k.Reference.String, vault.VerifyInput{
		Input:     string(message),
		Signature: string(sig),
	})
}

func FixWalletPublicKeys(ctx context.Context, b Backends, walletID string) error {
	ks, err := ListKeys(ctx, b, walletID)
	if err != nil {
		return err
	}

	for _, k := range ks {
		if k.PublicKey != "" {
			continue
		}
		if k.Type == keys.NonCustodial {
			_, err = b.DB().ExecContext(ctx, "UPDATE wallet_keys SET public_key=reference, reference=NULL WHERE id=$1 AND key_type=$2;", k.ID, keys.NonCustodial)
			if err != nil {
				return fmt.Errorf("%w %s", keys.ErrInternal, err)
			}
		}
		if k.Type == keys.Custodial && k.Location == "vault" {
			vaultPublicKey, err := b.Vault().GetPublicKey(k.Reference)
			if err != nil {
				return fmt.Errorf("%w %s", keys.ErrInternal, err)
			}

			_, err = b.DB().ExecContext(ctx, "UPDATE wallet_keys SET public_key=$3 WHERE id=$1 AND key_type=$2;", k.ID, keys.Custodial, vaultPublicKey)
			if err != nil {
				return fmt.Errorf("%w %s", keys.ErrInternal, err)
			}
		}
	}

	return nil
}

func convertToKeyPublic(keyDB keyDB) keys.Key {
	reference := ""
	if keyDB.Reference.Valid {
		reference = keyDB.Reference.String
	}

	publicKey := ""
	if keyDB.PublicKey.Valid {
		publicKey = keyDB.PublicKey.String
	}

	return keys.Key{
		ID:        keyDB.ID,
		Name:      keyDB.Name,
		KeyID:     keyDB.KeyID,
		WalletID:  keyDB.WalletID,
		Type:      keyDB.Type,
		Location:  keyDB.Location,
		Reference: reference,
		PublicKey: publicKey,
		CreatedAt: keyDB.CreatedAt,
		UpdatedAt: keyDB.UpdatedAt,
	}
}

func RemoveCustodialKeysForWallet(ctx context.Context, b Backends, walletID string) error {
	ks, err := ListKeys(ctx, b, walletID)

	if err != nil {
		return err
	}

	tx, err := b.DB().BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		err = tx.Rollback()
	}()

	txStmtWK, err := tx.PrepareContext(ctx, "DELETE FROM wallet_keys WHERE id = $1 AND key_type = $2")
	if err != nil {
		return fmt.Errorf("%w %s", keys.ErrInternal, err.Error())
	}

	txStmtRWK, err := tx.PrepareContext(ctx, "DELETE FROM rafiki_wallet_keys WHERE internal_id = $1")
	if err != nil {
		return fmt.Errorf("%w %s", keys.ErrInternal, err.Error())
	}

	for _, k := range ks {
		if k.Type == keys.Custodial {
			_, err = txStmtWK.ExecContext(ctx, k.ID, keys.Custodial.String())
			if err != nil {
				return fmt.Errorf("%w %s", keys.ErrInternal, err)
			}

			_, err = txStmtRWK.ExecContext(ctx, k.ID)
			if err != nil {
				return fmt.Errorf("%w %s", keys.ErrInternal, err)
			}

			err = b.Rafiki().RevokePaymentPointerKey(ctx, k.ID)
			if err != nil {
				return fmt.Errorf("%w %s", keys.ErrInternal, err)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("%w %s", keys.ErrInternal, err)
	}

	return nil
}
