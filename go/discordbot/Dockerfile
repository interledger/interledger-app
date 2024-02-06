FROM public.ecr.aws/docker/library/golang:1.20.1 AS base

WORKDIR /build

# Fetch dependencies.
COPY go.mod go.sum ./
RUN go mod download
COPY backend ./backend
COPY env ./env
COPY proto ./proto
COPY log ./log
COPY tracing ./tracing
COPY httpmessagesignatures ./httpmessagesignatures
COPY discordbot ./discordbot
COPY pacioli ./pacioli

FROM base as builder

ENV ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target="/root/.cache/go-build" go build -o main -C discordbot
ENTRYPOINT ["/build/discordbot/main"]

FROM gcr.io/distroless/cc:nonroot
WORKDIR /dist

COPY --from=builder /build/discordbot/main .
ENTRYPOINT ["/dist/main"]
