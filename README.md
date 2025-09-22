# Interledger Wallet

## Local Development Environment

Go to ```./local``` and follow [getting-started](./local/docs/getting-started.md).

### Wallet development
#### Running protea *locally*
Comment `protea.yaml` in the `docker-compose.yaml` so it would not start the frontend app as a container.
Create a `.env` in the `backend` directory starting from the `.env-example` and fill in the Gatehub credentials.

Use `pnpm dev` in order to start the frontend locally at `localhost:3000`.

#### Running protea as a *container*:
Generate new certificates using 
```
./generate-certs.sh
```
Add the generated cert in Chrome or your other browser as trusted cert.
Start the local development environment. 

## Legacy Development environment

It is also possible to start up the project using Nomad and Vagrant VMs. See [this](./legacy-dev.md) document for more information.

## Testing

### GoLang unittests

Unit tests currently require the database to run. Each test will create a new database to run against on the given instance
allowing tests to run in parralel.

To run the tests in your local environment, the easiest way is to first start up a postgres instance on your local host, and then running the tests from the command line.

```
# Starts a Postgres 17 instance
docker run -d --name ilf-postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=password -p 5432:5432 postgres:17

# Install Atlas CLI on your local host
curl -sSf https://atlasgo.sh | sh

# Use Atlas to generate test migrations
export DB_URL=postgres://postgres:password@127.0.0.1:5432?sslmode=disable
atlas migrate diff create_all \
    --dir "file://go/backend/db/testmigrations" \
    --to  "file://go/backend/db/schema.hcl" \
    --dev-url "${DB_URL}"
```

You will now be able to run specific test files like this
```
go test -count=4 -v go/backend/kyc/ops/persona_test.go
```

Some notes
- In previous iterations the project used postgres 15, so be on the lookout for issues relating to the move to pg 17
- After performing the steps above you should be able to run the tests directly from vscode


