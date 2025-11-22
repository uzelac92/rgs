FROM golang:1.25 AS builder

WORKDIR /app

COPY . .

# Build the mock from its subfolder
RUN go build -o walletmock ./walletmock

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/walletmock .
EXPOSE 9000
CMD ["./walletmock"]