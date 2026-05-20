FROM golang:1.25-alpine AS builder
WORKDIR /workspace
COPY gen ./gen
COPY booking-service ./booking-service
WORKDIR /workspace/booking-service
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/booking-service ./cmd/booking-service

FROM alpine:3.20
RUN adduser -D -u 10001 app
USER app
COPY --from=builder /out/booking-service /booking-service
EXPOSE 50053
ENTRYPOINT ["/booking-service"]
