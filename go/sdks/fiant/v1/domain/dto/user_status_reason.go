package dto

// https://github.com/provenancetech/pti-platform-sdks/blob/master/java/src/main/java/com/pti/sdk/types/UserStatusReason.java
type UserStatusReasonEnum string

func (usr UserStatusReasonEnum) String() string {
	return string(usr)
}

const (
	FRAUD_SUSPICION        UserStatusReasonEnum = "FRAUD_SUSPICION"
	BUSINESS_CO_OWNER      UserStatusReasonEnum = "BUSINESS_CO_OWNER"
	COMPLIANCE_FLAG        UserStatusReasonEnum = "COMPLIANCE_FLAG"
	HIGH_RISK_IP           UserStatusReasonEnum = "HIGH_RISK_IP"
	INFORMATION_MISMATCH   UserStatusReasonEnum = "INFORMATION_MISMATCH"
	HIGH_RISK_EMAIL_DOMAIN UserStatusReasonEnum = "HIGH_RISK_EMAIL_DOMAIN"
	UNUSUAL_HIGH_VELOCITY  UserStatusReasonEnum = "UNUSUAL_HIGH_VELOCITY"
	CHARGEBACK             UserStatusReasonEnum = "CHARGEBACK"
)
