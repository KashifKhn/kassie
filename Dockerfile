FROM node:20-alpine AS web-builder

WORKDIR /web

RUN corepack enable && corepack prepare pnpm@9 --activate

COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY web/ ./
RUN pnpm run build

FROM golang:1.24-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git make protobuf protobuf-dev bash

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=web-builder /web/dist ./web/dist

RUN make setup && make proto && make embed-web && go build -o kassie cmd/kassie/main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /build/kassie /app/kassie

RUN addgroup -g 1000 kassie && \
    adduser -D -u 1000 -G kassie kassie && \
    chown -R kassie:kassie /app

USER kassie

EXPOSE 50051 8080 9090 9091

ENTRYPOINT ["/app/kassie"]
CMD ["server"]
