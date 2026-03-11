FROM golang:1.25.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o subscription-service cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/subscription-service .
COPY --from=builder /app/.env .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./subscription-service"]