FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@latest && \
    go mod tidy && \
    swag init -g cmd/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/main.go

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 4200

CMD ["./server"]
