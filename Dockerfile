FROM golang:1.24-alpine AS builder
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
      -ldflags="-s -w" -trimpath \
      -o /out/npan-server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build \
      -ldflags="-s -w" -trimpath \
      -o /out/npan-cli ./cmd/cli

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S npan && adduser -S npan -G npan
COPY --from=builder /out/npan-server /usr/local/bin/npan-server
COPY --from=builder /out/npan-cli /usr/local/bin/npan-cli
COPY web/ /app/web/
RUN mkdir -p /app/data && chown -R npan:npan /app
VOLUME ["/app/data"]
WORKDIR /app
USER npan
EXPOSE 1323
HEALTHCHECK --interval=15s --timeout=3s --retries=3 \
    CMD wget -q -O /dev/null http://localhost:1323/healthz || exit 1
ENTRYPOINT ["npan-server"]
