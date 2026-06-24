# Pacioli
This provides a multi-tenant accounts and policy service. Account and transfer
data, along with their metadata (account categories and transaction types), are
stored in CockroachDB.

The ledger logic lives in the `ledger/tigerroach` package, which implements an
accounts-and-transfers ledger API (modelled on TigerBeetle's double-entry
semantics) backed by CockroachDB.

## Local dev
Run the tests with:
```sh
go test ./...
```
