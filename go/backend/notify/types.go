package notify

type NotificationType string

const (
	NotificationTypeLinkedAccount = "linkedAccount"
	NotificationTypeTransaction   = "transaction"
	NotificationTypeKyc           = "kyc"
)
