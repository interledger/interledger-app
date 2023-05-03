package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/kyc/persona"
)

func GetPersonaInquiry(ctx context.Context, b Backends, cl persona.Client, walletID, idempotencyKey string) (*kyc.PersonaInquiry, error) {
	// Check current KYC status fro the user.
	kycStatus, err := GetKYCStatus(ctx, b, walletID)
	if err != nil {
		return nil, err
	}

	if kycStatus == kyc.StatusApproved || kycStatus == kyc.StatusDenied {
		return nil, fmt.Errorf("%w KYC status (%d)", kyc.ErrKYCCompleted, kycStatus)
	}

	// Check if there is an ongoing inquiry for this wallet that we can resume
	var ongoingID string
	err = b.DB().GetContext(ctx, &ongoingID, "SELECT external_id FROM kyc_persona_inquiries WHERE wallet_id=$1 AND state IN ($2, $3, $4)",
		walletID, "pending", "created", "started")
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w %s", kyc.ErrInternal, err)
	}

	var inquiry *persona.InquiryData
	if ongoingID != "" {
		// The same idempotency key may have been used to create the initial Inquiry, so append the ID.
		if idempotencyKey != "" {
			idempotencyKey += "-" + ongoingID
		}
		inquiry, err = cl.ResumeInquiry(ctx, ongoingID, idempotencyKey)
		if err != nil {
			return nil, fmt.Errorf("%w %s", kyc.ErrInternal, err)
		}

		return &kyc.PersonaInquiry{
			ID:           inquiry.ID,
			SessionToken: inquiry.Meta.SessionToken,
		}, nil
	}

	// Fill in some data if we have it for a new Inquiry
	var id *kyc.IndividualDetails
	id, err = GetIndividualDetails(ctx, b, walletID)
	if err != nil && !errors.Is(err, kyc.ErrNoKYCInfo) {
		return nil, err
	}

	ul, err := b.Users().ListUsers(ctx, walletID)
	if err != nil {
		return nil, err
	}

	args := persona.IndividualAttributes{
		ReferenceID:  walletID,
		EmailAddress: ul[0].Email,
		PhoneNumber:  ul[0].PhoneNumber,
	}
	if id != nil {
		args = persona.IndividualAttributes{
			ReferenceID:  walletID,
			NameFirst:    id.FirstName,
			NameLast:     id.LastName,
			EmailAddress: ul[0].Email,
			PhoneNumber:  ul[0].PhoneNumber,
			CountryCode:  id.CountryCode,
			Birthdate:    id.DateOfBirth.Format("2006-01-02"),
		}
	}

	inquiry, err = cl.CreateInquiry(ctx, args, idempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("%w %s", kyc.ErrInternal, err)
	}

	_, err = b.DB().ExecContext(ctx, "INSERT INTO kyc_persona_inquiries(external_id, wallet_id, state) VALUES ($1, $2, $3)",
		inquiry.ID, walletID, inquiry.Attributes.Status)
	if err != nil {
		return nil, fmt.Errorf("%w %s", kyc.ErrInternal, err)
	}

	return &kyc.PersonaInquiry{
		ID: inquiry.ID,
	}, nil
}

func GetPersonaIDNumbers(ctx context.Context, b Backends, cl persona.Client, walletID string) ([]kyc.PersonaIDNumber, error) {
	var accID string
	err := b.DB().GetContext(ctx, &accID, "SELECT external_id FROM kyc_persona_accounts WHERE wallet_id=$1", walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w %s", kyc.ErrNoKYCInfo, err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", kyc.ErrInternal, err)
	}

	acc, err := cl.GetAccount(ctx, accID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", kyc.ErrInternal, err)
	}

	var resp []kyc.PersonaIDNumber
	for _, idNum := range acc.Attributes.IdentificationNumbers {
		resp = append(resp, kyc.PersonaIDNumber{
			IssuingCountry:       idNum.IssuingCountry,
			IdentificationClass:  idNum.IdentificationClass,
			IdentificationNumber: idNum.IdentificationNumber,
		})
	}

	return resp, nil
}
