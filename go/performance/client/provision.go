package client

import (
	"context"

	pb "github.com/interledger/interledger-app/go/proto/backend/v1"
)

// The calls below exist only for the provision command. They are on Wallet
// because they need the same connection pool and per-call auth, but none of them
// run during a measured load run.

// SetSignupUserData starts a signup. It is called before any identity exists, so
// the Wallet's token may be empty.
func (w *Wallet) SetSignupUserData(ctx context.Context, req *pb.SetSignupUserDataRequest) (*pb.SetSignupUserDataResponse, error) {
	callCtx, cancel := w.ctx(ctx)
	defer cancel()
	return w.client.SetSignupUserData(callCtx, req)
}

// CompleteSignup links a Kratos identity to a signup record.
func (w *Wallet) CompleteSignup(ctx context.Context, signupID, userID string) error {
	callCtx, cancel := w.ctx(ctx)
	defer cancel()
	_, err := w.client.CompleteSignup(callCtx, &pb.CompleteSignupRequest{Id: signupID, UserId: userID})
	return err
}

// CreateUserDefaultWallet creates the platform wallet for a user.
func (w *Wallet) CreateUserDefaultWallet(ctx context.Context, userID string) error {
	callCtx, cancel := w.ctx(ctx)
	defer cancel()
	_, err := w.client.CreateUserDefaultWallet(callCtx, &pb.CreateUserDefaultWalletRequest{UserID: userID})
	return err
}

// CreateWalletAddress assigns a wallet address and registers the matching payment
// pointer with Rafiki.
func (w *Wallet) CreateWalletAddress(ctx context.Context, req *pb.CreateWalletAddressRequest) error {
	callCtx, cancel := w.ctx(ctx)
	defer cancel()
	_, err := w.client.CreateWalletAddress(callCtx, req)
	return err
}

// AddXagoBalanceAccount creates a Xago balance linked account to fund.
func (w *Wallet) AddXagoBalanceAccount(ctx context.Context, req *pb.AddXagoBalanceAccountRequest) (*pb.LinkedAccount, error) {
	callCtx, cancel := w.ctx(ctx)
	defer cancel()
	return w.client.AddXagoBalanceAccount(callCtx, req)
}

// DepositTestXago credits the wallet's Xago balance. The backend refuses this
// call when the environment mode is prod.
func (w *Wallet) DepositTestXago(ctx context.Context) error {
	callCtx, cancel := w.ctx(ctx)
	defer cancel()
	_, err := w.client.DepositTestXago(callCtx, &pb.Empty{})
	return err
}

// GetGatehubOnboardingWidget creates the managed GateHub user and linked account.
func (w *Wallet) GetGatehubOnboardingWidget(ctx context.Context, req *pb.Empty) (*pb.GatehubWidget, error) {
	callCtx, cancel := w.ctx(ctx)
	defer cancel()
	return w.client.GetGatehubOnboardingWidget(callCtx, req)
}

// GetKYCProviderWidget fetches the PTI widget details for a wallet.
func (w *Wallet) GetKYCProviderWidget(ctx context.Context, req *pb.GetKYCProviderWidgetRequest) (*pb.KYCProviderWidget, error) {
	callCtx, cancel := w.ctx(ctx)
	defer cancel()
	return w.client.GetKYCProviderWidget(callCtx, req)
}

// SetKYCStatusPending kicks off the PTI assessment workflow for a wallet.
func (w *Wallet) SetKYCStatusPending(ctx context.Context, req *pb.Empty) (*pb.Empty, error) {
	callCtx, cancel := w.ctx(ctx)
	defer cancel()
	return w.client.SetKYCStatusPending(callCtx, req)
}

// CreatePtiBankAccount creates the ACH source account required for PTI deposits.
func (w *Wallet) CreatePtiBankAccount(ctx context.Context, req *pb.CreatePtiBankAccountRequest) (*pb.LinkedAccount, error) {
	callCtx, cancel := w.ctx(ctx)
	defer cancel()
	return w.client.CreatePtiBankAccount(callCtx, req)
}

// DepositBalance moves funds between linked accounts for PTI funding.
func (w *Wallet) DepositBalance(ctx context.Context, req *pb.TransferBalanceRequest) (*pb.Payment, error) {
	callCtx, cancel := w.ctx(ctx)
	defer cancel()
	return w.client.DepositBalance(callCtx, req)
}

// PtiCreateDeposit executes the PTI deposit workflow for a payment.
func (w *Wallet) PtiCreateDeposit(ctx context.Context, req *pb.PtiCreateDepositRequest) (*pb.Empty, error) {
	callCtx, cancel := w.ctx(ctx)
	defer cancel()
	return w.client.PtiCreateDeposit(callCtx, req)
}

// GetPtiBalances lists the wallet's PTI balances.
func (w *Wallet) GetPtiBalances(ctx context.Context) (*pb.GetPtiBalancesResponse, error) {
	callCtx, cancel := w.ctx(ctx)
	defer cancel()
	return w.client.GetPtiBalances(callCtx, &pb.Empty{})
}
