#!/bin/bash

# Create certs directory if it doesn't exist
mkdir -p local/certs

# Generate private key
openssl genrsa -out local/certs/interledger.test.key 2048

# Create a config file for the certificate with SAN
cat > local/certs/cert.conf <<EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
CN = interledger.test

[v3_req]
keyUsage = critical, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = interledger.test
DNS.2 = *.interledger.test
DNS.3 = admin.mgnt.interledger.test
DNS.4 = temporal.mgnt.interledger.test
DNS.5 = rafiki.mgnt.interledger.test
DNS.6 = auth.interledger.test
DNS.7 = traefik.test
EOF

# Generate certificate with SAN
openssl req -new -x509 -key local/certs/interledger.test.key -out local/certs/interledger.test.crt -days 365 -config local/certs/cert.conf -extensions v3_req

# Clean up the config file
rm local/certs/cert.conf

echo "Generated SSL certificates for interledger.test with SAN support"
echo "Private key: local/certs/interledger.test.key"
echo "Certificate: local/certs/interledger.test.crt"
echo ""
echo "The certificate includes these domains:"
echo "  - interledger.test"
echo "  - *.interledger.test" 
echo "  - admin.mgnt.interledger.test"
echo "  - temporal.mgnt.interledger.test"
echo "  - rafiki.mgnt.interledger.test"
echo "  - auth.interledger.test"
echo "  - traefik.test"
