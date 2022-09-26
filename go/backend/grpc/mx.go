package grpc

import (
	"context"

	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) GetBankAccountWidget(
	ctx context.Context,
	req *backendv1.GetBankAccountWidgetRequest,
) (*backendv1.GetBankAccountWidgetResponse, error) {
	return nil, nil
	//user, err := s.b.Users().UserForContext(ctx)
	//if err != nil {
	//	return nil, ForbiddenError("Unauthenticated.")
	//}
	//
	////FIXME: change to wallet
	//walletId := uuid.NewString()
	//url, err := s.b.MX().GetConnectWidget(ctx, walletId, user.ID)
	//if err != nil {
	//	return nil, InternalError("Unable to get widget.")
	//}
	//
	//return &backendv1.GetBankAccountWidgetResponse{
	//	Url: url,
	//}, nil
}

func (s *rpcService) AddBankAccount(
	ctx context.Context,
	req *backendv1.AddBankAccountRequest,
) (*backendv1.AddBankAccountResponse, error) {
	return nil, nil
	//user, err := s.b.Users().UserForContext(ctx)
	//if err != nil {
	//	return nil, ForbiddenError("Unauthenticated.")
	//}
	//
	////FIXME: change to wallet
	//walletId := uuid.NewString()
	//workflowUuid, err := s.b.MX().InitiateCreateAccount(ctx, mx.InitiateCreateAccountArgs{
	//	AccountID:  walletId,
	//	IdentityID: user.ID,
	//	UserGuid:   req.GetUserGuid(),
	//	MemberGuid: req.GetMemberGuid(),
	//})
	//if err != nil {
	//	return nil, InternalError("Unable to create bank account.")
	//}
	//
	//return &backendv1.AddBankAccountResponse{
	//	FundingsourceId: workflowUuid,
	//}, nil
}

func (s *rpcService) GetBankAccountDetails(
	ctx context.Context,
	req *backendv1.GetBankAccountDetailsRequest,
) (*backendv1.BankAccountDetails, error) {
	return nil, nil
	//_, err := s.b.Users().UserForContext(ctx)
	//if err != nil {
	//	return nil, ToGRPCError(err)
	//}
	//
	////FIXME: change to wallet
	//walletId := uuid.NewString()
	//
	//// wait till workflow has completed
	//err = s.b.MX().WaitForCreateAccount(ctx, req.GetFundingsourceId())
	//if err != nil {
	//	return nil, ToGRPCError(err)
	//}
	//
	//bankAccount, err := s.b.MX().GetAccountByFundingsource(ctx, req.GetFundingsourceId())
	//if errors.Is(err, mx.ErrNotFound) {
	//	return nil, ToGRPCError(err)
	//}
	//if bankAccount.AccountID != walletId {
	//	return nil, ForbiddenError("Unauthorized.")
	//}
	//
	//details, err := s.b.MX().ReadAccount(ctx, bankAccount.Guid)
	//if err != nil {
	//	return nil, ToGRPCError(err)
	//}
	//
	//maskStart := len(details.AccountNumber) - 4
	//if maskStart < 0 {
	//	maskStart = 0
	//}
	//return &backendv1.BankAccountDetails{
	//	FundingsourceId: req.GetFundingsourceId(),
	//	Type:            details.Type,
	//	Institution:     details.InstitutionCode,
	//	Mask:            details.AccountNumber[maskStart:],
	//}, nil
}

//type validateContinueAddingBankAccount struct {
//	Otp      string `validate:"required"`
//	Nickname string `validate:"required"`
//}
//
//func validateContinueAddingBankAccountDescription(err validator.FieldError) string {
//	switch err.Tag() {
//	case "required":
//		return "required."
//	}
//	return ""
//}

func (s *rpcService) ContinueAddingBankAccount(
	ctx context.Context,
	req *backendv1.ContinueAddingBankAccountRequest,
) (*backendv1.ContinueAddingBankAccountResponse, error) {
	return nil, nil
	//if err := s.b.Validator().Struct(&validateContinueAddingBankAccount{
	//	Otp:      req.GetOtp(),
	//	Nickname: req.GetNickName(),
	//}); err != nil {
	//	return nil, ValidationError(err, validateContinueAddingBankAccountDescription)
	//}
	//_, err := s.b.Users().UserForContext(ctx)
	//if err != nil {
	//	return nil, ForbiddenError("Unauthenticated.")
	//}
	//
	////FIXME: change to wallet
	//walletId := uuid.NewString()
	//
	//bankAccount, err := s.b.MX().GetAccountByFundingsource(ctx, req.GetFundingsourceId())
	//if err != nil {
	//	return nil, InternalError("Unable to get bank account details.")
	//}
	//if bankAccount.AccountID != walletId {
	//	return nil, ForbiddenError("Unauthorized.")
	//}
	//
	//err = s.b.MX().InitiateCreateFundingsource(ctx, mx.InitiateCreateFundingsourceArgs{
	//	AccountID:     walletId,
	//	Otp:           req.GetOtp(),
	//	Name:          req.GetNickName(),
	//	MxAccountGuid: bankAccount.Guid,
	//})
	//if err != nil {
	//	return nil, InternalError("Unable to create fundingsource.")
	//}
	//
	//return &backendv1.ContinueAddingBankAccountResponse{
	//	Success: true,
	//}, nil
}
