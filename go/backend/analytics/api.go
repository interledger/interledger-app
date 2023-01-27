package analytics

type Client interface {
	Identify(args IdentifyArgs)
	TrackUserSignup(userID string)
	TrackUserLogin(userID string)
	TrackUserLogout(userID string)
	Close()
}
