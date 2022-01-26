# Pacioli
This provides a multi-tenant accounts and policy service. Metadata about accounts and transfers are stored in CockroachDB. This takes the form of account categories and transaction types. TigerBeetle is used to store the actual account and transfer data.

## TigerBeetle client
The tigerbeetle client code is found in `../tigerbeetle_go`. It only supports linux for the moment and the will be removed from the repo once the TigerBeetle team officially release the Go client.

This does mean that for the moment building the `pacioli` docker image requires first manually installing zig and building the C ABI for TigerBeetle.

```sh
# make sure you have zig installed https://github.com/coilhq/tigerbeetle/blob/main/scripts/install_zig.sh
cd ../tigerbeetle_go

git submodule update --remote
zig build-lib --main-pkg-path ./internal -dynamic -lc -ODebug internal/client_c/client_c.zig
```

## Local dev
Make sure you have built the TigerBeetle C ABI. Then run
```sh
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:../../tigerbeetle_go
go test ./...
```