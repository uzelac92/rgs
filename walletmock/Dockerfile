FROM golang:1.25 AS builder

WORKDIR /app

COPY walletmock/ ./walletmock

WORKDIR /app/walletmock

RUN go mod init walletmock || true
RUN go mod tidy || true

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /walletmock .

FROM alpine:3.19
WORKDIR /app

COPY --from=builder /walletmock .

EXPOSE 9000

CMD ["./walletmock"]