ARG NODE_IMAGE_PREFIX=""

# Install dependencies only when needed
FROM ${NODE_IMAGE_PREFIX}node:20-alpine AS deps
RUN apk add --no-cache python3 make g++ curl py3-setuptools
RUN corepack enable && corepack prepare pnpm@8.5.0 --activate

WORKDIR /app
COPY package.json pnpm-lock.yaml ./

RUN pnpm install --frozen-lockfile --shamefully-hoist

# Rebuild the source code only when needed
FROM ${NODE_IMAGE_PREFIX}node:20-alpine AS dev 
WORKDIR /app
COPY . .
COPY --from=deps /app/node_modules ./node_modules
RUN corepack enable && corepack prepare pnpm@8.5.0 --activate
CMD ["pnpm", "dev"]

# Rebuild the source code only when needed
FROM ${NODE_IMAGE_PREFIX}node:20-alpine AS builder

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
COPY --chown=remix:nodejs . .
COPY --from=deps --chown=remix:nodejs /app/node_modules ./node_modules
RUN corepack prepare pnpm@8.5.0 --activate
RUN mkdir -p "${PNPM_HOME}" && printf '#!/bin/sh\nexec corepack pnpm "$@"\n' > "${PNPM_HOME}/pnpm" && chmod +x "${PNPM_HOME}/pnpm"
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

# Production image, copy all the files and run remix
FROM ${NODE_IMAGE_PREFIX}node:20-alpine AS runner
RUN addgroup -g 1001 -S nodejs
RUN adduser -S remix -u 1001

ARG SENTRY_RELEASE

WORKDIR /app
ENV COREPACK_HOME=/app/.corepack
RUN mkdir -p /app/.corepack && chown -R remix:nodejs /app && chmod 755 /app/.corepack

# You only need to copy remix.config.js if you are NOT using the default configuration
# COPY --from=builder /app/remix.config.js ./
COPY --from=builder --chown=remix:nodejs /app/public ./public
COPY --from=builder --chown=remix:nodejs /app/build ./build
COPY --from=builder --chown=remix:nodejs /app/node_modules ./node_modules
COPY --from=builder --chown=remix:nodejs /app/package.json ./package.json

USER remix
RUN corepack prepare pnpm@8.5.0 --activate
RUN find ./build -type f -name '*.map' -delete

ENV NODE_ENV=production
ENV SENTRY_RELEASE=${SENTRY_RELEASE}

EXPOSE 3000

CMD ["corepack", "pnpm", "start"]
