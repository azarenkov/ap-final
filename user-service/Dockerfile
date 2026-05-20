FROM golang:1.25-alpine AS builder
WORKDIR /workspace
COPY gen ./gen
COPY user-service ./user-service
WORKDIR /workspace/user-service
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/user-service ./cmd/user-service

FROM alpine:3.20
RUN adduser -D -u 10001 app
USER app
COPY --from=builder /out/user-service /user-service
EXPOSE 50052
ENTRYPOINT ["/user-service"]
