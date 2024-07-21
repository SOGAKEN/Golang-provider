# ビルドステージ
FROM golang:1.22 as builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/server

# 実行ステージ
FROM alpine:latest  
RUN apk --no-cache add ca-certificates
RUN apk add --no-cache tzdata
WORKDIR /app/
COPY --from=builder /app/main .
COPY .env ./
# COPY --from=builder /app/config.toml .
EXPOSE 8080
CMD ["./main"]
