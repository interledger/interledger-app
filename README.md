# Interledger Wallet

## Local Development Environment

### Create local dev environment

We provide a `docker compose` managed environment that will attempt to automatically rebuild the sources upon any change. 

The development environment will serve the various services under the following URLs.

| URL                              | Description                                 |
|-----------------------------------|---------------------------------------------|
| http://interledger.test                  | Wallet frontend (previously called Protea)  |
| http://admin.mgnt.interledger.test       | Wallet admin portal (previously Botanist)   |
| http://temporal.mgnt.interledger.test    | Temporal admin portal                       |
| http://local.fynbos.me                   | Backend endpoint                            |
| http://local.ilp.link                    | Wallet address                              |
| http://traefik.test                      | Traefik dashboard                           |


Prerequisites:
- Docker locally installed

Start environment
1. Add host entries
Edit your `/etc/hosts` file with an appropriate text editor and add the following line
```
127.0.0.1 interledger.test admin.mgnt.interledger.test temporal.mgnt.interledger.test local.fynbos.me local.ilp.link rafiki.mgnt.interledger.test auth.interledger.test traefik.test
```

All the domains used for local development will now point to your local host from where it will be served.

2. Start all services using Docker Compose
```sh
# From repository folder

cd ./local
docker compose up -d
```

All services should start asyncronously and automatically rebuild sources whenever code changes have been made.

## Troubleshooting

### View log for a service
The example below demonstrates how to view the last 100 logs for the protea service.
```sh
docker compose logs -f --tail=100 protea
```

### Recreating environment
Removes the whole environment to start fresh
```sh
# Remove containers and volumes
docker compose down -v
```

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


