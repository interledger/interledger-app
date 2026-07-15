package ops

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/interledger/interledger-app/go/backend/country"
	"github.com/interledger/interledger-app/go/backend/wallets"
	"github.com/interledger/interledger-app/go/log"
	"github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/interledger/interledger-app/go/backend/user"
	client "github.com/ory/kratos-client-go"
)

const (
	kratosTimeout       = 1500 * time.Millisecond
	kratosCookieName    = "ory_kratos_session"
	aal2RequiredErrorID = "session_aal2_required"
)

type sessionRetrievalErrorResponse struct {
	Error *struct {
		ID string `json:"id"`
	} `json:"error,omitempty"`
}

func getSessionRetrievalError(resp *http.Response, err error) error {
	if resp != nil && resp.StatusCode == http.StatusForbidden {
		var sessionErrResp sessionRetrievalErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&sessionErrResp)
		if err != nil {
			log.Error("Error decoding error body of kratos to session response.", zap.Error(err))
		}

		if sessionErrResp.Error.ID == aal2RequiredErrorID {
			log.Info("User must complete 2FA, as AAL2 is required.")
			return user.ErrAAL2Required
		}
	}

	if resp != nil && resp.StatusCode == http.StatusUnauthorized {
		return user.ErrAAL1Required
	}

	return err
}

func UserForCookie(ctx context.Context, b Backends, cookie string) (*user.User, error) {
	if cookie == "" {
		return nil, user.ErrNoUserFound
	}

	ctx, cancel := context.WithTimeout(ctx, kratosTimeout)
	defer cancel()

	session, resp, err := b.Kratos().FrontendAPI.ToSession(ctx).Cookie(kratosCookieName + "=" + cookie).Execute()
	if err != nil {
		return nil, getSessionRetrievalError(resp, err)
	}

	u := convertTraits(session.Identity.Id, session.Identity.Traits)
	return &u, nil
}

func UserForToken(ctx context.Context, b Backends, token string) (*user.User, error) {
	if token == "" {
		return nil, user.ErrNoUserFound
	}

	ctx, cancel := context.WithTimeout(ctx, kratosTimeout)
	defer cancel()

	session, resp, err := b.Kratos().FrontendAPI.ToSession(ctx).XSessionToken(token).Execute()
	if err != nil {
		return nil, getSessionRetrievalError(resp, err)
	}

	u := convertTraits(session.Identity.Id, session.Identity.Traits)
	return &u, nil
}

func GetUser(ctx context.Context, b Backends, userID string) (*user.User, error) {
	id, _, err := b.Kratos().IdentityAPI.GetIdentity(ctx, userID).Execute()
	if err != nil {
		return nil, fmt.Errorf("%w %s", user.ErrInternal, err)
	}

	u := convertTraits(id.Id, id.Traits)
	return &u, nil
}

func convertTraits(userID string, traits interface{}) user.User {
	traitsMap := traits.(map[string]interface{})
	u := user.User{
		ID:          userID,
		Email:       traitsMap["email"].(string),
		PhoneNumber: traitsMap["phone"].(string),
		Country:     country.ParseCountry(traitsMap["countryCode"].(string)),
		FirstName:   traitsMap["firstName"].(string),
		LastName:    traitsMap["lastName"].(string),
	}
	// All trait values:  "email", "phone", "firstName", "lastName", "countryCode"
	return u
}

func UserForContext(ctx context.Context) (*user.User, error) {
	u, ok := ctx.Value(user.CtxKey).(*user.User)
	if !ok || u == nil {
		return nil, user.ErrNoUserFound
	}
	return u, nil
}

// GetTotpURL retrieves the TOTP URL for the given userID.
//
// Errors:
//   - user.ErrNoCredentials if the identity has no credentials
//   - user.ErrTotpNotConfigured if the user has no TOTP credential configured
//   - user.ErrInvalidTotpConfig if the TOTP configuration is malformed
//   - user.ErrInternal if fetching the identity from Kratos fails
//
// Usage:
//
//	totpURL, err := GetTotpURL(ctx, b, userID)
//	if err != nil {
//		// Check errors if more granularity is needed
//		return err
//	}
//
//	// Use totpURL
func GetTotpURL(ctx context.Context, b Backends, userID string) (string, error) {
	identity, _, err := b.Kratos().IdentityAPI.GetIdentity(ctx, userID).IncludeCredential([]string{"totp"}).Execute()
	if err != nil {
		return "", fmt.Errorf("%w %s", user.ErrInternal, err)
	}

	if identity.Credentials == nil {
		return "", user.ErrNoCredentials
	}

	totpURL, err := searchTotpURL(*identity.Credentials)
	if err != nil {
		return "", err
	}

	return totpURL, nil
}

func GetUserIDForWallet(ctx context.Context, b Backends, walletID string) (string, error) {
	var userID string
	err := b.DB().GetContext(ctx, &userID, "SELECT user_id FROM user_wallets WHERE wallet_id = $1", walletID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", user.ErrNoUserFound
		}
		return "", fmt.Errorf("%w %s", user.ErrInternal, err)
	}

	return userID, nil
}

// FindWalletIDByEmail looks up a Kratos identity by credential identifier (email)
// and returns the associated wallet ID from user_wallets. Returns "" when no match.
func FindWalletIDByEmail(ctx context.Context, b Backends, email string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, kratosTimeout)
	defer cancel()

	// Kratos admin API is at server index 1.
	kratosCtx := context.WithValue(ctx, client.ContextServerIndex, 1)
	identities, _, err := b.Kratos().IdentityAPI.ListIdentities(kratosCtx).CredentialsIdentifier(email).Execute()
	if err != nil {
		return "", fmt.Errorf("%w %s", user.ErrInternal, err)
	}
	if len(identities) == 0 {
		return "", nil
	}

	// A Kratos identity can be linked to more than one wallet. Return the most
	// recently created one so the result is deterministic.
	var walletID string
	err = b.DB().GetContext(ctx, &walletID,
		`SELECT uw.wallet_id FROM user_wallets uw
		 JOIN wallets w ON w.id = uw.wallet_id
		 WHERE uw.user_id = $1
		 ORDER BY w.created_at DESC
		 LIMIT 1`,
		identities[0].Id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("%w %s", user.ErrInternal, err)
	}
	return walletID, nil
}

// maxIdentifierPrefixWalletIDs defensively bounds the wallet-ID set resolved
// from a Kratos identifier-prefix search (e.g. a broad 1-char email prefix).
const maxIdentifierPrefixWalletIDs = 1000

// FindWalletIDsByIdentifierPrefix resolves ALL Kratos identities whose
// credential identifier (email or phone) starts with term to their wallet
// IDs. Unlike FindWalletIDByEmail (which takes only the first match), every
// matching identity's wallet is returned, deduplicated.
func FindWalletIDsByIdentifierPrefix(ctx context.Context, b Backends, term string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, kratosTimeout)
	defer cancel()

	//Kratos SDK generates client code from OpenAPI spec with multiple servers entries (public API, admin API — usually index 0 = public, index 1 = admin)
	kratosCtx := context.WithValue(ctx, client.ContextServerIndex, 1)

	// preview_credentials_identifier_similar is a preview-flagged Ory API with no
	// documented contract; behavior below is observed, not guaranteed, and may
	// change on server upgrades. Official openapi schema:
	// https://github.com/ory/kratos-client-go/blob/v26.2.0/api/openapi.yaml
	identities, _, err := b.Kratos().IdentityAPI.ListIdentities(kratosCtx).
		PreviewCredentialsIdentifierSimilar(strings.ToLower(term)).
		PerPage(int64(maxIdentifierPrefixWalletIDs)).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("%w %s", user.ErrInternal, err)
	}
	if len(identities) == 0 {
		return nil, nil
	}

	userIDs := make([]string, len(identities))
	for i, id := range identities {
		userIDs[i] = id.Id
	}

	var walletIDs []string
	err = b.DB().SelectContext(ctx, &walletIDs,
		`SELECT DISTINCT wallet_id FROM user_wallets WHERE user_id = ANY($1)`,
		pq.Array(userIDs),
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", user.ErrInternal, err)
	}

	if len(walletIDs) > maxIdentifierPrefixWalletIDs {
		log.Warn("FindWalletIDsByIdentifierPrefix: resolved wallet-ID set exceeds cap, truncating",
			zap.Int("resolved", len(walletIDs)), zap.Int("cap", maxIdentifierPrefixWalletIDs))
		walletIDs = walletIDs[:maxIdentifierPrefixWalletIDs]
	}

	return walletIDs, nil
}

// TODO: Modify?
func ListUsers(ctx context.Context, b Backends, walletID string) ([]user.User, error) {
	if walletID == wallets.WebMonetizationWalletID {
		return []user.User{
			{
				ID:          "6b5ada19-1638-4c09-a0f6-9cdbb34abc42",
				Email:       "openpayments.dev@interledger.test",
				PhoneNumber: "",
			},
		}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, kratosTimeout)
	defer cancel()

	var userIDs []string
	err := b.DB().SelectContext(ctx, &userIDs, "SELECT user_id FROM user_wallets WHERE wallet_id=$1", walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, user.ErrNoUserFound
	}
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	var mx sync.Mutex

	// Required for kratos to use admin server
	ctx = context.WithValue(ctx, client.ContextServerIndex, 1)

	var resp []user.User
	var anyErr error

	// Required to check so test can pass due to kratos not being mocked.
	if b.Kratos() != nil {
		for _, userID := range userIDs {
			wg.Add(1)
			go func(uID string) {
				defer wg.Done()
				id, _, err := b.Kratos().IdentityAPI.GetIdentity(ctx, uID).Execute()
				if err != nil {
					anyErr = err
					return
				}

				// lock
				mx.Lock()
				defer mx.Unlock()
				resp = append(resp, convertTraits(id.Id, id.Traits))
			}(userID)
		}
	}

	wg.Wait()
	if anyErr != nil {
		return nil, anyErr
	}

	return resp, nil
}

func CheckUserTotpEnabled(ctx context.Context, b Backends, identityID string) (bool, error) {
	identity, _, err := b.Kratos().IdentityAPI.GetIdentity(ctx, identityID).Execute()
	if err != nil {
		return false, fmt.Errorf("%w %s", user.ErrInternal, err)
	}

	creds := *identity.Credentials

	totp, ok := creds["totp"]
	if !ok {
		return false, nil
	}

	return len(totp.Identifiers) > 0, nil
}

func Delete2FATotpEnrollment(ctx context.Context, b Backends, identityID string) error {
	identity, _, err := b.Kratos().IdentityAPI.GetIdentity(ctx, identityID).Execute()
	if err != nil {
		return fmt.Errorf("%w %s", user.ErrInternal, err)
	}

	req := b.Kratos().IdentityAPI.DeleteIdentityCredentials(ctx, identity.Id, "totp")
	_, err = req.Execute()
	if err != nil {
		return fmt.Errorf("%w %s", user.ErrInternal, err)
	}

	return nil
}

// searchTotpURL searches for the TOTP URL in the provided identity credentials.
// It returns the TOTP URL if found.
// Errors:
//   - user.ErrTotpNotConfigured if no TOTP credential exists or the TOTP URL is missing.
//   - user.ErrInvalidTotpConfig if the TOTP configuration is malformed.
func searchTotpURL(credentials map[string]client.IdentityCredentials) (string, error) {
	for _, cred := range credentials {
		if cred.Type == nil {
			continue
		}

		if *cred.Type != "totp" {
			continue
		}

		if cred.Config == nil {
			return "", user.ErrTotpNotConfigured
		}

		raw, exists := cred.Config["totp_url"]
		if !exists {
			return "", user.ErrTotpNotConfigured
		}

		url, ok := raw.(string)
		if !ok {
			return "", user.ErrInvalidTotpConfig
		}

		return url, nil
	}

	return "", user.ErrTotpNotConfigured
}

func SetPhoneVerified(ctx context.Context, b Backends, userID string) error {
	// Required for kratos to use admin server
	ctx = context.WithValue(ctx, client.ContextServerIndex, 1)

	identity, _, err := b.Kratos().IdentityAPI.GetIdentity(ctx, userID).Execute()
	if err != nil {
		return fmt.Errorf("%w %s", user.ErrInternal, err)
	}

	traits, ok := identity.Traits.(map[string]any)
	if !ok {
		return fmt.Errorf("%w: invalid traits format", user.ErrInternal)
	}
	traits["phoneVerified"] = true

	update := client.UpdateIdentityBody{Traits: traits}
	_, _, err = b.Kratos().IdentityAPI.UpdateIdentity(ctx, userID).UpdateIdentityBody(update).Execute()
	if err != nil {
		return fmt.Errorf("%w %s", user.ErrInternal, err)
	}

	return nil
}

func UpdateUserPhone(ctx context.Context, b Backends, userID string, phone string) error {
	// Required for kratos to use admin server
	ctx = context.WithValue(ctx, client.ContextServerIndex, 1)

	identity, _, err := b.Kratos().IdentityAPI.GetIdentity(ctx, userID).Execute()
	if err != nil {
		return fmt.Errorf("%w %s", user.ErrInternal, err)
	}

	traits, ok := identity.Traits.(map[string]any)
	if !ok {
		return fmt.Errorf("%w: invalid traits format", user.ErrInternal)
	}
	traits["phone"] = phone
	traits["phoneVerified"] = false

	update := client.UpdateIdentityBody{Traits: traits}
	_, response, err := b.Kratos().IdentityAPI.UpdateIdentity(ctx, userID).UpdateIdentityBody(update).Execute()
	if err != nil {
		if response != nil && response.StatusCode == 400 {
			return fmt.Errorf("%w %s", user.ErrInvalidArgument, err)
		}
		if response != nil && response.StatusCode == http.StatusConflict {
			return fmt.Errorf("%w %s", user.ErrDuplicatePhone, err)
		}
		return fmt.Errorf("%w %s", user.ErrInternal, err)
	}

	return nil
}
