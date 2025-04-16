FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY src/go.mod src/go.sum ./

RUN go mod download

COPY src/. .

RUN go build -o app .

FROM alpine:latest AS production

WORKDIR /app

COPY --from=builder /app/app .
COPY .env .

CMD ["./app"]