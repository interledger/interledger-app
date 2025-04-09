ARG NODE_IMAGE_PREFIX=""

# Install dependencies only when needed
FROM ${NODE_IMAGE_PREFIX}node:18.20-alpine AS deps
RUN apk add --no-cache python3 make g++ curl py3-setuptools
RUN corepack enable && corepack prepare pnpm@8.5.0 --activate
WORKDIR /app
COPY package.json pnpm-lock.yaml ./

RUN pnpm install --frozen-lockfile

# Rebuild the source code only when needed
FROM ${NODE_IMAGE_PREFIX}node:18.20-alpine AS dev 
WORKDIR /app
COPY . .
COPY --from=deps /app/node_modules ./node_modules
RUN corepack enable && corepack prepare pnpm@8.5.0 --activate
CMD ["pnpm", "dev"]

# Rebuild the source code only when needed
FROM ${NODE_IMAGE_PREFIX}node:18-alpine AS builder
# TODO: Uploading the client bundle and source maps will be handled later

# RUN apk add --no-cache curl unzip
# RUN curl -O https://downloads.rclone.org/rclone-current-linux-${BUILDARCH}.zip && unzip rclone-current-linux-${BUILDARCH}.zip && cd rclone-*-linux-${BUILDARCH} && cp rclone /usr/bin/rclone

ARG IMAGE_HASH
ARG BUILDARCH
ARG SENTRY_ORG
ARG SENTRY_PROJECT
ARG SENTRY_AUTH_TOKEN
ARG UPLOAD_SOURCEMAPS
WORKDIR /app
COPY . .
COPY --from=deps /app/node_modules ./node_modules
RUN corepack enable && corepack prepare pnpm@8.5.0 --activate
RUN pnpm build

RUN pnpm build --sourcemap

ENV SENTRY_ORG=${SENTRY_ORG}
ENV SENTRY_PROJECT=${SENTRY_PROJECT}
ENV SENTRY_AUTH_TOKEN=${SENTRY_AUTH_TOKEN}

RUN [ -n "${SENTRY_PROJECT}" ] && [ "${UPLOAD_SOURCEMAPS}" = "true" ] && IMAGE_HASH=${IMAGE_HASH} pnpm upload-sourcemaps || echo "Skipping sourcemap upload: SENTRY_PROJECT is empty or UPLOAD_SOURCEMAPS is not 'true'"
# REMIX_PUBLIC_PATH=${REMIX_PUBLIC_PATH}
# RUN --mount=type=secret,id=cdn,dst=.env.rclone rclone --config=.env.rclone --progress sync public/build cdn:protea/build/

# Production image, copy all the files and run remix
FROM ${NODE_IMAGE_PREFIX}node:18.20-alpine AS runner
ARG IMAGE_HASH

WORKDIR /app

ENV NODE_ENV=production
ENV SENTRY_RELEASE=${IMAGE_HASH}

RUN addgroup -g 1001 -S nodejs
RUN adduser -S remix -u 1001
RUN corepack enable && corepack prepare pnpm@8.5.0 --activate

# You only need to copy remix.config.js if you are NOT using the default configuration
# COPY --from=builder /app/remix.config.js ./
COPY --from=builder /app/public ./public
COPY --from=builder --chown=remix:nodejs /app/build ./build
COPY --from=builder --chown=remix:nodejs /app/public/build ./public/build
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./package.json
RUN find ./public/build -type f -name '*.map' -delete
RUN find ./build -type f -name '*.map' -delete

USER remix

EXPOSE 3000

CMD ["pnpm", "start"]
