ARG NODE_IMAGE_PREFIX=""

# Install workspace dependencies needed to build/run botanist (and its
# workspace dependency, @interledger/configa). Build context is the
# typescript/ workspace root (see local/botanist.yaml, k8s/Makefile, CI).
FROM ${NODE_IMAGE_PREFIX}node:24-alpine AS deps
RUN apk add --no-cache python3 make g++ curl py3-setuptools
RUN corepack enable && corepack prepare pnpm@8.5.0 --activate
WORKDIR /app
COPY pnpm-workspace.yaml package.json pnpm-lock.yaml .npmrc ./
COPY packages/configa/package.json ./packages/configa/package.json
COPY botanist/package.json ./botanist/package.json

RUN pnpm install --frozen-lockfile --filter botanist...

# @interledger/configa ships pre-built (dist/) — build it once here, shared
# by the dev and builder stages below.
FROM deps AS configa-build
COPY packages/configa ./packages/configa
RUN pnpm --filter @interledger/configa build

# Dev image: full workspace source; app/ is bind-mounted over this for live
# reload (see local/botanist.yaml).
FROM ${NODE_IMAGE_PREFIX}node:24-alpine AS dev
RUN corepack enable && corepack prepare pnpm@8.5.0 --activate
WORKDIR /app
COPY --from=configa-build /app ./
COPY botanist ./botanist
WORKDIR /app/botanist
CMD ["pnpm", "dev"]

# Build botanist's production bundle
FROM ${NODE_IMAGE_PREFIX}node:24-alpine AS builder
RUN corepack enable && corepack prepare pnpm@8.5.0 --activate
WORKDIR /app
COPY --from=configa-build /app ./
COPY botanist ./botanist
WORKDIR /app/botanist
RUN pnpm build

# Production image. Preserves the workspace's relative layout
# (/app/{node_modules,packages,botanist}) so botanist's workspace-dependency
# and hoisted-dependency symlinks (e.g. @interledger/configa, @react-router/serve)
# keep resolving.
FROM ${NODE_IMAGE_PREFIX}node:24-alpine AS runner
WORKDIR /app/botanist

ENV NODE_ENV=production

RUN addgroup -g 1001 -S nodejs
RUN adduser -S remix -u 1001
RUN corepack enable && corepack prepare pnpm@8.5.0 --activate

COPY --from=builder --chown=remix:nodejs /app/botanist/build ./build
COPY --from=builder --chown=remix:nodejs /app/botanist/node_modules ./node_modules
COPY --from=builder --chown=remix:nodejs /app/botanist/package.json ./package.json
COPY --from=builder --chown=remix:nodejs /app/botanist/bootstrap.mjs ./bootstrap.mjs
COPY --from=builder --chown=remix:nodejs /app/botanist/app/config.schema.mjs ./app/config.schema.mjs
COPY --from=builder --chown=remix:nodejs /app/node_modules /app/node_modules
COPY --from=builder --chown=remix:nodejs /app/packages /app/packages

USER remix

EXPOSE 3000

CMD ["pnpm", "start"]
