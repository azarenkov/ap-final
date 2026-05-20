FROM golang:1.25-alpine AS builder
WORKDIR /workspace
COPY gen ./gen
COPY notification-service ./notification-service
WORKDIR /workspace/notification-service
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/notification-service ./cmd/notification-service

FROM alpine:3.20
RUN adduser -D -u 10001 app
USER app
COPY --from=builder /out/notification-service /notification-service
EXPOSE 50054
ENTRYPOINT ["/notification-service"]
