package ops

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/env"
)

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
	if env.IsLocal() {
		_, privateKey, err := ed25519.GenerateKey(nil)
		if err != nil {
			return fmt.Errorf("%w %s", keys.ErrInternal, err)
		}

		var id string
		err = b.DB().GetContext(ctx, &id,
			"INSERT INTO wallet_keys (wallet_id,key_type,location, reference, name) values ($1, $2, $3, $4, $5) returning id",
			walletID, keys.Custodial.String(), "database", base64.StdEncoding.EncodeToString(privateKey.Seed()), "Fynbos Managed")
		if err != nil {
			return fmt.Errorf("%w %s", keys.ErrInternal, err)
		}
	}

	return nil
}

func ListKeys(ctx context.Context, b Backends, walletID string) ([]keys.Key, error) {
	sql := "SELECT id, wallet_id, key_type, location, reference, name FROM wallet_keys where wallet_id = $1 AND deleted_at IS NULL;"

	var ks []keys.Key
	err := b.DB().SelectContext(ctx, &ks, sql, walletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", keys.ErrInternal, err)
	}

	return ks, nil
}

func ListPublicKeys(ctx context.Context, b Backends, walletID string) ([]keys.Key, error) {
	sql := "SELECT id, wallet_id, key_type, location, reference, name FROM wallet_keys where wallet_id = $1 AND deleted_at IS NULL AND key_type=$2;"

	var ks []keys.Key
	err := b.DB().SelectContext(ctx, &ks, sql, walletID, keys.NonCustodial)
	if err != nil {
		return nil, fmt.Errorf("%w %s", keys.ErrInternal, err)
	}

	return ks, nil
}

func AddPublicKey(ctx context.Context, b Backends, walletID string, publicKeyBase64 string, name string) (*keys.Key, error) {
	var id string
	err := b.DB().GetContext(ctx, &id,
		"select id from wallet_keys where wallet_id = $1 and key_type = $2 and reference = $3", walletID, keys.NonCustodial.String(), publicKeyBase64)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w %s", keys.ErrInternal, err)
	}
	if id != "" {
		return nil, keys.ErrDuplicate
	}

	createdAt := time.Now()
	err = b.DB().GetContext(ctx, &id,
		"INSERT INTO wallet_keys (wallet_id,key_type,location, reference, name, created_at) values ($1, $2, $3, $4, $5, $6) returning id",
		walletID, keys.NonCustodial.String(), "database", publicKeyBase64, name, createdAt)
	if err != nil {
		return nil, fmt.Errorf("%w %s", keys.ErrInternal, err)
	}

	return &keys.Key{
		ID:        id,
		Name:      name,
		WalletID:  walletID,
		Type:      keys.NonCustodial,
		Location:  "database",
		Reference: publicKeyBase64,
		CreatedAt: createdAt,
	}, nil
}

func DeletePublicKey(ctx context.Context, b Backends, id string) error {
	_, err := b.DB().ExecContext(ctx, "UPDATE wallet_keys SET deleted_at=now():::TIMESTAMP WHERE id=$1 AND key_type=$2;", id, keys.NonCustodial)
	if err != nil {
		return fmt.Errorf("%w %s", keys.ErrInternal, err)
	}

	return nil
}

func getKey(ctx context.Context, b Backends, keyID string, walletID string) (*keys.Key, error) {
	sqlQuery := "SELECT id, wallet_id, key_type, location, reference, name FROM wallet_keys where id = $1 and wallet_id = $2"

	var k keys.Key
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

	refBytes, err := base64.StdEncoding.DecodeString(k.Reference)
	if err != nil {
		return nil, err
	}

	pk := ed25519.NewKeyFromSeed(refBytes)
	return ed25519.Sign(pk, message), nil
}

func Verify(ctx context.Context, b Backends, keyID string, walletID string, message, sig []byte) (bool, error) {
	k, err := getKey(ctx, b, keyID, walletID)
	if err != nil {
		return false, err
	}

	refBytes, err := base64.StdEncoding.DecodeString(k.Reference)
	if err != nil {
		return false, err
	}
	var pubKey ed25519.PublicKey
	if k.Type == keys.NonCustodial {
		pubKey = refBytes
	} else {
		pk := ed25519.NewKeyFromSeed(refBytes)
		pubKey = pk.Public().(ed25519.PublicKey)
	}

	return ed25519.Verify(pubKey, message, sig), nil
}
