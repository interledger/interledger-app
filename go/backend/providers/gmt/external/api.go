package external

import "context"

type Client interface {
	InsertTransaction(ctx context.Context, tx InsertTransaction) (*WsResponse, error)
	UpdateTransactionStatus(ctx context.Context, tx UpdateTransactionStatus) (*WsResponse, error)
	OfacVerification(ctx context.Context, req OfacVerification) (*WsOfac, error)
	ComplianceCheck(ctx context.Context, req ComplianceCheck) (*WsResponse, error)
	SetVerified(ctx context.Context, req SetVerified) (*WsResult, error)
	ConfirmCollection(ctx context.Context, req ConfirmCollection) (*WsResponse, error)
	ConfirmPayment(ctx context.Context, req ConfirmPayment) (*WsResponse, error)
	GetNotifications(ctx context.Context) ([]*WsNotifications, error)
	RequestCancellation(ctx context.Context, txID, msg string) (*WsResponse, error)
}
