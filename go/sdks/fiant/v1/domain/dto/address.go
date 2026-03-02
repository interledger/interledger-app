package dto

// https://github.com/provenancetech/pti-platform-sdks/blob/master/java/src/main/java/com/pti/sdk/types/Address.java

/*
var address = NewAddress(

	WithAddressStreet("123 Main St"),
	WithAddressCity("Anytown"),
	WithAddressPostalCode("12345"),
	WithAddressStateCode("CA"),
	WithAddressCountry("USA"),
	AsDefaultAddress(true),

)
*/
type Address struct {
	Street     string `json:"streetAddress,omitempty"`
	City       string `json:"city,omitempty"`
	PostalCode string `json:"postalCode,omitempty"`
	StateCode  string `json:"stateCode,omitempty"`
	Country    string `json:"country,omitempty"`
	IsDefault  bool   `json:"default,omitempty"`
}

type AddressOption func(*Address)

func WithAddressStreet(street string) AddressOption {
	return func(a *Address) {
		a.Street = street
	}
}

func WithAddressCity(city string) AddressOption {
	return func(a *Address) {
		a.City = city
	}
}

func WithAddressPostalCode(postalCode string) AddressOption {
	return func(a *Address) {
		a.PostalCode = postalCode
	}
}

func WithAddressStateCode(stateCode string) AddressOption {
	return func(a *Address) {
		a.StateCode = stateCode
	}
}

func WithAddressCountry(country string) AddressOption {
	return func(a *Address) {
		a.Country = country
	}
}

func AsDefaultAddress(isDefault bool) AddressOption {
	return func(a *Address) {
		a.IsDefault = isDefault
	}
}

func NewAddress(opts ...AddressOption) Address {
	a := Address{}
	for _, opt := range opts {
		opt(&a)
	}
	return a
}
