package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/bxcodec/faker/v3"
	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	accounttransactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/deposits"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/graph/generated"
	_identity "gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/withdrawals"
)

func (r *accountResolver) RecentTransactions(ctx context.Context, obj *generated.Account) ([]*generated.Transaction, error) {
	var trxs []*generated.Transaction
	err := crdbsqlx.ExecuteTx(ctx, r.Db, nil, func(tx *sqlx.Tx) error {
		dbTrxs, err := r.AccountTransactions.GetByAccount(
			ctx,
			tx,
			&accounttransactions.GetByAccountArgs{
				AccountID: obj.ID,
				Limit:     3,
				OrderBy:   "DESC",
			})
		if err != nil {
			return err
		}

		for _, trx := range dbTrxs {
			trxs = append(trxs, &generated.Transaction{
				ID:          trx.ID,
				Type:        generated.TransactionType(strings.ToUpper(trx.Type)),
				Description: trx.Description,
				Amount:      fmt.Sprintf("$ %.2f", float64(trx.NetAmount)/float64(100)),
				Timestamp:   trx.CreatedAt,
				Status:      trx.State,
			})
		}

		return nil
	})
	if err != nil {
		switch err.(type) {
		default:
			InternalServerError(ctx)
			return nil, nil
		}
	}

	return trxs, nil
}

func (r *mutationResolver) OnboardAccount(ctx context.Context) (*generated.CreateAccountMutationResponse, error) {
	user, err := r.UserService.ForContext(ctx)
	if err != nil {
		ForbiddenError(ctx)
		return nil, nil
	}

	// Check if account exists
	acc, _ := r.AccountService.GetByIdentityID(ctx, user.ID)
	if acc != nil {
		floatBalance := float64(acc.AvailableBalance) / float64(100)
		return &generated.CreateAccountMutationResponse{
			Code:    "200",
			Success: true,
			Message: "Onboarded account.",
			Account: &generated.Account{
				ID:      acc.ID,
				Balance: fmt.Sprintf("$ %.2f", floatBalance),
			},
		}, nil
	}

	acc, err = r.Os.CreateAccount(ctx, &onboarding.CreateAccountArgs{
		IdentityID:   user.ID,
		FirstName:    faker.FirstName(),
		LastName:     faker.FirstName(),
		MobileNumber: faker.E164PhoneNumber(),
		Email:        user.Email,
		Country:      "US",
	})
	if err != nil {
		InternalServerError(ctx)
		return nil, nil
	}

	acc, err = r.Os.VerifyAccount(ctx, &onboarding.VerifyAccountArgs{
		IdentityID:  user.ID,
		AccountID:   acc.ID,
		DateOfBirth: faker.Date(),
		Address:     []string{faker.Name()},
		State:       faker.FirstName(),
		City:        faker.FirstName(),
		PostalCode:  faker.Currency(),
		TaxIDNumber: faker.CCNumber(),
	})
	if err != nil {
		switch err.(type) {
		case *_identity.ErrNotFound:
			NotFoundError(ctx)
			return nil, nil
		case *_identity.ErrInvalidArgument:
			InvalidArgument(ctx, err.Error())
			return nil, nil
		default:
			InternalServerError(ctx)
			return nil, nil
		}
	}

	floatBalance := float64(acc.AvailableBalance) / float64(100)
	return &generated.CreateAccountMutationResponse{
		Code:    "200",
		Success: true,
		Message: "Onboarded account.",
		Account: &generated.Account{
			ID:      acc.ID,
			Balance: fmt.Sprintf("$ %.2f", floatBalance),
		},
	}, nil
}

func (r *mutationResolver) CreateAccount(ctx context.Context, input generated.CreateAccountInput) (*generated.CreateAccountMutationResponse, error) {
	user, err := r.UserService.ForContext(ctx)
	if err != nil {
		ForbiddenError(ctx)
		return nil, nil
	}

	acc, err := r.Os.CreateAccount(ctx, &onboarding.CreateAccountArgs{
		IdentityID:   user.ID,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		MobileNumber: input.MobileNumber,
		Email:        user.Email,
		Country:      input.Country,
	})
	if err != nil {
		switch err.(type) {
		case *_identity.ErrInvalidArgument:
			InvalidArgument(ctx, err.Error())
			return nil, nil
		default:
			InternalServerError(ctx)
			return nil, nil
		}
	}

	floatBalance := float64(acc.AvailableBalance) / float64(100)
	return &generated.CreateAccountMutationResponse{
		Code:    "200",
		Success: true,
		Message: "Created account.",
		Account: &generated.Account{
			ID:      acc.ID,
			Balance: fmt.Sprintf("$ %.2f", floatBalance),
		},
	}, nil
}

func (r *mutationResolver) VerifyAccount(ctx context.Context, input generated.VerifyAccountInput) (*generated.VerifyAccountMutationResponse, error) {
	user, err := r.UserService.ForContext(ctx)
	if err != nil {
		ForbiddenError(ctx)
		return nil, nil
	}

	acc, err := r.AccountService.GetByIdentityID(ctx, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, accounts.ErrInvalidArgument):
			InvalidArgument(ctx, err.Error())
			return nil, nil
		case errors.Is(err, accounts.ErrNotFound):
			return &generated.VerifyAccountMutationResponse{
				Code:    "404",
				Success: false,
				Message: "Verification failed: Account not found.",
			}, nil
		default:
			InternalServerError(ctx)
			return nil, nil
		}
	}

	acc, err = r.Os.VerifyAccount(ctx, &onboarding.VerifyAccountArgs{
		IdentityID:  user.ID,
		AccountID:   acc.ID,
		DateOfBirth: input.DateOfBirth,
		Address:     input.Address,
		State:       input.State,
		City:        input.City,
		PostalCode:  input.PostalCode,
		TaxIDNumber: input.TaxIDNumber,
	})
	if err != nil {
		switch err.(type) {
		case *_identity.ErrNotFound:
			NotFoundError(ctx)
			return nil, nil
		case *_identity.ErrInvalidArgument:
			InvalidArgument(ctx, err.Error())
			return nil, nil
		default:
			InternalServerError(ctx)
			return nil, nil
		}
	}

	floatBalance := float64(acc.AvailableBalance) / float64(100)
	return &generated.VerifyAccountMutationResponse{
		Code:    "200",
		Success: true,
		Message: "Verified.",
		Account: &generated.Account{
			ID:      acc.ID,
			Balance: fmt.Sprintf("$ %.2f", floatBalance),
		},
	}, nil
}

func (r *mutationResolver) LinkUsdBankAccount(ctx context.Context, input generated.LinkUsdBankAccountInput) (*generated.LinkFundingSourceMutationResponse, error) {
	user, err := r.UserService.ForContext(ctx)
	if err != nil {
		ForbiddenError(ctx)
		return nil, nil
	}

	acc, err := r.AccountService.GetByIdentityID(ctx, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, accounts.ErrInvalidArgument):
			InvalidArgument(ctx, err.Error())
			return nil, nil
		case errors.Is(err, accounts.ErrNotFound):
			return &generated.LinkFundingSourceMutationResponse{
				Code:    "422",
				Success: false,
				Message: "Create bank account failed: Account not found.",
			}, nil
		default:
			InternalServerError(ctx)
			return nil, nil
		}
	}

	bankAccount, err := r.Fs.CreateBankAccount(ctx, &fundingsources.CreateBankAccountArgs{
		IdentityID:    user.ID,
		AccountID:     acc.ID,
		Name:          input.Name,
		AccountNumber: input.AccountNumber,
		RoutingNumber: input.RoutingNumber,
		Institution:   input.Institution,
		Type:          input.Type,
	})
	if err != nil {
		switch err.(type) {
		case *fundingsources.ErrInvalidArgument:
			InvalidArgument(ctx, err.Error())
			return nil, nil
		default:
			InternalServerError(ctx)
			return nil, nil
		}
	}

	return &generated.LinkFundingSourceMutationResponse{
		Code:    "200",
		Success: true,
		Message: "Linked account.",
		FundingSource: &generated.FundingSource{
			ID:                 bankAccount.ID,
			Name:               bankAccount.Name,
			VerificationStatus: bankAccount.VerificationState,
			Mask:               bankAccount.Mask,
			Type:               bankAccount.Type,
			SubType:            bankAccount.SubType,
		},
	}, nil
}

func (r *mutationResolver) VerifyUsdBankAccount(ctx context.Context, input generated.VerifyUsdBankAccountInput) (*generated.VerifyUsdBankAccountMutationResponse, error) {
	user, err := r.UserService.ForContext(ctx)
	if err != nil {
		ForbiddenError(ctx)
		return nil, nil
	}

	fundingSource, err := r.Fs.Verify(ctx, &fundingsources.VerifyArgs{
		IdentityID:      user.ID,
		FundingSourceID: input.FundingSourceID,
	})
	if err != nil {
		switch err.(type) {
		case *fundingsources.ErrInvalidArgument:
			InvalidArgument(ctx, err.Error())
			return nil, nil
		default:
			InternalServerError(ctx)
			return nil, nil
		}
	}

	return &generated.VerifyUsdBankAccountMutationResponse{
		Code:    "200",
		Success: true,
		Message: "Verified account.",
		FundingSource: &generated.FundingSource{
			ID:                 fundingSource.ID,
			Name:               fundingSource.Name,
			VerificationStatus: fundingSource.VerificationState,
			Mask:               fundingSource.Mask,
			Type:               fundingSource.Type,
			SubType:            fundingSource.SubType,
		},
	}, nil
}

func (r *mutationResolver) InitiateDeposit(ctx context.Context, input generated.DepositInput) (*generated.DepositMutationResponse, error) {
	user, err := r.UserService.ForContext(ctx)
	if err != nil {
		ForbiddenError(ctx)
		return nil, nil
	}

	parsedAmount, err := strconv.ParseUint(input.Amount, 10, 64)
	if err != nil {
		InvalidArgument(ctx, err.Error())
		return nil, nil
	}

	acc, err := r.AccountService.GetByIdentityID(ctx, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, accounts.ErrInvalidArgument):
			InvalidArgument(ctx, err.Error())
			return nil, nil
		case errors.Is(err, accounts.ErrNotFound):
			return &generated.DepositMutationResponse{
				Code:    "404",
				Success: false,
				Message: "Deposit failed: Account not found.",
			}, nil
		default:
			InternalServerError(ctx)
			return nil, nil
		}
	}

	depositTransaction, err := r.Ds.InitiateDeposit(ctx, &deposits.InitiateDepositArgs{
		IdentityID:      user.ID,
		AccountID:       acc.ID,
		FundingSourceID: input.FundingSourceID,
		Amount:          parsedAmount,
	})
	if err != nil {
		switch err.(type) {
		case *deposits.ErrInvalidArgument:
			InvalidArgument(ctx, err.Error())
			return nil, nil
		case *deposits.ErrUnverifiedFundingSource:
			return &generated.DepositMutationResponse{
				Code:    "403",
				Success: false,
				Message: "Deposit failed: Funding source is not verified.",
			}, nil
		default:
			InternalServerError(ctx)
			return nil, nil
		}
	}
	return &generated.DepositMutationResponse{
		Code:    "200",
		Success: true,
		Message: "Deposit initiated.",
		Transaction: &generated.Transaction{
			ID:          depositTransaction.ID,
			Timestamp:   depositTransaction.CreatedAt, // TODO: decide format
			Amount:      strconv.FormatInt(depositTransaction.NetAmount, 10),
			Type:        generated.TransactionTypeDeposit,
			Status:      depositTransaction.State,
			Description: depositTransaction.Description,
		},
	}, nil
}

func (r *mutationResolver) InitiateWithdrawal(ctx context.Context, input generated.WithdrawalInput) (*generated.WithdrawalMutationResponse, error) {
	user, err := r.UserService.ForContext(ctx)
	if err != nil {
		ForbiddenError(ctx)
		return nil, nil
	}

	parsedAmount, err := strconv.ParseUint(input.Amount, 10, 64)
	if err != nil {
		InvalidArgument(ctx, err.Error())
		return nil, nil
	}
	acc, err := r.AccountService.GetByIdentityID(ctx, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, accounts.ErrInvalidArgument):
			InvalidArgument(ctx, err.Error())
			return nil, nil
		case errors.Is(err, accounts.ErrNotFound):
			return &generated.WithdrawalMutationResponse{
				Code:    "404",
				Success: false,
				Message: "Deposit failed: Account not found.",
			}, nil
		default:
			InternalServerError(ctx)
			return nil, nil
		}
	}

	// TODO: determine provider to use
	withdrawal, err := r.Ws.InitiateWithdrawal(ctx, &withdrawals.InitiateWithdrawalArgs{
		IdentityID:      user.ID,
		AccountID:       acc.ID,
		FundingSourceID: input.FundingSourceID,
		Amount:          parsedAmount,
	})
	if err != nil {
		switch {
		case errors.Is(err, withdrawals.ErrInvalidArgument):
			InvalidArgument(ctx, err.Error())
			return nil, nil
		case errors.Is(err, withdrawals.ErrUnverifiedFundingSource):
			return &generated.WithdrawalMutationResponse{
				Code:    "403",
				Success: false,
				Message: "Withdrawal failed: Destination is unverified.",
			}, nil
		case errors.Is(err, withdrawals.ErrInsufficientBalance):
			return &generated.WithdrawalMutationResponse{
				Code:    "500",
				Success: false,
				Message: "Withdrawal failed: Insufficient balance.",
			}, nil
		default:
			InternalServerError(ctx)
			return nil, nil
		}
	}

	return &generated.WithdrawalMutationResponse{
		Code:    "200",
		Success: true,
		Message: "Withdrawal initiated.",
		Transaction: &generated.Transaction{
			ID:          withdrawal.ID,
			Timestamp:   withdrawal.CreatedAt, // TODO: decide format
			Amount:      strconv.FormatInt(withdrawal.NetAmount, 10),
			Type:        generated.TransactionTypeWithdrawal,
			Status:      withdrawal.State,
			Description: withdrawal.Description,
		},
	}, nil
}

func (r *mutationResolver) InitiateOutgoingPayment(ctx context.Context, input generated.OutgoingPaymentInput) (*generated.OutgoingPaymentMutationResponse, error) {
	user, err := r.UserService.ForContext(ctx)
	if err != nil {
		ForbiddenError(ctx)
		return nil, nil
	}

	parsedAmount, err := strconv.ParseUint(input.Amount, 10, 64)
	if err != nil {
		InvalidArgument(ctx, err.Error())
		return nil, nil
	}

	acc, err := r.AccountService.GetByIdentityID(ctx, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, accounts.ErrInvalidArgument):
			InvalidArgument(ctx, err.Error())
			return nil, nil
		default:
			InternalServerError(ctx)
			return nil, nil
		}
	}
	outgoingPayment, err := r.Ps.InitiateOutgoingPayment(ctx, &payments.InitiateOutgoingPaymentArgs{
		IdentityID: user.ID,
		AccountID:  acc.ID,
		Amount:     parsedAmount,
		To:         input.To,
	})
	if err != nil {
		switch {
		case errors.Is(err, payments.ErrInvalidArgument):
			InvalidArgument(ctx, err.Error())
			return nil, nil
		case errors.Is(err, payments.ErrInsufficientBalance):
			return &generated.OutgoingPaymentMutationResponse{
				Code:    "422",
				Success: false,
				Message: "Outgoing payment failed: Insufficient balance.",
			}, nil
		case errors.Is(err, payments.ErrUnverifiedAccount):
			return &generated.OutgoingPaymentMutationResponse{
				Code:    "403",
				Success: false,
				Message: "Outgoing payment failed: Account unverified.",
			}, nil
		default:
			InternalServerError(ctx)
			return nil, nil
		}
	}

	return &generated.OutgoingPaymentMutationResponse{
		Code:    "200",
		Success: true,
		Message: "Outgoing payment initiated.",
		Transaction: &generated.Transaction{
			ID:          outgoingPayment.ID,
			Timestamp:   outgoingPayment.CreatedAt, // TODO: decide format
			Amount:      strconv.FormatInt(outgoingPayment.NetAmount, 10),
			Type:        generated.TransactionTypeSent,
			Status:      outgoingPayment.State,
			Description: outgoingPayment.Description,
		},
	}, nil
}

func (r *queryResolver) Identity(ctx context.Context) (*_identity.Identity, error) {
	user, err := r.UserService.ForContext(ctx)
	if err != nil {
		ForbiddenError(ctx)
		return nil, nil
	}

	var identity *_identity.Identity
	err = crdbsqlx.ExecuteTx(ctx, r.Db, nil, func(tx *sqlx.Tx) error {
		id, err := r.IdentityService.Get(ctx, tx, user.ID)
		if err != nil {
			return err
		}

		identity = id
		return nil
	})
	if err != nil {
		switch err.(type) {
		case *_identity.ErrNotFound:
			NotFoundError(ctx)
			return nil, nil
		default:
			InternalServerError(ctx)
			return nil, nil
		}
	}

	return identity, nil
}

func (r *queryResolver) FundingSources(ctx context.Context) ([]*generated.FundingSource, error) {
	user, err := r.UserService.ForContext(ctx)
	if err != nil {
		ForbiddenError(ctx)
		return nil, nil
	}
	acc, err := r.AccountService.GetByIdentityID(ctx, user.ID)
	if err != nil {
		switch err.(type) {
		default:
			InternalServerError(ctx)
			return nil, nil
		}
	}

	var fundingSources []fundingsources.FundingSource
	err = crdbsqlx.ExecuteTx(ctx, r.Db, nil, func(tx *sqlx.Tx) error {
		_fundingSources, err := r.Fs.GetByAccountId(ctx, tx, acc.ID)
		if err != nil {
			return err
		}

		fundingSources = _fundingSources
		return nil
	})
	if err != nil {
		switch err.(type) {
		case *_identity.ErrNotFound:
			NotFoundError(ctx)
			return nil, nil
		default:
			InternalServerError(ctx)
			return nil, nil
		}
	}
	ret := make([]*generated.FundingSource, len(fundingSources))
	for i, trx := range fundingSources {
		ret[i] = &generated.FundingSource{
			ID:                 trx.ID,
			Name:               trx.Name,
			VerificationStatus: trx.VerificationState,
			Mask:               trx.Mask,
			Type:               trx.Type,
			SubType:            trx.SubType,
		}
	}

	return ret, nil
}

func (r *queryResolver) Account(ctx context.Context) (*generated.Account, error) {
	user, err := r.UserService.ForContext(ctx)
	if err != nil {
		ForbiddenError(ctx)
		return nil, nil
	}

	var account *generated.Account
	err = crdbsqlx.ExecuteTx(ctx, r.Db, nil, func(tx *sqlx.Tx) error {
		identity, err := r.IdentityService.Get(ctx, tx, user.ID)
		if err != nil {
			return err
		}

		acc, err := r.AccountService.GetByIdentityIDWithTrx(ctx, tx, identity.ID)
		if err != nil {
			return err
		}

		floatBalance := float64(acc.AvailableBalance) / float64(100)

		account = &generated.Account{
			ID:      acc.ID,
			Balance: fmt.Sprintf("$ %.2f", floatBalance),
		}
		return nil
	})
	if err != nil {
		return nil, nil
	}

	return account, nil
}

func (r *queryResolver) Countries(ctx context.Context) ([]*generated.Country, error) {
	var countries []*country.Country
	countries, err := r.CountryService.GetAll(ctx)
	if err != nil {
		InternalServerError(ctx)
		return nil, nil
	}
	ret := make([]*generated.Country, len(countries))
	for i, trx := range countries {
		ret[i] = &generated.Country{
			ID:   trx.Alpha_2,
			Name: trx.Name,
		}
	}

	return ret, err
}

// Account returns generated.AccountResolver implementation.
func (r *Resolver) Account() generated.AccountResolver { return &accountResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type accountResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
