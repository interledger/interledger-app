package dto

// https://github.com/provenancetech/pti-platform-sdks/blob/master/java/src/main/java/com/pti/sdk/types/UserAssessStatus.java
type UserAssessStatusEnum string

func (uas UserAssessStatusEnum) String() string {
	return string(uas)
}

const (
	PENDING                    UserAssessStatusEnum = "PENDING"
	ERROR                      UserAssessStatusEnum = "ERROR"
	UNDER_REVIEW               UserAssessStatusEnum = "UNDER_REVIEW"
	REQUESTED_MORE_INFORMATION UserAssessStatusEnum = "REQUESTED_MORE_INFORMATION"
	ACCEPTED                   UserAssessStatusEnum = "ACCEPTED"
	REFUSED                    UserAssessStatusEnum = "REFUSED"
	INVALID                    UserAssessStatusEnum = "INVALID"
)

// https://github.com/provenancetech/pti-platform-sdks/blob/master/java/src/main/java/com/pti/sdk/types/UserAssessStatusObject.java
type UserAssessment struct {
	ResourceType  string               `json:"resourceType"`
	ClientID      string               `json:"clientId"`
	RequestID     string               `json:"requestId"`
	UserId        string               `json:"userId"`
	Date          string               `json:"date"`
	Assessment    UserAssessStatusEnum `json:"assessment"`
	Tier          int                  `json:"tier"`
	RefusalReason string               `json:"refusalReason"`
}
