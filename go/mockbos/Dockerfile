ARG GOLANG_VERSION=1.24

FROM public.ecr.aws/docker/library/golang:${GOLANG_VERSION} AS base

ARG ATLAS_VERSION=v0.31.0
ARG AIR_VERSE_VERSION=v1.52
ARG GRPC_HEALTH_PROBE_VERSION=v0.4.7
ARG BUILDARCH=amd64
ARG BUILDOS=linux

WORKDIR /build

RUN wget -qO/bin/atlas https://release.ariga.io/atlas/atlas-${BUILDOS}-${BUILDARCH}-${ATLAS_VERSION} && \
	chmod +x /bin/atlas

RUN go install github.com/air-verse/air@${AIR_VERSE_VERSION}

# Fetch dependencies.
COPY go.mod go.sum ./
RUN go mod download
COPY . ./

FROM base AS builder

ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target="/root/.cache/go-build" go build -C mockbos -o main
ENTRYPOINT ["/build/mockbos/main"]

FROM gcr.io/distroless/cc:nonroot
WORKDIR /dist

COPY --from=builder /bin/atlas /bin/atlas
COPY --from=builder /build/mockbos/main .
ENTRYPOINT ["/dist/main"]
