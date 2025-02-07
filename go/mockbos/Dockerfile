FROM public.ecr.aws/docker/library/golang:1.23.6 AS base

WORKDIR /build
ARG BUILDARCH

RUN wget -qO/bin/atlas https://release.ariga.io/atlas/atlas-linux-arm64-latest && \
	chmod +x /bin/atlas

RUN go install github.com/air-verse/air@latest

# Fetch dependencies.
COPY go.mod go.sum ./
RUN go mod download
COPY . ./

FROM base as builder

ENV ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target="/root/.cache/go-build" go build -o main -C mockbos
ENTRYPOINT ["/build/mockbos/main"]

FROM gcr.io/distroless/cc:nonroot
WORKDIR /dist

COPY --from=builder /bin/atlas /bin/atlas
COPY --from=builder /build/mockbos/main .
ENTRYPOINT ["/dist/main"]
