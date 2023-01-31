package analytics

type Client interface {
	Identify(args IdentifyArgs)
	TrackUserSignup(userID string)
	TrackUserLogin(userID string)
	TrackUserLogout(userID string)
	GroupUserWallet(walletID, userID string)
	TrackWalletCreated(walletID, userID string)
	TrackWalletPaymentPointerCreated(walletID string)
	Close()
}
