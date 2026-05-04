FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o place-search ./cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/place-search .
COPY config.yaml .
COPY .env .
EXPOSE 2322
CMD ["./place-search"]
