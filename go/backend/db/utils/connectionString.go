package utils

import (
	"io/ioutil"
	"net/url"
)

// We read the ssl certs into memory and add them inline to the connection string.
// This side steps the file permission issues (postgres spec requires
// private key to have 0600) when mounting secrets into the pod and running
// as non-root user.
// https://github.com/hashicorp/vault/issues/10925
// https://github.com/lib/pq/commit/b9bb726ebf154627a21b50f9ffa4b28c6ed3f4d8
func InlineSslCreds(baseUrl string, pathToPrivateKey string, pathToClientCert string, pathToRootCert string) (string, error) {
	connString := baseUrl + "&sslinline=true"
	sslkeyBytes, err := ioutil.ReadFile(pathToPrivateKey)
    if err != nil {
        return "", err
    }

    sslcertBytes, err := ioutil.ReadFile(pathToClientCert)
    if err != nil {
        return "", err
    }

    sslrootcertBytes, err := ioutil.ReadFile(pathToRootCert)
    if err != nil {
        return "", err
    }

    connString += "&sslinline=true"
    connString += "&sslkey=" + url.QueryEscape(string(sslkeyBytes))
    connString += "&sslcert=" + url.QueryEscape(string(sslcertBytes))
    connString += "&sslrootcert=" + url.QueryEscape(string(sslrootcertBytes))

    return connString, nil
}