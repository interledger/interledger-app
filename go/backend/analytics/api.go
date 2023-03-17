package analytics

type Client interface {
	Identify(args IdentifyArgs)
	TrackUserSignup(userID string)
	TrackUserLogin(userID string)
	TrackUserLogout(userID string)
	GroupUserWallet(walletID, userID string)
	TrackWalletCreated(walletID, userID string)
	TrackWalletPaymentPointerCreated(walletID string)
	TrackWalletTransactionCreated(walletID string, args WalletTransactionArgs)
	TrackWalletTransactionCompleted(walletID string, args WalletTransactionArgs)
	TrackWalletTransactionFailed(walletID string, args WalletTransactionArgs)
	Close()
}
