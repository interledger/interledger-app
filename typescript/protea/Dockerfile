# Install dependencies only when needed
FROM public.ecr.aws/docker/library/node:18-alpine AS deps
RUN apk add --no-cache python3 make g++ curl
RUN corepack enable && corepack prepare pnpm@8.5.0 --activate
WORKDIR /app
COPY package.json pnpm-lock.yaml ./

RUN pnpm install --frozen-lockfile

# Rebuild the source code only when needed
FROM public.ecr.aws/docker/library/node:18-alpine AS dev
WORKDIR /app
COPY . .
COPY --from=deps /app/node_modules ./node_modules
RUN corepack enable && corepack prepare pnpm@8.5.0 --activate
CMD ["pnpm", "dev"]

# Rebuild the source code only when needed
FROM public.ecr.aws/docker/library/node:18-alpine AS builder
ARG REMIX_PUBLIC_PATH
ARG IMAGE_HASH
RUN apk add --no-cache curl unzip
ARG BUILDARCH=amd64
RUN curl -O https://downloads.rclone.org/rclone-current-linux-${BUILDARCH}.zip && unzip rclone-current-linux-${BUILDARCH}.zip && cd rclone-*-linux-${BUILDARCH} && cp rclone /usr/bin/rclone
WORKDIR /app
COPY . .
COPY --from=deps /app/node_modules ./node_modules
RUN corepack enable && corepack prepare pnpm@8.5.0 --activate
RUN REMIX_PUBLIC_PATH=${REMIX_PUBLIC_PATH} pnpm build --sourcemap
RUN --mount=type=secret,id=sentry,dst=.env.sentry SENTRY_DOTENV_PATH=.env.sentry IMAGE_HASH=${IMAGE_HASH} REMIX_PUBLIC_PATH=${REMIX_PUBLIC_PATH} pnpm upload-sourcemaps
RUN --mount=type=secret,id=cdn,dst=.env.rclone rclone --config=.env.rclone --progress sync public/build cdn:protea/build/

## Production image, copy all the files and run remix
FROM node:18-alpine AS runner
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
RUN find ./public/build -type f -name '*.map' -delete
RUN find ./build -type f -name '*.map' -delete

USER remix

EXPOSE 3000

CMD ["pnpm", "start"]
