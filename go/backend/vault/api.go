package vault

type Client interface {
	CreateKey(keyName string) error
	GetPublicKey(keyName string) (string, error)
	Sign(keyName string, input string) ([]byte, error)
	Verify(keyName string, input VerifyInput) (bool, error)
}
