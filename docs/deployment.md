# Deployment

## Docker Compose（推奨）

最も簡単な方法。PostgreSQLとサーバーをまとめて起動する。

```bash
cp .env.example .env
# .env を編集: SESSION_SECRET を変更

docker compose up -d
```

`http://localhost:8080` でアクセス。初回起動時にアカウント作成画面が表示される。

### 本番環境（Traefik連携）

```bash
# .env を編集
# POSTGRES_PASSWORD, SESSION_SECRET, KONBU_DOMAIN を設定

docker compose -f docker-compose.prod.yml up -d
```

Traefikの外部ネットワーク `web` が必要。Let's EncryptでTLS証明書を自動取得する。

### ビルド済みイメージ

Dockerfileはマルチステージビルド:

1. Node.js: フロントエンドビルド
2. Go: サーバー+CLIバイナリビルド
3. scratch: 最終イメージ（バイナリ+静的ファイルのみ）

## ネイティブ（Docker不要）

### 前提

- Go 1.25+
- Node.js 22+
- PostgreSQL 16+

### 手順

```bash
# 1. フロントエンドビルド
cd web/frontend && npm ci && npm run build && cd ../..

# 2. サーバービルド
go build -o bin/server ./cmd/server

# 3. DB準備
createdb konbu
psql konbu -f sql/migrations/0001_initial.up.sql
psql konbu -f sql/migrations/0002_auth_password.up.sql
psql konbu -f sql/migrations/0003_recurring_events.up.sql
psql konbu -f sql/migrations/0004_tool_category.up.sql

# 4. 起動
DATABASE_URL="postgres://user:pass@localhost:5432/konbu?sslmode=disable" \
SESSION_SECRET="your-secret" \
./bin/server
```

### pg_bigm（任意）

日本語全文検索を使う場合はpg_bigm拡張をインストールする。

```sql
CREATE EXTENSION IF NOT EXISTS pg_bigm;
```

pg_bigmなしでも動作するが、日本語検索のパフォーマンスが低下する。

## systemdサービス

ネイティブビルドをsystemdで管理する例:

```ini
[Unit]
Description=konbu server
After=postgresql.service

[Service]
Type=simple
User=konbu
WorkingDirectory=/opt/konbu
ExecStart=/opt/konbu/bin/server
Environment=DATABASE_URL=postgres://konbu:password@localhost:5432/konbu?sslmode=disable
Environment=SESSION_SECRET=your-secret
Environment=PORT=8080
Restart=always

[Install]
WantedBy=multi-user.target
```

```bash
sudo cp konbu.service /etc/systemd/system/
sudo systemctl enable --now konbu
```

## リバースプロキシ

### Caddy

```
konbu.example.com {
    reverse_proxy localhost:8080
}
```

### nginx

```nginx
server {
    server_name konbu.example.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 環境変数

| 変数 | 必須 | デフォルト | 説明 |
|---|---|---|---|
| `DATABASE_URL` | Yes | — | PostgreSQL接続文字列 |
| `SESSION_SECRET` | Yes | 開発用フォールバック | セッション署名キー |
| `PORT` | No | `8080` | サーバーポート |
| `DEV_USER` | No | — | 開発用自動ログイン (メール形式) |
