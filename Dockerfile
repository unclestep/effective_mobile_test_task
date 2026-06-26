FROM golang:alpine AS builder

WORKDIR /app

RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN swag init -g cmd/api/main.go
RUN CGO_ENABLED=0 go build -o server cmd/api/*.go

FROM alpine:latest

WORKDIR /app
RUN adduser -D -g '' -H -u 10001 appuser
RUN apk add --no-cache ca-certificates

COPY --from=builder --chown=appuser:appuser /app/server ./

USER 10001:10001
CMD ["./server"]
