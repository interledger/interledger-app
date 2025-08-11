# Interledger Wallet

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

## Vagrant based dev environment

The text below explains how to fire up the Vagrant and Nomad based development environment. This mechanism will become deprecated
to be replaced by kubernetes kind based environment that runs directly on the host machine. The vagrant environment requires too
much RAM for the average workstation and makes development inaccessible for most developers.

### Create local dev environment

This spins up a Nomad and Consul deployment in a VM using VirtualBox and Vagrant.
```shell
make devup
```

### Delete local dev environment

To delete the local environment run the following command
```shell
make devdown
```

### Rerunning deployment 

To deploy all the services or when deployment config has changed, run
```shell
make devdeploy
```

### SSH into dev environment

To ssh into the vagrant VM, run
```shell
make devssh
```

> [!WARNING]
> Sometimes the files on your host are not mounted into the VM. This oftwn appears as a file not found error.
> Rerun `make devdeploy`.

> [!WARNING]
> Sometimes the Nomad jobs might fail because of a placement error: "Constraint `${attr.consul.version} semver >= 1.8.0` filtered 1 node".
> Run `make devnomad` to fix this issue.
