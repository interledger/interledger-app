package vault

type Client interface {
	CreateKey(keyName string) error
	Sign(keyName string, input string) (string, error)
	Verify(keyName string, input VerifyInput) (bool, error)
}
