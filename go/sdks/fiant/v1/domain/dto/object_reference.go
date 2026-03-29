package dto

// https://github.com/provenancetech/pti-platform-sdks/blob/master/java/src/main/java/com/pti/sdk/types/ObjectReference.java
type ObjectReference struct {
	ID   string `json:"id,omitempty"`
	Link string `json:"link,omitempty"`
}
