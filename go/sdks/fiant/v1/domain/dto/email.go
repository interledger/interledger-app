package dto

// https://github.com/provenancetech/pti-platform-sdks/blob/master/java/src/main/java/com/pti/sdk/types/Email.java

/*
var email = NewEmail(WithEmailAddress("example@example.com"), AsDefaultEmail(true))
*/
type Email struct {
	Address   string `json:"address,omitempty"`
	IsDefault bool   `json:"default,omitempty"`
}

type EmailOption func(*Email)

func AsDefaultEmail(isDefault bool) EmailOption {
	return func(e *Email) {
		e.IsDefault = isDefault
	}
}

func WithEmailAddress(address string) EmailOption {
	return func(e *Email) {
		e.Address = address
	}
}

func NewEmail(opts ...EmailOption) Email {
	e := Email{}
	for _, opt := range opts {
		opt(&e)
	}
	return e
}
