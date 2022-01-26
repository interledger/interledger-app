# cant use alpine as we need glibc for CGO
FROM golang:1.17 AS deps

WORKDIR /build

# Fetch dependencies.
COPY go.mod go.sum ./
RUN go mod download

FROM deps as dev

# make sure that libclient.so has been built by zig and is in tigerbeetle-go folder.
# doing it this way for now to avoid installing zig in this docker file.
COPY tigerbeetle_go ./tigerbeetle_go
COPY proto ./proto
COPY pacioli ./pacioli
ENTRYPOINT ["/build/backend/entrypoint.sh"]

FROM dev as builder

RUN go build -o main ./pacioli/main.go

FROM gcr.io/distroless/cc:nonroot
WORKDIR /dist

COPY --from=builder /build/main .
ENTRYPOINT ["/dist/main"]