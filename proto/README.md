### Install for compilation

```shell
$ go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
```

Add to path

```shell
export PATH="$PATH:$(go env GOPATH)/bin"
```

Install https://docs.buf.build/installation

```shell
# Will generate backend/frontend generated files, and ensure that the frontend client is formatted appropriately.
make gen
```

### Google proto

We use some of googles API message definitions for rpc.

https://github.com/googleapis/googleapis.git
