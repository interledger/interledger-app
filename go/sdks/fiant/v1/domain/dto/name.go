package dto

// https://github.com/provenancetech/pti-platform-sdks/blob/master/java/src/main/java/com/pti/sdk/types/Name.java

/*
var name = NewName(WithFirstName("John"), WithLastName("Doe"), WithMiddleName("Smith"))
*/
type Name struct {
	First  string `json:"firstName,omitempty"`
	Last   string `json:"lastName,omitempty"`
	Middle string `json:"middleName,omitempty"`
}

type NameOption func(*Name)

func WithFirstName(first string) NameOption {
	return func(n *Name) {
		n.First = first
	}
}

func WithLastName(last string) NameOption {
	return func(n *Name) {
		n.Last = last
	}
}

func WithMiddleName(middle string) NameOption {
	return func(n *Name) {
		n.Middle = middle
	}
}

func NewName(opts ...NameOption) *Name {
	n := Name{}
	for _, opt := range opts {
		opt(&n)
	}
	return &n
}
