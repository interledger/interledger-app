This uses [Caddy](https://caddyserver.com) to deploy mock servers used to test Rafiki deployment.

**Mock GNAP server**

The mock server is discoverable in the cluster at http://mockauth and
will return a 200 OK for any request. The body contains a grant that
allows the read/create of incoming-payments and reading of outgoing-payments.
