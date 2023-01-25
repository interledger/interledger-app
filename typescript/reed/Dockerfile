# Install dependencies only when needed
FROM public.ecr.aws/docker/library/node:18-alpine AS deps
# Check https://github.com/nodejs/docker-node/tree/b4117f9333da4138b03a546ec926ef50a31506c3#nodealpine to understand why libc6-compat might be needed.
# RUN apk add --no-cache libc6-compat
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
RUN npm i -g pnpm
RUN pnpm i --frozen-lockfile


FROM deps AS builder
WORKDIR /app
COPY tsconfig.json ./
COPY src ./src
RUN pnpm build


FROM builder as runner
ENV NODE_ENV production

RUN pnpm i --prod

RUN addgroup -g 1001 -S nodejs
RUN adduser -S reed -u 1001

USER reed

EXPOSE 8080 8081

CMD ["pnpm", "start"]
