package dto

// https://github.com/provenancetech/pti-platform-sdks/blob/master/java/src/main/java/com/pti/sdk/types/Phone.java

/*
var phone = NewPhone("123-456-7890", WithPhoneType("mobile"), AsDefaultPhone(true))
*/
type Phone struct {
	Number    string `json:"number,omitempty"`
	Type      string `json:"type,omitempty"`
	IsDefault bool   `json:"default,omitempty"`
}

type PhoneOption func(*Phone)

func WithPhoneType(phoneType string) PhoneOption {
	return func(p *Phone) {
		p.Type = phoneType
	}
}

func AsDefaultPhone(isDefault bool) PhoneOption {
	return func(p *Phone) {
		p.IsDefault = isDefault
	}
}

func NewPhone(number string, opts ...PhoneOption) Phone {
	p := Phone{Number: number}
	for _, opt := range opts {
		opt(&p)
	}
	return p
}
