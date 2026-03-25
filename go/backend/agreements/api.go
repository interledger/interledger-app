package agreements

import "context"

type Client interface {
	Sign(ctx context.Context, args *SignArgs) error
	GetSignatures(ctx context.Context, userID string) ([]Signature, error)
	Get(ctx context.Context, id string) (*Agreement, error)
	ListAffectedUserIDsPaginated(ctx context.Context, changes []AgreementChange, limit, offset int) ([]string, error)
	GetAgreementNamesSignedByUsersFromSet(ctx context.Context, userIDs []string, changes []AgreementChange) (map[string][]string, error)
	MarkUsersNotified(ctx context.Context, userIDs []string, changes []AgreementChange) error
}
