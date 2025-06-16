# Minecraft Watcher  
Minecraftサーバーのログ監視・Discord通知・ヘルスチェックを行うGo製ツール

---

## 概要

このプロジェクトは、Minecraftサーバーのログファイルを監視し、  
プレイヤーの入退出やサーバーの起動・停止イベントをDiscordに通知するシステムです。  
また、Dockerコンテナのヘルスチェックや定期的な参加者リスト通知もサポートしています。

---

## 主な機能

- **ログ監視**  
  サーバーログからプレイヤーの入退出を検知し、Discordに通知

- **プレイヤーリスト管理**  
  現在の参加者リストをファイルで管理し、定期的にDiscord通知

- **サーバーヘルスチェック**  
  Dockerコンテナの状態を監視し、起動・停止時にDiscord通知

- **柔軟な設定**  
  環境変数でログファイルパスやWebhook URL、監視間隔などを指定可能

---

## ディレクトリ構成と役割

```
pkg/
├── application/         // サービス層（業務ロジック）
│   └── player_service.go
├── cmd/                 // エントリポイント
│   └── main.go
├── domain/              // ドメイン層（インターフェース定義）
│   ├── nortifier.go
│   └── player.go
├── infrastructure/      // インフラ層（外部サービス・ファイル操作）
│   ├── discord_notifier.go
│   └── player_file_repository.go
├── interfaces/          // インターフェース層（監視・通知の実装）
│   ├── container_health.go
│   ├── player_check.go
│   └── watcher.go
└── internal/            // 内部ユーティリティ
    ├── message.go
    ├── timezone.go
    └── zerolog.go
```

---

## 主なファイルの解説

### 1. `cmd/main.go`  
アプリケーションのエントリポイント。  
- 環境変数の読み込み・ロガー初期化
- 各種サービス・リポジトリ・Notifierの生成
- ログ監視、ヘルスチェック、定期通知のgoroutine起動
- 終了シグナルのハンドリング

### 2. `application/player_service.go`  
プレイヤーの入退出やリスト通知など、業務ロジックを提供。

### 3. `domain/player.go` / `domain/nortifier.go`  
- `PlayerRepository`/`Notifier` インターフェースを定義

### 4. `infrastructure/player_file_repository.go`  
- プレイヤーリストのファイル保存・取得を実装（`PlayerRepository`実装）

### 5. `infrastructure/discord_notifier.go`  
- Discord Webhookへの通知を実装（`Notifier`実装）

### 6. `interfaces/watcher.go`  
- サーバーログファイルを監視し、入退出イベントを検知して通知

### 7. `interfaces/player_check.go`  
- 定期的に現在の参加者リストをDiscordに通知

### 8. `interfaces/container_health.go`  
- Dockerコンテナのヘルスチェックを監視し、起動・停止時に通知

### 9. `internal/message.go`  
- Discord通知用のメッセージ生成（イベント・リスト）

### 10. `internal/timezone.go`  
- タイムゾーン対応の現在時刻取得

### 11. `internal/zerolog.go`  
- ログ出力の初期化・環境変数のダンプ

---

## 環境変数

| 変数名                  | 用途                                 | 例                      |
|-------------------------|--------------------------------------|-------------------------|
| LOG_FILE                | 監視するMinecraftログファイルパス     | `/data/logs/latest.log` |
| DISCORD_WEBHOOK_URL     | Discord WebhookのURL                 | `https://discord.com/api/webhooks/...` |
| LOG_LEVEL               | ログレベル (`debug` or `info`)       | `debug`                 |
| PLAYER_CHECK_INTERVAL   | 参加者リスト通知の間隔（例: `5m`）   | `5m`                    |
| MC_CONTAINER_NAME       | 監視対象のDockerコンテナ名            | `mc-server`             |
| MC_MONITOR_INTERVAL     | ヘルスチェック間隔（例: `10s`）      | `10s`                   |
| TIMEZONE                | タイムゾーン（例: `Asia/Tokyo`）     | `Asia/Tokyo`            |

---

## 実行例

```sh
export LOG_FILE="/data/logs/latest.log"
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..."
export MC_CONTAINER_NAME="mc-server"
export PLAYER_CHECK_INTERVAL="5m"
export MC_MONITOR_INTERVAL="10s"
export TIMEZONE="Asia/Tokyo"
go run ./pkg/cmd/main.go
```

---

## 拡張・カスタマイズ

- プレイヤーリストの保存先や通知内容は`infrastructure`や`internal/message.go`で調整可能
- Discord以外の通知先を追加したい場合は`domain.Notifier`を実装
- 他の永続化方式（DB等）を使いたい場合は`domain.PlayerRepository`を実装

---

## ライセンス

MIT License

---

## 作者

taiki2523

---

## 補足

- DDD（ドメイン駆動設計）を意識したレイヤ分割
- テストや拡張がしやすい構成です
- ご質問・要望はIssueまたはPRでどうぞ

---