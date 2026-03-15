# Interledger Wallet

## Local Development Environment

Go to [local/README.md](local/README.md) for the local development setup.

## Testing

### GoLang unittests

Unit tests currently require the database to run. Each test will create a new database to run against on the given instance
allowing tests to run in parralel.

To run the tests in your local environment, the easiest way is to first start up a Postgres instance on your local host, and then running the tests from the command line.

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
- In previous iterations the project used Postgres 15, so be on the lookout for issues relating to the move to Postgres 17
- After performing the steps above you should be able to run the tests directly from vscode
