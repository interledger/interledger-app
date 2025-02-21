FROM public.ecr.aws/docker/library/golang:1.23.0 AS base

WORKDIR /build
ARG BUILDARCH=amd64

RUN wget -qO/bin/atlas https://release.ariga.io/atlas/atlas-linux-${BUILDARCH}-v0.28.0 && \
	chmod +x /bin/atlas

RUN go install github.com/air-verse/air@latest

# Fetch dependencies.
COPY go.mod go.sum ./
RUN go mod download
COPY . ./

FROM base AS builder

ENV ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target="/root/.cache/go-build" go build -o main -C mockbos
ENTRYPOINT ["/build/mockbos/main"]

FROM gcr.io/distroless/cc:nonroot
WORKDIR /dist

COPY --from=builder /bin/atlas /bin/atlas
COPY --from=builder /build/mockbos/main .
ENTRYPOINT ["/dist/main"]
