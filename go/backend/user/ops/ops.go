package ops

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"

	client "github.com/ory/kratos-client-go"
	"gitlab.com/fynbos/backend/user"
)

const (
	kratosTimeout    = 1500 * time.Millisecond
	kratosCookieName = "ory_kratos_session"
	aal2RequiredErrorID = "session_aal2_required"
)

type sessionRetrievalErrorResponse struct {
	Error *struct {
		ID      string `json:"id"`
	} `json:"error,omitempty"`
}

func getSessionRetrievalError(resp *http.Response, err error) (error) {
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

	session, resp, err := b.Kratos().FrontendApi.ToSession(ctx).Cookie(kratosCookieName + "=" + cookie).Execute()
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

	session, resp, err := b.Kratos().FrontendApi.ToSession(ctx).XSessionToken(token).Execute()
	if err != nil {
		return nil, getSessionRetrievalError(resp, err)
	}

	u := convertTraits(session.Identity.Id, session.Identity.Traits)
	return &u, nil
}

func GetUser(ctx context.Context, b Backends, userID string) (*user.User, error) {
	id, _, err := b.Kratos().IdentityApi.GetIdentity(ctx, userID).Execute()
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

// TODO: Modify?
func ListUsers(ctx context.Context, b Backends, walletID string) ([]user.User, error) {
	if walletID == wallets.WebMonetizationWalletID {
		return []user.User{
			{
				ID:          "6b5ada19-1638-4c09-a0f6-9cdbb34abc42",
				Email:       "openpayments.dev@fynbos.dev",
				PhoneNumber: "",
			},
		}, nil
	}
	if walletID == wallets.AstraBusinessWalletID {
		return []user.User{
			{
				ID:          "a734d15a-b4d5-4a78-a434-1b03f39ecc44",
				Email:       "astra@fynbos.dev",
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
				id, _, err := b.Kratos().IdentityApi.GetIdentity(ctx, uID).Execute()
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
