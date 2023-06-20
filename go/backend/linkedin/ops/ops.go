package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"gitlab.com/fynbos/backend/linkedin"
	"gitlab.com/fynbos/backend/linkedin/workflows"
	"go.temporal.io/api/enums/v1"
	temporal "go.temporal.io/sdk/client"
	"golang.org/x/oauth2"
	"time"
)

func CreateAuthURL(ctx context.Context, b Backends, args *CreateAuthURLArgs) (string, error) {
	oauthConfig := &oauth2.Config{
		ClientID:    args.ClientID,
		RedirectURL: args.RedirectURL,
		Scopes:      args.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL: args.AuthEndpoint,
		},
	}

	_, err := b.DB().ExecContext(ctx, `INSERT INTO linkedin_authorizations (client_id, state, scopes, wallet_id, redirect_url) VALUES ($1, $2, $3, $4, $5)`, args.ClientID, args.State, pq.Array(args.Scopes), args.WalletID, args.RedirectURL)
	if err != nil {
		return "", fmt.Errorf("%w %s", linkedin.ErrInternal, err)
	}

	url := oauthConfig.AuthCodeURL(args.State)

	return url, nil
}

func CreateConnection(ctx context.Context, b Backends, args *linkedin.CreateConnectionArgs) (*linkedin.Connection, error) {
	var authorization LinkedinAuth

	err := b.DB().GetContext(ctx, &authorization, "SELECT * FROM linkedin_authorizations WHERE state = $1", args.State)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w no authorization found for state %q", linkedin.ErrNotFound, args.State)
		}
		return nil, fmt.Errorf("%w %s", linkedin.ErrInternal, err)
	}
	if args.AuthCode == "" {
		return nil, fmt.Errorf("%w couldn't find an auth code for this grant", linkedin.ErrInternal)
	}

	token, err := b.External().CreateToken(ctx, args.AuthCode)
	if err != nil {
		return nil, fmt.Errorf("%w %s", linkedin.ErrInternal, err)
	}

	user, err := b.External().GetAuthorizedUser(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("%w %s", linkedin.ErrInternal, err)
	}

	var connection linkedin.Connection
	query := `
		INSERT INTO linkedin_connections (user_id, wallet_id, access_token, refresh_token, token_type, scopes, username, expiry, client_id)
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
		return nil, fmt.Errorf("%w %s", linkedin.ErrInternal, err)
	}

	return &connection, nil
}

func GetWalletConnections(ctx context.Context, b Backends, walletID string) ([]linkedin.Connection, error) {
	var connections []linkedin.Connection

	err := b.DB().SelectContext(ctx, &connections, "SELECT * FROM linkedin_connections WHERE wallet_id $1", walletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", linkedin.ErrInternal, err)
	}

	return connections, nil
}

func Share(ctx context.Context, b Backends, connectionID, text string) (string, error) {
	conntection, err := getConnection(ctx, b, connectionID)
	if err != nil {
		return "", fmt.Errorf("%w %s", linkedin.ErrInternal, err)
	}

	return b.External().Share(ctx, conntection, text)
}

func PublishPublicProof(ctx context.Context, b Backends, identityID string) error {
	workflowOptions := temporal.StartWorkflowOptions{
		ID:                       "publish-linkedin-public-proof" + identityID,
		TaskQueue:                "backend",
		WorkflowExecutionTimeout: time.Hour * 24 * 8, // 8 days
		WorkflowIDReusePolicy:    enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
	}

	_, err := b.Temporal().ExecuteWorkflow(ctx, workflowOptions, workflows.PublishLinkedinProofWorkflow, identityID)
	if err != nil {
		return fmt.Errorf("%w %s", linkedin.ErrInternal, err)
	}

	return nil
}

func getConnection(ctx context.Context, b Backends, id string) (*linkedin.Connection, error) {
	var connection linkedin.Connection
	err := b.DB().GetContext(ctx, &connection, "SELECT * FROM linkedin_connections WHERE id = $1", id)
	if err != nil {
		return nil, err
	}

	return &connection, nil
}
