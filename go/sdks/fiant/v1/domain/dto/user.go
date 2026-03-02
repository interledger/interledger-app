package dto

import (
	"encoding/json"

	"github.com/google/uuid"
)

// https://github.com/provenancetech/pti-platform-sdks/blob/master/java/src/main/java/com/pti/sdk/types/UserType.java
type UserTypeEnum string

func (ut UserTypeEnum) String() string {
	return string(ut)
}

const (
	PERSON UserTypeEnum = "PERSON"
	// BUSINESS   UserTypeEnum = "BUSINESS" // not currently supported by us, but may be added in the future
)

// https://github.com/provenancetech/pti-platform-sdks/blob/master/java/src/main/java/com/pti/sdk/types/UserStatus.java
type UserStatusEnum string

func (us UserStatusEnum) String() string {
	return string(us)
}

const (
	ACTIVE   UserStatusEnum = "ACTIVE"
	INACTIVE UserStatusEnum = "INACTIVE"
	BLOCKED  UserStatusEnum = "BLOCKED"
)

// UserPage is the response type for the ListAll method of the userHandler
// https://github.com/provenancetech/pti-platform-sdks/blob/master/java/src/main/java/com/pti/sdk/types/UserPage.java
type UserPage struct {
	// (stub) there are other fields in the response, but we only care about the content field for now
	Content []User `json:"content,omitempty"`
}

func (up UserPage) EnumerateUsers() []User {
	return up.Content
}

// https://github.com/provenancetech/pti-platform-sdks/blob/master/java/src/main/java/com/pti/sdk/types/User.java

/*
var user = NewUser(
	WithUserType(PERSON),
	WithUserStatus(ACTIVE),
	WithUserStatusReason(""),
	WithUserTags([]string{"tag1", "tag2"}),
	WithUserSourceOfFunds("source of funds"),
	WithUserCreationDate("2024-01-01T00:00:00Z"),
	WithUserPTIMetaData(map[string]any{"key": "value"}),
	WithUserClientMetaData(map[string]any{"key": "value"}),
	WithUserName(NewName(WithFirstName("John"), WithLastName("Doe"))),
	WithUserDateOfBirth("1990-01-01"),
	WithUserEmails(
		[]Email{
			NewEmail(WithEmailAddress("john.doe@home.com"), AsDefaultEmail(true)),
			NewEmail(WithEmailAddress("jane.doe@home.com"), AsDefaultEmail(false)),
		},
	),
	WithUserPhones([]Phone{NewPhone("123-456-7890", AsDefaultPhone(true))}),
)

*/

type User struct {
	ID string `json:"id,omitempty"`

	Type         UserTypeEnum   `json:"type,omitempty"`
	Status       UserStatusEnum `json:"status,omitempty"`
	StatusReason string         `json:"statusReason,omitempty"`

	Tags []string `json:"tags,omitempty"`

	SourceOfFunds    string `json:"sourceOfFunds,omitempty"`
	UserCreationDate string `json:"userCreationDate,omitempty"`

	Addresses []Address `json:"addresses,omitempty"`

	UserPTIMetaData    map[string]any `json:"userPtiMeta,omitempty"`
	UserClientMetaData map[string]any `json:"userClientMeta,omitempty"`

	Name        *Name   `json:"name,omitempty"`
	DateOfBirth string  `json:"dateOfBirth,omitempty"`
	Emails      []Email `json:"emails,omitempty"`
	Phones      []Phone `json:"phones,omitempty"`
}

// controller expects that all structs that are sent in the body of a request implement MarshalJSON and UnmarshalJSON
func (u User) MarshalJSON() ([]byte, error) {
	type alias User
	return json.Marshal(alias(u))
}

// controller expects that all structs that are sent in the body of a request implement MarshalJSON and UnmarshalJSON
func (u *User) UnmarshalJSON(data []byte) error {
	type alias User
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*u = User(a)
	return nil
}

type UserOption func(*User)

func WithNewUserUUID() UserOption {
	return func(u *User) {
		u.ID = uuid.NewString()
	}
}

func WithUserID(id string) UserOption {
	return func(u *User) {
		u.ID = id
	}
}

func WithUserType(userType UserTypeEnum) UserOption {
	return func(u *User) {
		u.Type = userType
	}
}

func WithUserStatus(userStatus UserStatusEnum) UserOption {
	return func(u *User) {
		u.Status = userStatus
	}
}

func WithUserStatusReason(statusReason string) UserOption {
	return func(u *User) {
		u.StatusReason = statusReason
	}
}

func WithUserTags(tags []string) UserOption {
	return func(u *User) {
		u.Tags = tags
	}
}

func WithUserSourceOfFunds(sourceOfFunds string) UserOption {
	return func(u *User) {
		u.SourceOfFunds = sourceOfFunds
	}
}

func WithUserCreationDate(creationDate string) UserOption {
	return func(u *User) {
		u.UserCreationDate = creationDate
	}
}

func WithUserPTIMetaData(ptiMetaData map[string]any) UserOption {
	return func(u *User) {
		u.UserPTIMetaData = ptiMetaData
	}
}

func WithUserClientMetaData(clientMetaData map[string]any) UserOption {
	return func(u *User) {
		u.UserClientMetaData = clientMetaData
	}
}

func WithUserName(name Name) UserOption {
	return func(u *User) {
		u.Name = &name
	}
}

func WithUserDateOfBirth(dateOfBirth string) UserOption {
	return func(u *User) {
		u.DateOfBirth = dateOfBirth
	}
}

func WithUserEmails(emails []Email) UserOption {
	return func(u *User) {
		u.Emails = emails
	}
}

func WithUserPhones(phones []Phone) UserOption {
	return func(u *User) {
		u.Phones = phones
	}
}

func WithUserAddresses(addresses []Address) UserOption {
	return func(u *User) {
		u.Addresses = addresses
	}
}

func NewUser(opts ...UserOption) User {
	u := User{}
	for _, opt := range opts {
		opt(&u)
	}
	return u
}
