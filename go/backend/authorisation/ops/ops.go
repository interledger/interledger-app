package ops

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gitlab.com/fynbos/env"

	"github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/authorisation"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/httpmessagesignatures"
	"gitlab.com/fynbos/log"
)

func CreateClient(ctx context.Context, b Backends, clientURL string) (*authorisation.Client, error) {

	// Ensure the client URL is a fynbos payment pointer.
	_, err := b.Wallets().GetFromAddress(ctx, clientURL)
	if err != nil {
		return nil, err
	}

	var client authorisation.Client
	err = b.DB().GetContext(ctx, &client, "INSERT INTO authorisation_clients (url) VALUES ($1) RETURNING id, url;", clientURL)
	if db.IsErrorCode(err, db.UniqueViolationError) {
		return LookupClient(ctx, b, clientURL)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", authorisation.ErrInternal, err)
	}

	return &client, nil
}

func LookupClient(ctx context.Context, b Backends, clientURL string) (*authorisation.Client, error) {
	var client authorisation.Client
	err := b.DB().GetContext(ctx, &client, "SELECT id, url FROM authorisation_clients WHERE url=$1", clientURL)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w %s", authorisation.ErrNotFound, err)
	} else if err != nil {
		return nil, fmt.Errorf("%w %s", authorisation.ErrInternal, err)
	}

	return &client, nil
}

func CreateGrant(ctx context.Context, b Backends, args authorisation.GrantRequest) (*authorisation.Grant, error) {
	err := b.Validator().StructCtx(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", authorisation.ErrInvalidArgument, err)
	}

	cl, err := LookupClient(ctx, b, args.Client)
	if err != nil {
		return nil, err
	}

	gid := uuid.NewString()

	tokens, err := validateTokenAccess(ctx, b, args)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, fmt.Errorf("%w no valid access token requests found", authorisation.ErrInvalidArgument)
	}

	err = crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {

		_, err := tx.ExecContext(ctx, "INSERT INTO authorisation_grants (id, client_id, state, continue_token, wait) VALUES ($1, $2, $3, $4, $5)",
			gid, cl.ID, authorisation.GrantStateApproved, uuid.NewString(), 0)
		if err != nil {
			return fmt.Errorf("%w %s", authorisation.ErrInternal, err)
		}

		for _, tkn := range tokens {
			tokenID := uuid.NewString()
			label := tkn.Label
			if label == "" {
				label = fmt.Sprintf("auto_label_%d", rand.Int31n(10000))
			}
			_, err = tx.ExecContext(ctx, "INSERT INTO authorisation_tokens (id, grant_id, state, token, label, expires_in) VALUES ($1, $2, $3, $4, $5, $6)",
				tokenID, gid, authorisation.TokenStateEnabled, uuid.NewString(), label, 60*60)
			if err != nil {
				return fmt.Errorf("%w %s", authorisation.ErrInternal, err)
			}

			for _, acc := range tkn.Access {
				_, err = tx.ExecContext(ctx, "INSERT INTO authorisation_token_access (token_id, type, actions, identifier, locations, data_types) VALUES ($1, $2, $3, $4, $5, $6)",
					tokenID, acc.Type, pq.Array(acc.Actions), acc.Identifier, pq.Array(acc.Locations), pq.Array(acc.Datatypes))
				if err != nil {
					return fmt.Errorf("%w %s", authorisation.ErrInternal, err)
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return lookupGrant(ctx, b, gid)
}

// validateTokenAccess returns all the tokens for access that can automatically be granted.
// Currently, the only supported access is for "incoming-payments" type and "read,write" actions
func validateTokenAccess(ctx context.Context, b Backends, args authorisation.GrantRequest) ([]authorisation.AccessTokenReq, error) {

	// Check that the request is for one of the Fynbos payment pointers
	wa, err := b.Wallets().GetFromAddress(ctx, args.Client)
	if err != nil {
		return nil, err
	}

	var resp []authorisation.AccessTokenReq
	for _, at := range args.AccessToken {
		var access []authorisation.Access
		for _, acc := range at.Access {

			// Only allow access to your own payment pointer for now.
			if !strings.EqualFold(acc.Identifier, wa.AddressString()) {
				continue
			}

			if !strings.EqualFold(acc.Type, "incoming-payment") && !strings.EqualFold(acc.Type, "outgoing-payment") {
				continue
			}

			var actions []string
			for _, act := range acc.Actions {
				if strings.EqualFold(act, "read") || strings.EqualFold(act, "write") {
					actions = append(actions, act)
				}
			}

			// No valid actions where found, ignore the rest of this access
			if len(actions) == 0 {
				continue
			}

			location, err := url.JoinPath(env.OpenPaymentsURL(), "incoming")
			if err != nil {
				return nil, err
			}

			if strings.EqualFold(acc.Type, "outgoing-payment") {
				location, err = url.JoinPath(env.OpenPaymentsURL(), "outgoing")
			}
			if err != nil {
				return nil, err
			}

			access = append(access, authorisation.Access{
				Type:       acc.Type,
				Actions:    actions,
				Locations:  []string{location},
				Identifier: acc.Identifier,
			})
		}
		// No valid access requests where found for this token. Ignore it.
		if len(access) == 0 {
			continue
		}

		resp = append(resp, authorisation.AccessTokenReq{
			Access: access,
			Label:  at.Label,
		})
	}

	return resp, nil
}

type dbGrant struct {
	ID            string                   `db:"id"`
	State         authorisation.GrantState `db:"state"`
	ClientID      string                   `db:"client_id"`
	ContinueToken string                   `db:"continue_token"`
	Wait          string                   `db:"wait"`
}

type dbAccessToken struct {
	ID        string       `db:"id"`
	GrantID   string       `db:"grant_id"`
	Value     string       `db:"token"`
	Label     string       `db:"label"`
	ExpiresIn int          `db:"expires_in"`
	RevokedAt sql.NullTime `db:"revoked_at"`
	CreatedAt time.Time    `db:"created_at"`
}

type dbAccess struct {
	ID         string         `db:"id"`
	Type       string         `db:"type"`
	Actions    pq.StringArray `db:"actions"`
	Identifier string         `db:"identifier"`
	Locations  pq.StringArray `db:"locations"`
	DataTypes  pq.StringArray `db:"data_types"`
}

func lookupGrant(ctx context.Context, b Backends, id string) (*authorisation.Grant, error) {
	var dbg dbGrant
	err := b.DB().GetContext(ctx, &dbg, "SELECT id, client_id, state, continue_token, wait FROM authorisation_grants WHERE id=$1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, authorisation.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", authorisation.ErrInternal, err)
	}

	var client authorisation.Client
	err = b.DB().GetContext(ctx, &client, "SELECT id, url FROM authorisation_clients WHERE id=$1;", dbg.ClientID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", authorisation.ErrInternal, err)
	}

	resp := &authorisation.Grant{
		ID:            dbg.ID,
		State:         dbg.State,
		ContinueToken: dbg.ContinueToken,
		Wait:          dbg.Wait,
		Client:        client.URL,
	}

	var tokens []dbAccessToken
	err = b.DB().SelectContext(ctx, &tokens, "SELECT id, token, label, expires_in, created_at FROM authorisation_tokens WHERE grant_id=$1", id)
	if err != nil {
		return nil, fmt.Errorf("%w %s", authorisation.ErrInternal, err)
	}

	respTokens := make([]authorisation.AccessToken, len(tokens))
	for i, tk := range tokens {
		var access []dbAccess
		err = b.DB().SelectContext(ctx, &access, "SELECT id, type, actions, identifier, locations, data_types FROM authorisation_token_access WHERE token_id=$1", tk.ID)
		if err != nil {
			return nil, fmt.Errorf("%w %s", authorisation.ErrInternal, err)
		}

		respAccess := make([]authorisation.Access, len(access))
		for y, ac := range access {
			respAccess[y] = authorisation.Access{
				Type:       ac.Type,
				Actions:    ac.Actions,
				Locations:  ac.Locations,
				Datatypes:  ac.DataTypes,
				Identifier: ac.Identifier,
			}
		}

		respTokens[i] = authorisation.AccessToken{
			Value:     tk.Value,
			Access:    respAccess,
			ExpiresIn: int(math.Max(float64(tk.CreatedAt.Unix()+int64(tk.ExpiresIn)-time.Now().Unix()), 0)),
		}
	}

	resp.Tokens = respTokens

	return resp, nil
}

// VerifyRequestSig assumes that EdDSA is used with the ed25519 curve.
func VerifyRequestSig(ctx context.Context, req *http.Request, clientPaymentPointer string, requiredParts []string) bool {
	keySetURL := clientPaymentPointer
	if !strings.Contains(keySetURL, "jwks.json") {
		keySetURL += "/jwks.json"
	}

	keyID := httpmessagesignatures.ExtractKeyIDForSignature(ctx, req, "sig-1") // assume sig-1 for now
	if keyID == "" {
		log.Error("failed to extract keyID for signature", zap.String("signature id", "sig-1"))
		return false
	}

	resp, err := http.Get(keySetURL)
	if err != nil {
		log.Error("failed to get public key", zap.String("keySetURL", keySetURL), zap.String("keyID", keyID), zap.Error(err))
		return false
	}

	var keySet struct{ Keys []authorisation.Jwk }
	err = json.NewDecoder(resp.Body).Decode(&keySet)
	if err != nil {
		log.Error("failed to unmarshal keyset", zap.String("keySetURL", keySetURL), zap.String("keyID", keyID), zap.Error(err))
		return false
	}

	var key *authorisation.Jwk
	for _, k := range keySet.Keys {
		if k.Kid == keyID {
			key = &k
			break
		}
	}
	if key == nil {
		log.Error("public key not found", zap.String("keySetURL", keySetURL), zap.String("keyID", keyID))
		return false
	}
	if !key.IsEdDSAPublicKey() {
		log.Error("public key is not a edDSA-ed25519 public key", zap.String("keySetURL", keySetURL), zap.String("keyID", keyID))
		return false
	}

	publicKeyBytes, err := base64.StdEncoding.DecodeString(key.X)
	if err != nil {
		log.Error("failed to parse public key", zap.String("keySetURL", keySetURL), zap.String("keyID", keyID))
		return false
	}

	return httpmessagesignatures.VerifySignature(ctx, req, ed25519.PublicKey(publicKeyBytes), ed25519Verifier{}, requiredParts)
}

type ed25519Verifier struct {
}

func (s ed25519Verifier) Verify(publicKey crypto.PublicKey, digest []byte, signature []byte) bool {
	return ed25519.Verify(publicKey.(ed25519.PublicKey), digest, signature)
}

func Introspect(ctx context.Context, b Backends, token string) (*authorisation.Grant, error) {
	var dbToken dbAccessToken
	err := b.DB().GetContext(ctx, &dbToken, "SELECT id, grant_id, token, label, expires_in, revoked_at, created_at FROM authorisation_tokens WHERE token=$1;", token)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, authorisation.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", authorisation.ErrInternal, err)
	}

	if dbToken.RevokedAt.Valid {
		return nil, authorisation.ErrTokenRevoked
	}

	tokenExpiry := dbToken.CreatedAt.Add(time.Duration(dbToken.ExpiresIn) * time.Second)
	if time.Now().After(tokenExpiry) {
		return nil, authorisation.ErrTokenExpired
	}

	grant, err := lookupGrant(ctx, b, dbToken.GrantID)
	if err != nil {
		return nil, err
	}

	// don't leak all other tokens
	var filteredTokens []authorisation.AccessToken
	for _, t := range grant.Tokens {
		if t.Value == token {
			filteredTokens = append(filteredTokens, t)
			break
		}
	}
	grant.Tokens = filteredTokens

	return grant, nil
}
