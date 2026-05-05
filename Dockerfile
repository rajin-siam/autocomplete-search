FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o place-search ./cmd/place-search/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/place-search .
COPY config.yaml .
EXPOSE 2322
CMD ["./place-search", "serve"]
