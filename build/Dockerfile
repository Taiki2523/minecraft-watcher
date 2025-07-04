# Stage 1: Build Go binary statically
FROM golang:1.23 AS builder
WORKDIR /app

# モジュールファイルコピーとダウンロード
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY pkg ./pkg

# 静的にビルド（CGOを無効化、Linux向け）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o minecraft-watcher ./pkg/cmd

# Stage 2: ランタイムイメージ
FROM ubuntu:24.04

RUN apt-get update && apt-get install -y ca-certificates docker.io tzdata

WORKDIR /root/

# 実行ファイルコピー
COPY --from=builder /app/minecraft-watcher /usr/local/bin/minecraft-watcher

# 実行権限を明示（念のため）
RUN chmod +x /usr/local/bin/minecraft-watcher

# エントリポイント指定
ENTRYPOINT ["/usr/local/bin/minecraft-watcher"]
