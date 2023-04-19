package vault

type Client interface {
	CreateKey(keyName string) error
	Sign(keyName string, input *SignInput) (string, error)
	Verify(keyName string, input *VerifyInput) (bool, error)
}
