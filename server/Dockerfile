# syntax=docker/dockerfile:1

FROM golang:1.22 AS builder
WORKDIR /app

# Cache deps
COPY server/go.mod ./
RUN go mod download

# Copy the rest
COPY server/. .

# Build static binary
ENV CGO_ENABLED=0
RUN go build -o /out/mss-server ./cmd/mss-server

FROM alpine:3.20
RUN adduser -D -H -s /sbin/nologin app && mkdir -p /data && chown -R app:app /data
COPY --from=builder /out/mss-server /usr/local/bin/mss-server
USER app
EXPOSE 8080
ENV MSS_LISTEN_ADDR=:8080
ENV MSS_DB_PATH=/data/mss.db
CMD ["/usr/local/bin/mss-server"]
