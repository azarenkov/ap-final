FROM golang:1.25-alpine AS builder
WORKDIR /workspace
COPY gen ./gen
COPY api-gateway ./api-gateway
WORKDIR /workspace/api-gateway
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api-gateway ./cmd/api-gateway

FROM alpine:3.20
RUN adduser -D -u 10001 app
USER app
COPY --from=builder /out/api-gateway /api-gateway
EXPOSE 8080
ENTRYPOINT ["/api-gateway"]
