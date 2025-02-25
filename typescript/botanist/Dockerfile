ARG NODE_IMAGE_PREFIX=""

# Install dependencies only when needed
FROM ${NODE_IMAGE_PREFIX}node:16-alpine AS deps
RUN apk add --no-cache python3 make g++ curl
RUN corepack enable && corepack prepare pnpm@8.5.0 --activate
WORKDIR /app
COPY package.json pnpm-lock.yaml ./

RUN pnpm install --frozen-lockfile

# Rebuild the source code only when needed
FROM ${NODE_IMAGE_PREFIX}node:16-alpine AS dev
WORKDIR /app
COPY . .
COPY --from=deps /app/node_modules ./node_modules
RUN corepack enable && corepack prepare pnpm@8.5.0 --activate
CMD ["pnpm", "dev"]

# Rebuild the source code only when needed
FROM ${NODE_IMAGE_PREFIX}node:16-alpine AS builder
WORKDIR /app
COPY . .
COPY --from=deps /app/node_modules ./node_modules
RUN corepack enable && corepack prepare pnpm@8.5.0 --activate
RUN pnpm build

# Production image, copy all the files and run remix
FROM ${NODE_IMAGE_PREFIX}node:16-alpine AS runner
WORKDIR /app

ENV NODE_ENV production

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

USER remix

EXPOSE 3000

CMD ["pnpm", "start"]
