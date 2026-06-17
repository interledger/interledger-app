package ops

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/interledger/interledger-app/go/backend/twitter/workflows"
	"go.temporal.io/api/enums/v1"
	temporal "go.temporal.io/sdk/client"
	"io"
	"time"

	"github.com/interledger/interledger-app/go/backend/twitter"
	"github.com/interledger/interledger-app/go/backend/twitter/external"
	"github.com/lib/pq"
	"golang.org/x/oauth2"
)

func CreateAuthURL(ctx context.Context, b Backends, args *CreateAuthURLArgs) (string, error) {
	oauthConfig := &oauth2.Config{
		ClientID:    args.ClientID,
		RedirectURL: args.RedirectURL,
		Scopes:      args.Scopes,
		Endpoint:    oauth2.Endpoint{AuthURL: args.AuthEndpoint},
	}

	codeVerifier, err := randomBytesInBase64URL(32)
	if err != nil {
		return "", fmt.Errorf("%w %s", twitter.ErrInternal, err)
	}
	sha := sha256.New()
	sha.Write([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(sha.Sum(nil))

	var authId string

	err = b.DB().GetContext(ctx, &authId, "INSERT INTO twitter_authorizations (client_id, state, code_verifier, scopes, wallet_id, redirect_url) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id", args.ClientID, args.State, codeVerifier, pq.Array(args.Scopes), args.WalletID, args.RedirectURL)
	if err != nil {
		return "", fmt.Errorf("%w %s", twitter.ErrInternal, err)
	}

	url := oauthConfig.AuthCodeURL(args.State,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	return url, nil
}

func CreateConnection(ctx context.Context, b Backends, args *twitter.CreateConnectionArgs) (*twitter.Connection, error) {
	var authorization TwitterAuth

	err := b.DB().GetContext(ctx, &authorization, "SELECT * FROM twitter_authorizations WHERE state = $1", args.State)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w no authorization found for state %q", twitter.ErrNotFound, args.State)
		}
		return nil, fmt.Errorf("%w %s", twitter.ErrInternal, err)
	}
	if args.AuthCode == "" {
		return nil, fmt.Errorf("%w couldn't find an auth code for this grant", twitter.ErrInternal)
	}

	token, err := b.External().CreateToken(ctx, &external.CreateTokenArgs{
		AuthCode:     args.AuthCode,
		CodeVerifier: authorization.CodeVerifier,
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", twitter.ErrInternal, err)
	}

	user, err := b.External().GetAuthorizedUser(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("%w %s", twitter.ErrInternal, err)
	}

	var connection twitter.Connection
	query := `
		INSERT INTO twitter_connections (user_id, wallet_id, access_token, refresh_token, token_type, scopes, username, expiry, client_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id, wallet_id) 
		DO UPDATE SET 
			access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			token_type = EXCLUDED.token_type,
			scopes = EXCLUDED.scopes,
			username = EXCLUDED.username,
			expiry = EXCLUDED.expiry,
			updated_at = NOW()
		RETURNING *;
	`
	err = b.DB().GetContext(ctx, &connection, query,
		user.ID, authorization.WalletID, token.AccessToken, token.RefreshToken, token.TokenType, pq.Array(authorization.Scopes), user.Username, token.Expiry, authorization.ClientID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", twitter.ErrInternal, err)
	}

	return &connection, nil
}

func GetWalletConnections(ctx context.Context, b Backends, id string) ([]twitter.Connection, error) {
	var connections []twitter.Connection

	err := b.DB().SelectContext(ctx, &connections, "SELECT * FROM twitter_connections WHERE wallet_id = $1", id)
	if err != nil {
		return nil, fmt.Errorf("%w %s", twitter.ErrInternal, err)
	}

	return connections, nil
}

func PostTweet(ctx context.Context, b Backends, id string, text string) (*twitter.Tweet, error) {
	connection, err := getConnection(ctx, b, id)
	if err != nil {
		return nil, fmt.Errorf("twitter: could not get connection %w %s", twitter.ErrInternal, err)
	}

	oauthToken := oauth2.Token{
		AccessToken:  connection.AccessToken,
		RefreshToken: connection.RefreshToken,
		TokenType:    connection.TokenType,
		Expiry:       connection.Expiry,
	}

	// TODO handle refresh the token if not valid
	if !oauthToken.Valid() {
		connection, err = refreshToken(ctx, b, id)
		if err != nil {
			return nil, fmt.Errorf("twitter: could not refresh token %w %s", twitter.ErrInternal, err)
		}
	}

	oauthToken = oauth2.Token{
		AccessToken:  connection.AccessToken,
		RefreshToken: connection.RefreshToken,
		TokenType:    connection.TokenType,
		Expiry:       connection.Expiry,
	}

	tweet, err := b.External().PostTweet(ctx, &oauthToken, text)
	if err != nil {
		return nil, fmt.Errorf("twitter: could not post tweet %w %s", twitter.ErrInternal, err)
	}

	return tweet, nil
}

func PublishTweetProof(ctx context.Context, b Backends, id string) (string, error) {
	workflowOptions := temporal.StartWorkflowOptions{
		ID:                       "publish-tweet-proof" + id,
		TaskQueue:                "backend",
		WorkflowExecutionTimeout: time.Hour * 24 * 8, // 8 days
		WorkflowIDReusePolicy:    enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
	}

	workflow, err := b.Temporal().ExecuteWorkflow(ctx, workflowOptions, workflows.PublishTwitterProofWorkflow, id)
	if err != nil {
		return "", fmt.Errorf("%w %s", twitter.ErrInternal, err)
	}

	var tweetUrl string
	err = workflow.Get(ctx, &tweetUrl)
	if err != nil {
		return "", fmt.Errorf("%w %s", twitter.ErrInternal, err)
	}

	return tweetUrl, nil
}

func GetTweet(ctx context.Context, b Backends, id string) (*twitter.Tweet, error) {
	return b.External().GetTweet(ctx, id)
}

func getConnection(ctx context.Context, b Backends, id string) (*twitter.Connection, error) {
	var connection twitter.Connection
	err := b.DB().GetContext(ctx, &connection, "SELECT * FROM twitter_connections WHERE id = $1", id)
	if err != nil {
		return nil, err
	}

	return &connection, nil
}

func refreshToken(ctx context.Context, b Backends, id string) (*twitter.Connection, error) {
	return nil, nil
}

func randomBytesInBase64URL(count int) (string, error) {
	buf := make([]byte, count)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		return "", fmt.Errorf("%w %s", twitter.ErrInternal, err)
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}
