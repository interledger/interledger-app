package kubernetes

import (
	"crypto/sha1"
	"crypto/tls"
	"fmt"
	"log"
)

func FingerprintAddress(address string) string {
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:443", "oidc.eks.eu-west-1.amazonaws.com"), &tls.Config{})

	if err != nil {
		log.Println("Error in Dial", err)
		return ""
	}
	defer conn.Close()

	peerCerts := conn.ConnectionState().PeerCertificates
	cert := peerCerts[len(peerCerts)-1]

	fingerprint := sha1.Sum(cert.Raw)

	return fmt.Sprintf("%x", fingerprint)
}
