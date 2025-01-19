FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY test-rdb/ ./test-rdb/

RUN CGO_ENABLED=0 GOOS=linux go build -o redis-server ./cmd/

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/redis-server .
COPY --from=builder /app/test-rdb /test-rdb

RUN mkdir -p /data

EXPOSE 6379

ENTRYPOINT ["./redis-server"]
CMD ["--port", "6379", "--dir", "/test-rdb", "--dbfilename", "dump.rdb"]
