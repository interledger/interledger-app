package vault

type SignInput struct {
	Input string
}

type VerifyInput struct {
	Input     string
	Signature string
}
