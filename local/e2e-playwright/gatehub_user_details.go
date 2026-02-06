package main

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// UserDetails holds the details for a named user (e.g., 'kyc-user')
// Maps field names to values for later reference
type UserDetails struct {
	Fields map[string]string
}

// SignupContext needs to track current and stored user details
// These fields are added to the SignupContext struct definition

// theDetailsOfAre stores user details from a Gherkin table
// Usage: Given the details of 'kyc-user' are
//
//	| field           | value                        |
//	| emailSuffix     | alice@example.com            |
//	| password        | InterlEdger2025!TestPassword |
func (sc *SignupContext) theDetailsOfAre(userName string, table *godog.Table) error {
	// Initialize the map if it doesn't exist
	if sc.userDetails == nil {
		sc.userDetails = make(map[string]*UserDetails)
	}

	// Initialize field map for this user
	sc.userDetails[userName] = &UserDetails{
		Fields: make(map[string]string),
	}

	// Parse the table
	for _, row := range table.Rows {
		if len(row.Cells) < 2 {
			continue
		}
		fieldName := row.Cells[0].Value
		fieldValue := row.Cells[1].Value
		sc.userDetails[userName].Fields[fieldName] = fieldValue
	}

	debugPrintf("📝 Stored details for user '%s': %v\n", userName, sc.userDetails[userName].Fields)
	return nil
}

// iImpersonate sets up the current user context
// Usage: And I impersonate 'kyc-user'
func (sc *SignupContext) iImpersonate(userName string) error {
	details, exists := sc.userDetails[userName]
	if !exists {
		return fmt.Errorf("no details defined for user '%s'", userName)
	}

	// Store the current impersonated user
	sc.currentUser = userName

	// Extract and store standard fields from the details
	if email, ok := details.Fields["emailSuffix"]; ok {
		// Prefix email with random test identifier
		prefixedEmail := fmt.Sprintf("%s-%s", sc.testEmailPrefix, email)
		// Store in user details, not global sc.email
		details.Fields["email"] = prefixedEmail
		debugPrintf("✓ Stored email in userDetails: %s\n", prefixedEmail)
	}

	if password, ok := details.Fields["password"]; ok {
		sc.password = password
		debugPrintf("✓ Set password\n")
	}

	if firstName, ok := details.Fields["firstName"]; ok {
		sc.firstName = firstName
		debugPrintf("✓ Set firstName to: %s\n", firstName)
	}

	if lastName, ok := details.Fields["lastName"]; ok {
		sc.lastName = lastName
		debugPrintf("✓ Set lastName to: %s\n", lastName)
	}

	if country, ok := details.Fields["country"]; ok {
		sc.country = country
		debugPrintf("✓ Set country to: %s\n", country)
	}

	if countryCode, ok := details.Fields["countryCode"]; ok {
		sc.countryCode = countryCode
		debugPrintf("✓ Set countryCode to: %s\n", countryCode)
	}

	if dateOfBirth, ok := details.Fields["dateOfBirth"]; ok {
		sc.dateOfBirth = dateOfBirth
		debugPrintf("✓ Set dateOfBirth to: %s\n", dateOfBirth)
	}

	debugPrintf("✓ Impersonated user '%s'\n", userName)
	return nil
}

// getFieldValue retrieves a field value from the current impersonated user
func (sc *SignupContext) getFieldValue(fieldKey string) (string, error) {
	if sc.currentUser == "" {
		return "", fmt.Errorf("no user currently impersonated")
	}

	details, exists := sc.userDetails[sc.currentUser]
	if !exists {
		return "", fmt.Errorf("no details found for user '%s'", sc.currentUser)
	}

	value, ok := details.Fields[fieldKey]
	if !ok {
		return "", fmt.Errorf("field '%s' not found in user '%s' details. Available fields: %v", fieldKey, sc.currentUser, getMapKeys(details.Fields))
	}

	return value, nil
}

// getMapKeys returns the keys of a string map (for debugging)
func getMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// iFillInWithMy fills a form field with a value from the current user's details
// Usage: I fill in "first name" with my "firstName"
func (sc *SignupContext) iFillInWithMy(fieldName, fieldKey string) error {
	value, err := sc.getFieldValue(fieldKey)
	if err != nil {
		return err
	}

	debugPrintf("📝 Filling '%s' with '%s' (from field '%s')\n", fieldName, value, fieldKey)
	return sc.iFillInWith(fieldName, value)
}

// iFillInWithPrefixed fills a form field with a value prefixed with the random test identifier
// Usage: I fill in "email" with "emailSuffix" prefixed with the random identifier
func (sc *SignupContext) iFillInWithPrefixed(fieldName, fieldKey string) error {
	value, err := sc.getFieldValue(fieldKey)
	if err != nil {
		return err
	}

	// Prefix with test identifier
	prefixedValue := fmt.Sprintf("%s-%s", sc.testEmailPrefix, value)
	debugPrintf("📝 Filling '%s' with '%s' (prefixed, from field '%s')\n", fieldName, prefixedValue, fieldKey)

	// Store the prefixed email in context
	if strings.ToLower(fieldName) == "email" {
		sc.email = prefixedValue
	}

	return sc.iFillInWith(fieldName, prefixedValue)
}

// iTryToFillInWithMy tries to fill a form field with a value from the current user's details
// Doesn't fail if element doesn't exist
// Usage: I try to fill in "password" with my "password"
func (sc *SignupContext) iTryToFillInWithMy(fieldName, fieldKey string) error {
	value, err := sc.getFieldValue(fieldKey)
	if err != nil {
		return err
	}

	debugPrintf("📝 Trying to fill '%s' with '%s' (from field '%s')\n", fieldName, value, fieldKey)
	return sc.iTryToFillInWith(fieldName, value)
}

// getCurrentUserEmail retrieves the email for the current user
// Fails fast if user not set or email not stored
func (sc *SignupContext) getCurrentUserEmail() (string, error) {
	if sc.currentUser == "" {
		return "", fmt.Errorf("no current user set - cannot retrieve email")
	}
	details, exists := sc.userDetails[sc.currentUser]
	if !exists {
		return "", fmt.Errorf("no user details for '%s'", sc.currentUser)
	}
	email, exists := details.Fields["email"]
	if !exists {
		return "", fmt.Errorf("no email stored for user '%s'", sc.currentUser)
	}
	return email, nil
}

// getCurrentUserPhone retrieves the phone for the current user
// Fails fast if user not set or phone not stored
func (sc *SignupContext) getCurrentUserPhone() (string, error) {
	if sc.currentUser == "" {
		return "", fmt.Errorf("no current user set - cannot retrieve phone")
	}
	details, exists := sc.userDetails[sc.currentUser]
	if !exists {
		return "", fmt.Errorf("no user details for '%s'", sc.currentUser)
	}
	phone, exists := details.Fields["phone"]
	if !exists {
		return "", fmt.Errorf("no phone stored for user '%s'", sc.currentUser)
	}
	return phone, nil
}

// getCurrentUserTOTPSecret retrieves the TOTP secret for the current user
// Fails fast if user not set or TOTP secret not stored
func (sc *SignupContext) getCurrentUserTOTPSecret() (string, error) {
	if sc.currentUser == "" {
		return "", fmt.Errorf("no current user set - cannot retrieve TOTP secret")
	}
	details, exists := sc.userDetails[sc.currentUser]
	if !exists {
		return "", fmt.Errorf("no user details for '%s'", sc.currentUser)
	}
	totpSecret, exists := details.Fields["totp_secret"]
	if !exists {
		return "", fmt.Errorf("no TOTP secret stored for user '%s'", sc.currentUser)
	}
	return totpSecret, nil
}
