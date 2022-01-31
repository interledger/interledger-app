# cant use alpine as we need glibc for CGO
FROM golang:1.17 AS base

WORKDIR /build

# Fetch dependencies.
COPY go.mod go.sum ./
RUN go mod download

# make sure that libclient.so has been built by zig and is in tigerbeetle-go folder.
# doing it this way for now to avoid installing zig in this docker file.
COPY tigerbeetle_go ./tigerbeetle_go
COPY proto ./proto
COPY pacioli ./pacioli
ENTRYPOINT ["/build/backend/entrypoint.sh"]

FROM base as builder

RUN go build -o main ./pacioli/main.go

FROM gcr.io/distroless/cc:nonroot

WORKDIR /dist

COPY --from=builder /build/main .
# this is temporary until the official TB client is released.
COPY tigerbeetle_go/libclient_c.so .
ENV LD_LIBRARY_PATH=$LD_LIBRARY_PATH:./
ENTRYPOINT ["/dist/main"]