ARG NODE_IMAGE_PREFIX=""

# Install workspace dependencies needed to build/run protea (and its
# workspace dependency, @interledger/configa). Build context is the
# typescript/ workspace root (see local/protea.yaml, k8s/Makefile, CI).
FROM ${NODE_IMAGE_PREFIX}node:24-alpine AS deps
RUN apk add --no-cache python3 make g++ curl py3-setuptools
RUN corepack enable && corepack prepare pnpm@8.5.0 --activate

WORKDIR /app
COPY pnpm-workspace.yaml package.json pnpm-lock.yaml .npmrc ./
COPY packages/configa/package.json ./packages/configa/package.json
COPY protea/package.json ./protea/package.json

RUN pnpm install --frozen-lockfile --filter protea...

# @interledger/configa ships pre-built (dist/) — build it once here, shared
# by the dev and builder stages below.
FROM deps AS configa-build
COPY packages/configa ./packages/configa
RUN pnpm --filter @interledger/configa build

# Dev image: full workspace source; app/ is bind-mounted over this for live
# reload (see local/protea.yaml).
FROM ${NODE_IMAGE_PREFIX}node:24-alpine AS dev
RUN corepack enable && corepack prepare pnpm@8.5.0 --activate
WORKDIR /app
COPY --from=configa-build /app ./
COPY protea ./protea
WORKDIR /app/protea
CMD ["pnpm", "dev"]

# Build protea's production bundle
FROM ${NODE_IMAGE_PREFIX}node:24-alpine AS builder

ARG IMAGE_HASH
ARG BUILDARCH
ARG SENTRY_ORG
ARG SENTRY_PROJECT
ARG SENTRY_AUTH_TOKEN
ARG UPLOAD_SOURCEMAPS
ARG REMIX_PUBLIC_PATH
ARG CDN_BUCKET_NAME

RUN addgroup -g 1001 -S nodejs
RUN adduser -S remix -u 1001

RUN apk add --no-cache curl unzip
RUN curl -O https://downloads.rclone.org/rclone-current-linux-${BUILDARCH}.zip && unzip rclone-current-linux-${BUILDARCH}.zip && cd rclone-*-linux-${BUILDARCH} && cp rclone /usr/bin/rclone

WORKDIR /app
RUN chown remix:nodejs /app
USER remix
ENV COREPACK_HOME=/home/remix/.corepack
ENV PNPM_HOME=/home/remix/.local/bin
ENV PATH=${PNPM_HOME}:${PATH}
COPY --from=configa-build --chown=remix:nodejs /app ./
COPY --chown=remix:nodejs protea ./protea
RUN corepack prepare pnpm@8.5.0 --activate
RUN mkdir -p "${PNPM_HOME}" && printf '#!/bin/sh\nexec corepack pnpm "$@"\n' > "${PNPM_HOME}/pnpm" && chmod +x "${PNPM_HOME}/pnpm"

WORKDIR /app/protea
RUN pnpm build

ENV SENTRY_ORG=${SENTRY_ORG}
ENV SENTRY_PROJECT=${SENTRY_PROJECT}
ENV SENTRY_AUTH_TOKEN=${SENTRY_AUTH_TOKEN}

RUN [ -n "${SENTRY_PROJECT}" ] && [ "${UPLOAD_SOURCEMAPS}" = "true" ] && IMAGE_HASH=${IMAGE_HASH} pnpm upload-sourcemaps || echo "Skipping sourcemap upload: SENTRY_PROJECT is empty or UPLOAD_SOURCEMAPS is not 'true'"

RUN if [ -n "${REMIX_PUBLIC_PATH}" ]; then \
    rclone --config=rclone.conf -c --check-first --progress copy build/client cdn:${CDN_BUCKET_NAME}/protea/build/; \
    else \
    echo "Skipping CDN upload: REMIX_PUBLIC_PATH is not set"; \
    fi

# Production image. Preserves the workspace's relative layout
# (/app/{node_modules,packages,protea}) so protea's workspace-dependency
# symlinks (e.g. @interledger/configa) keep resolving.
FROM ${NODE_IMAGE_PREFIX}node:24-alpine AS runner
RUN addgroup -g 1001 -S nodejs
RUN adduser -S remix -u 1001

ARG SENTRY_RELEASE
# Same value as IMAGE_HASH / SENTRY_RELEASE in the publish workflow
# (i.e. the semantic-release tag, e.g. "v1.11.6"). Surfaced on the
# /healthz response as the `interledger-app-version` header so the
# Deploy workflow can poll until the new version is live.
ARG IMAGE_HASH

WORKDIR /app/protea
ENV COREPACK_HOME=/app/protea/.corepack
RUN mkdir -p /app/protea/.corepack && chown -R remix:nodejs /app && chmod 755 /app/protea/.corepack

# You only need to copy remix.config.js if you are NOT using the default configuration
# COPY --from=builder /app/protea/remix.config.js ./
COPY --from=builder --chown=remix:nodejs /app/protea/public ./public
COPY --from=builder --chown=remix:nodejs /app/protea/build ./build
COPY --from=builder --chown=remix:nodejs /app/protea/node_modules ./node_modules
COPY --from=builder --chown=remix:nodejs /app/protea/package.json ./package.json
COPY --from=builder --chown=remix:nodejs /app/protea/server.js ./server.js
COPY --from=builder --chown=remix:nodejs /app/protea/app/config.schema.mjs ./app/config.schema.mjs
COPY --from=builder --chown=remix:nodejs /app/node_modules /app/node_modules
COPY --from=builder --chown=remix:nodejs /app/packages /app/packages

USER remix
RUN corepack prepare pnpm@8.5.0 --activate
RUN find ./build -type f -name '*.map' -delete

ENV NODE_ENV=production
ENV SENTRY_RELEASE=${SENTRY_RELEASE}
# Exposed back out via the `interledger-app-version` header on /healthz.
# Left unset in non-production builds so dev images don't advertise a
# bogus version.
ENV INTERLEDGER_APP_VERSION=${IMAGE_HASH}

EXPOSE 3000

CMD ["corepack", "pnpm", "start"]
