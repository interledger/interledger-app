package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/country"
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
	var ongoing struct {
		ID     string                `db:"external_id"`
		Status persona.InquiryStatus `db:"state"`
	}
	err = b.DB().GetContext(ctx, &ongoing, "SELECT external_id, state FROM kyc_persona_inquiries WHERE wallet_id=$1 AND state IN ($2, $3, $4, $5)",
		walletID, persona.InquiryPending, persona.InquiryCreated, persona.InquiryStarted, persona.InquiryNeedsReview)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w %s", kyc.ErrInternal, err)
	}

	var inquiry *persona.InquiryData
	if ongoing.ID != "" && ongoing.Status == persona.InquiryNeedsReview {
		return &kyc.PersonaInquiry{
			ID:     ongoing.ID,
			Status: ongoing.Status,
		}, nil
	} else if ongoing.ID != "" {
		// The same idempotency key may have been used to create the initial Inquiry, so append the ID.
		if idempotencyKey != "" {
			idempotencyKey += "-" + ongoing.ID
		}
		inquiry, err = cl.ResumeInquiry(ctx, ongoing.ID, idempotencyKey)
		if err != nil {
			return nil, fmt.Errorf("%w %s", kyc.ErrInternal, err)
		}

		return &kyc.PersonaInquiry{
			ID:           inquiry.ID,
			SessionToken: inquiry.Meta.SessionToken,
			Status:       persona.InquiryStatus(inquiry.Attributes.Status),
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
		return nil, fmt.Errorf("%w %s", kyc.ErrInternal, err)
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

	// fallback to US
	if args.CountryCode == "" {
		args.CountryCode = "US"
	}

	args.InquiryTemplateID = string(persona.GetTemplateIDForCountry(ctx, country.Country(args.CountryCode)))
	inquiry, err = cl.CreateInquiry(ctx, args, idempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("%w %s", kyc.ErrInternal, err)
	}

	_, err = b.DB().ExecContext(ctx, "INSERT INTO kyc_persona_inquiries(external_id, wallet_id, state) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING;",
		inquiry.ID, walletID, inquiry.Attributes.Status)
	if err != nil {
		return nil, fmt.Errorf("%w %s", kyc.ErrInternal, err)
	}

	return &kyc.PersonaInquiry{
		ID:     inquiry.ID,
		Status: persona.InquiryStatus(inquiry.Attributes.Status),
	}, nil
}

func GetPersonaIDNumbers(ctx context.Context, b Backends, cl persona.Client, walletID string) (*kyc.PersonaIDNumbers, error) {
	// Lookup the latest "approved" inquiry.
	var inqID string
	err := b.DB().GetContext(ctx, &inqID, "SELECT external_id FROM kyc_persona_inquiries WHERE wallet_id=$1 AND state=$2 ORDER BY updated_at DESC ",
		walletID, "approved")
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w %s", kyc.ErrNoKYCInfo, err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", kyc.ErrInternal, err)
	}

	inq, err := cl.LookupInquiry(ctx, inqID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", kyc.ErrInternal, err)
	}

	resp := &kyc.PersonaIDNumbers{
		SocialSecurity: inq.Data.Attributes.SocialSecurityNumber,
	}

	for _, ii := range inq.Included {
		if ii.Type != "verification/government-id" {
			continue
		}

		resp.IdentificationNumber = ii.Attributes.IDNumber
		resp.IssuingCountry = ii.Attributes.CountryCode
		resp.IdentificationClass = ii.Attributes.IDClass
		resp.IssuingState = ii.Attributes.AddressSubdivision

		exDate, err := time.Parse("2006-01-02", ii.Attributes.ExpirationDate)
		if err == nil {
			resp.ExpirationDate = exDate
		}
	}

	return resp, nil
}
