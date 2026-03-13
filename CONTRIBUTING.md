# Contributing

konbuへのコントリビュートに興味を持っていただきありがとうございます。

## 開発環境のセットアップ

```bash
# リポジトリクローン
git clone https://github.com/krtw00/konbu.git
cd konbu

# PostgreSQL起動
docker compose up -d postgres

# フロントエンド
cd web/frontend && npm ci && cd ../..

# サーバー起動（開発モード）
DEV_USER=dev@local go run ./cmd/server

# フロントエンド開発サーバー（別ターミナル）
cd web/frontend && npm run dev
```

## 開発フロー

1. issueで事前に相談（大きな変更の場合）
2. featureブランチを切る
3. 変更を加える
4. テストが通ることを確認: `go test ./...`
5. Pull Requestを作成

## コーディング規約

### Go

- パッケージ名: 小文字単一語 (`handler`, `service`)
- ファイル名: スネークケース (`memo_handler.go`)
- レイヤー: handler → service → repository の一方向依存

### フロントエンド

- TypeScript必須
- コンポーネント: shadcn/uiベース
- 状態管理: Zustand（グローバル最小限、ローカルstate優先）
- i18n: 全テキストはen.json/ja.jsonに定義

### SQL

- テーブル名: スネークケース複数形 (`memos`, `calendar_events`)
- カラム名: スネークケース (`user_id`, `created_at`)
- マイグレーション: `NNNN_description.up.sql` / `.down.sql`

### API

- パス: ケバブケース複数形 (`/api/v1/api-keys`)
- JSONフィールド: スネークケース (`created_at`)

## Issue / Pull Request

- Issueにはできるだけ再現手順や期待する動作を書いてください
- PRは小さく保つ。1つのPRで1つの変更
- 日本語・英語どちらでも構いません

## ライセンス

コントリビュートされたコードはMITライセンスの下で公開されます。
