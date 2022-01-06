
Install for go compilation
```shell
$ go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Add to path
```shell
export PATH="$PATH:$(go env GOPATH)/bin"
```
