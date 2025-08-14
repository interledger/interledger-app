package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"time"

	"gitlab.com/fynbos/env"

	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/kyc/persona"
)

func GetPersonaInquiry(ctx context.Context, b Backends, cl persona.Client, walletID, idempotencyKey string) (*kyc.PersonaInquiry, error) {
	if env.IsLocal() {
		err := GenerateKycData(ctx, b, walletID)
		return &kyc.PersonaInquiry{
			ID:     "local-inquiry-id",
			Status: persona.InquiryStatus(persona.InquiryApproved),
		}, err
	}

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

	if ongoing.ID != "" {
		return &kyc.PersonaInquiry{
			ID:     ongoing.ID,
			Status: ongoing.Status,
		}, nil
	}

	wallet, err := b.Wallets().Get(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", kyc.ErrInternal, err)
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
		CountryCode:  wallet.Country.String(),
	}
	if id != nil {
		args = persona.IndividualAttributes{
			ReferenceID:  walletID,
			NameFirst:    id.FirstName,
			NameLast:     id.LastName,
			EmailAddress: ul[0].Email,
			PhoneNumber:  ul[0].PhoneNumber,
			Birthdate:    id.DateOfBirth.Format("2006-01-02"),
		}
	}

	// fallback to US
	if args.CountryCode == "" {
		args.CountryCode = "US"
	}

	args.InquiryTemplateID = string(persona.GetTemplateIDForCountry(ctx, wallet.Country))
	inquiry, err := cl.CreateInquiry(ctx, args, idempotencyKey)
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

func GetApprovedPersonaInquiryURL(ctx context.Context, b Backends, walletID string) (string, error) {
	urlFormat := "https://app.withpersona.com/dashboard/inquiries/%s"

	if env.IsLocal() {
		return fmt.Sprintf(urlFormat, "local-inquiry-id"), nil
	}

	var inquiryID string
	err := b.DB().GetContext(ctx, &inquiryID, "SELECT external_id FROM kyc_persona_inquiries WHERE wallet_id=$1 AND state=$2;", walletID, persona.InquiryApproved)
	if errors.Is(err, sql.ErrNoRows) {
		return "", kyc.ErrNoKYCInfo
	}
	if err != nil {
		return "", fmt.Errorf("%w %s", kyc.ErrInternal, err)
	}

	return fmt.Sprintf(urlFormat, inquiryID), nil
}

func GetPersonaIDNumbers(ctx context.Context, b Backends, cl persona.Client, walletID string) (*kyc.PersonaIDNumbers, error) {
	if env.IsLocal() {
		return generateLocalIDNumbers(), nil
	}

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

func GetZAIDNumber(ctx context.Context, b Backends, cl persona.Client, walletID string) (string, error) {
	// Local stuff
	if env.IsLocal() {
		return generateLocalTestIdNumber("male", true), nil
	}

	var accID string
	err := b.DB().GetContext(ctx, &accID, "SELECT external_id FROM kyc_persona_accounts WHERE wallet_id=$1", walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w no persona account found", kyc.ErrNoKYCInfo)
	}
	if err != nil {
		return "", fmt.Errorf("%w %s", kyc.ErrInternal, err)
	}

	acc, err := cl.GetAccount(ctx, accID)
	if err != nil {
		return "", fmt.Errorf("%w %s", kyc.ErrInternal, err)
	}

	for _, v := range acc.Attributes.IdentificationNumbers {
		for _, idNum := range v {
			if idNum.IssuingCountry == "ZA" && isValidZAID(idNum.IdentificationNumber) {
				return idNum.IdentificationNumber, nil
			}
		}
	}

	return "", fmt.Errorf("%w no valid ZA ID number on persona account found", kyc.ErrNoKYCInfo)
}

func isValidZAID(idNumber string) bool {
	// Check if the ID number has the correct length
	if len(idNumber) != 13 {
		return false
	}

	// Check if the ID number consists only of numeric digits
	numericRegex := regexp.MustCompile("^[0-9]+$")
	if !numericRegex.MatchString(idNumber) {
		return false
	}

	// Extract the birthdate from the ID number
	year, _ := strconv.Atoi(idNumber[0:2])
	month, _ := strconv.Atoi(idNumber[2:4])
	day, _ := strconv.Atoi(idNumber[4:6])

	// Validate the birthdate
	if year < 0 || month < 1 || month > 12 || day < 1 || day > 31 {
		return false
	}

	// Validate the last four digits (the sequence number)
	sequenceNumber, _ := strconv.Atoi(idNumber[9:13])
	if sequenceNumber < 0 {
		return false
	}

	// Validate the citizenship indicator
	citizenshipIndicator, _ := strconv.Atoi(idNumber[10:11])
	if citizenshipIndicator < 0 || citizenshipIndicator > 8 {
		return false
	}

	// Check the checksum
	checksum, _ := strconv.Atoi(idNumber[12:13])

	return calculateZAIDChecksum(idNumber[0:12]) == checksum
}

func calculateZAIDChecksum(idBody string) int {
	weights := []int{1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2}
	sum := 0

	for i, digitStr := range idBody {
		digit, _ := strconv.Atoi(string(digitStr))
		weighted := digit * weights[i]

		if weighted > 9 {
			weighted -= 9
		}

		sum += weighted
	}

	return (10 - (sum % 10)) % 10
}

func generateLocalIDNumbers() *kyc.PersonaIDNumbers {
	return &kyc.PersonaIDNumbers{
		SocialSecurity:       "123456789",
		IssuingCountry:       "US",
		IssuingState:         "CA",
		IdentificationClass:  "1",
		IdentificationNumber: "123456789",
		ExpirationDate:       time.Now().Add(365 * 24 * time.Hour),
	}
}

func generateLocalTestIdNumber(gender string, isCitizen bool) string {
	// Set current year and calculate year range for 18-40 years old
	currentYear := time.Now().Year()
	minYear := currentYear - 40
	maxYear := currentYear - 18

	// Generate a random date of birth within the age range
	year := rand.Intn(maxYear-minYear+1) + minYear
	month := rand.Intn(12) + 1
	day := rand.Intn(28) + 1

	// Generate sequence number based on gender
	seqStart := 0000
	seqEnd := 9999
	if gender == "male" {
		seqStart = 5000
	} else if gender == "female" {
		seqEnd = 4999
	}
	sequence := rand.Intn(seqEnd-seqStart+1) + seqStart

	// Citizen status
	citizenDigit := 0
	if !isCitizen {
		citizenDigit = 1
	}

	// Combine all parts
	idNumber := fmt.Sprintf("%02d%02d%02d%04d%d8", year%100, month, day, sequence, citizenDigit)

	return idNumber
}
