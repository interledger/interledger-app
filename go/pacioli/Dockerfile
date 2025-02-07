# cant use alpine as we need glibc for CGO
FROM public.ecr.aws/docker/library/golang:1.23.6 AS base

WORKDIR /build

# Fetch grpc health prober until kubernetes 1.23 releases on EKS
RUN GRPC_HEALTH_PROBE_VERSION=v0.4.7 && \
	wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
	chmod +x /bin/grpc_health_probe

RUN wget -qO/bin/atlas https://release.ariga.io/atlas/atlas-linux-arm64-latest && \
	chmod +x /bin/atlas

# Fetch dependencies.
COPY go.mod go.sum ./
RUN go mod download

COPY proto ./proto
COPY pacioli ./pacioli
COPY log ./log
COPY tracing ./tracing
COPY env ./env
ENTRYPOINT ["/build/backend/entrypoint.sh"]

FROM base as builder

ENV ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target="/root/.cache/go-build" go build -o main ./pacioli/pacioli/main.go

FROM gcr.io/distroless/cc:nonroot

WORKDIR /dist

COPY --from=builder /bin/grpc_health_probe ./grpc_health_probe
COPY --from=builder /build/pacioli/db/schema.hcl /dist/pacioli/db/schema.hcl
COPY --from=builder /bin/atlas /bin/atlas
COPY --from=builder /build/main .

ENTRYPOINT ["/dist/main"]
