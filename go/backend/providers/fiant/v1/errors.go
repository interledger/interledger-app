package v1

import "errors"

var (
	ErrMissingClientID = errors.New("fiant: missing client ID")

	ErrMissingPrivateKey         = errors.New("fiant: missing private key")
	ErrFailedToRemoveKid         = errors.New("fiant: failed to remove 'kid' from private key")
	ErrFailedToDerivePublicKey   = errors.New("fiant: failed to derive public key from private key")
	ErrFailedToComputeThumbprint = errors.New("fiant: failed to compute public key thumbprint")

	ErrSettingCidInJWSHeaders = errors.New("fiant: error setting 'cid' in JWS headers")
	ErrSettingKidInJWSHeaders = errors.New("fiant: error setting 'kid' in JWS headers")
	ErrComputingSignature     = errors.New("fiant: error computing signature")

	ErrMissingHTTPClient = errors.New("fiant: missing HTTP client")
)
