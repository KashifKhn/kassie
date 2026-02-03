FROM golang:1.24-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git make protobuf protobuf-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make setup && make proto && make build-server

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /build/kassie /app/kassie

RUN addgroup -g 1000 kassie && \
    adduser -D -u 1000 -G kassie kassie && \
    chown -R kassie:kassie /app

USER kassie

EXPOSE 50051 8080

ENTRYPOINT ["/app/kassie"]
CMD ["server"]
