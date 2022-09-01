package grpc

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/providers/unit"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) GetStatements(
	ctx context.Context,
	req *backendv1.GetStatementsRequest,
) (*backendv1.GetStatementsResponse, error) {
	user, err := s.userService.ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	unitCustomer, err := s.unitProvider.GetCustomerByIdentityID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, unit.ErrNotFound) {
			return nil, InternalError("Failed to find customer.")
		}
		return nil, InternalError("Get customer.")
	}

	statements, err := s.unitProvider.GetStatements(ctx, unitCustomer.ID)
	if err != nil {
		if errors.Is(err, unit.ErrNotFound) {
			return nil, NotFoundError("No statements found.")
		}
		return nil, InternalError("Get statements.")
	}

	var statementsOut []*backendv1.GetStatementsResponse_Statement
	for _, statement := range statements {
		statementsOut = append(statementsOut, &backendv1.GetStatementsResponse_Statement{
			Id:        statement.ID,
			Period:    statement.Period,
			AccountId: statement.AccountID,
		})
	}

	return &backendv1.GetStatementsResponse{
		Statements: statementsOut,
	}, nil
}

type validateGetStatementPDF struct {
	StatementID string `validate:"required"`
}

func validateGetStatementPDFDescription(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "Missing required fields."
	}
	return ""
}

func (s *rpcService) GetStatementPDF(
	ctx context.Context,
	req *backendv1.GetStatementPDFRequest,
) (*backendv1.GetStatementPDFResponse, error) {
	if err := s.validator.Struct(&validateGetStatementPDF{
		StatementID: req.GetStatementId(),
	}); err != nil {
		return nil, ValidationError(err, validateGetStatementPDFDescription)
	}

	user, err := s.userService.ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	customer, err := s.unitProvider.GetCustomerByIdentityID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, unit.ErrNotFound) {
			return nil, NotFoundError("Failed to find customer.")
		}
		return nil, InternalError("Get customer.")
	}

	statementPdf, err := s.unitProvider.GetStatementPDF(ctx, &unit.GetStatementPDFArgs{
		StatementID: req.GetStatementId(),
		CustomerID:  customer.ID,
	})
	if err != nil {
		if errors.Is(err, unit.ErrNotFound) {
			return nil, NotFoundError("Failed to find statement.")
		}
		return nil, InternalError("Get statement.")
	}

	return &backendv1.GetStatementPDFResponse{
		Id:           statementPdf.ID,
		StatementPdf: statementPdf.PDF,
	}, nil
}
