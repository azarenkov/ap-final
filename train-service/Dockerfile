FROM golang:1.25-alpine AS builder
WORKDIR /workspace
COPY gen ./gen
COPY train-service ./train-service
WORKDIR /workspace/train-service
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/train-service ./cmd/train-service

FROM alpine:3.20
RUN adduser -D -u 10001 app
USER app
COPY --from=builder /out/train-service /train-service
EXPOSE 50051
ENTRYPOINT ["/train-service"]
