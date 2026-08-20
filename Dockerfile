FROM oven/bun:1.2.23-alpine AS base

LABEL org.opencontainers.image.source="https://github.com/arman/fiber-boilerplate"
LABEL org.opencontainers.image.description="Boilerplate repository build with Fiber"
LABEL org.opencontainers.image.licenses="MIT"

WORKDIR /app

COPY package.json bun.lock ./

RUN bun install --frozen-lockfile

COPY . .

CMD ["bun", "--version"]
