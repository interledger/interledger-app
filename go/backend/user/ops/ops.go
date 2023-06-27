package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	client "github.com/ory/kratos-client-go"
	"gitlab.com/fynbos/backend/user"
)

const (
	kratosTimeout    = 500 * time.Millisecond
	kratosCookieName = "ory_kratos_session"
)

func UserForCookie(ctx context.Context, b Backends, cookie string) (*user.User, error) {
	if cookie == "" {
		return nil, user.ErrNoUserFound
	}

	ctx, cancel := context.WithTimeout(ctx, kratosTimeout)
	defer cancel()

	session, resp, err := b.Kratos().V0alpha2Api.ToSession(ctx).Cookie(kratosCookieName + "=" + cookie).Execute()
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusUnauthorized {
			return nil, nil
		}
		return nil, err
	}

	u := convertTraits(session.Identity.Id, session.Identity.Traits)
	return &u, nil
}

func GetUser(ctx context.Context, b Backends, userID string) (*user.User, error) {
	id, _, err := b.Kratos().V0alpha2Api.AdminGetIdentity(ctx, userID).Execute()
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

func ListUsers(ctx context.Context, b Backends, walletID string) ([]user.User, error) {
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
				id, _, err := b.Kratos().V0alpha2Api.AdminGetIdentity(ctx, uID).Execute()
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
