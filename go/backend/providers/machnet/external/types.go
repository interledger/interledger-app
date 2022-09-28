package external

const (
	StatusVerified   = "VERIFIED"
	StatusUnverified = "UNVERIFIED"
)

type User struct {
	ID                string             `json:"id,omitempty"`
	FirstName         string             `json:"first_name,omitempty"`
	MiddleName        string             `json:"middle_name,omitempty"`
	LastName          string             `json:"last_name,omitempty"`
	Email             string             `json:"email,omitempty"`
	Gender            string             `json:"gender,omitempty"`
	DateOfBirth       string             `json:"date_of_birth,omitempty"`
	AddressLine1      string             `json:"address_line1,omitempty"`
	AddressLine2      string             `json:"address_line2,omitempty"`
	MobilePhone       string             `json:"mobile_phone,omitempty"`
	City              string             `json:"city,omitempty"`
	Zipcode           string             `json:"zipcode,omitempty"`
	State             string             `json:"state,omitempty"`
	Country           string             `json:"country,omitempty"`
	IPAddress         string             `json:"ip_address,omitempty"`
	Type              string             `json:"type,omitempty"`
	PhysicalDocuments []PhysicalDocument `json:"physical_documents,omitempty"`
	Status            string             `json:"status,omitempty"`
	Business          bool               `json:"business,omitempty"`
	BusinessType      string             `json:"business_type,omitempty"`
}

type PhysicalDocument struct {
	DocumentValue     string `json:"document_value,omitempty"`
	DocumentValueBack string `json:"document_value_back,omitempty"`
	DocumentType      string `json:"document_type,omitempty"`
	DocumentCountry   string `json:"country,omitempty"`
	DocumentState     string `json:"state,omitempty"`
}
