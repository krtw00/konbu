<p align="center">
  <img src="web/static/favicon.svg" width="64" height="64" alt="konbu">
</p>

<h1 align="center">konbu</h1>

<p align="center">セルフホスト型の個人ツール基盤<br>メモ・ToDo・カレンダー・ツールランチャーを一箇所に</p>

---

## 概要

konbu は「これ一つ開けば個人の管理が全部できる」を目指すWebアプリ + CLI。

- **Memos** — Markdown メモ。タグ分類、全画面 CodeMirror 6 エディタ、ライブプレビュー
- **ToDo** — インラインタスク追加、日付・タグ管理、右パネル詳細表示
- **Calendar** — 月間カレンダー、右パネルでイベント作成・編集
- **Tools** — ブックマーク管理。favicon 自動取得

## 技術スタック

| レイヤー | 技術 |
|----------|------|
| Backend | Go, chi, PostgreSQL (pg_bigm) |
| Frontend | Vanilla JS, CodeMirror 6, CSS Variables テーマ |
| CLI | Go, cobra |
| Infra | Docker, マルチステージビルド |

## セットアップ

```bash
# 起動
docker compose up -d

# CLI ビルド
go build -o bin/konbu ./cmd/konbu
```

デフォルトで `http://localhost:8080` でアクセス可能。

## CLI

```bash
konbu memo list                        # メモ一覧
konbu memo add "title" -c "content"    # メモ作成
konbu memo show <id>                   # 内容表示

konbu todo list                        # ToDo一覧
konbu todo add "task name"             # タスク追加
konbu todo done <id>                   # 完了

konbu tool list                        # ツール一覧
konbu tool add "name" "https://..."    # ツール追加
```

`KONBU_API` 環境変数または `--api` フラグで接続先を変更可能。

## テーマ

7 種類のカラーテーマを同梱（右上のドットで切り替え）:

Konbu / Notion / Solarized / Catppuccin Latte / Nord / Linear / Catppuccin Mocha

## ディレクトリ構成

```
cmd/
  server/       # API サーバー
  konbu/        # CLI
internal/
  handler/      # HTTP ハンドラ
  service/      # ビジネスロジック
  repository/   # DB アクセス (sqlc)
  client/       # CLI 用 API クライアント
web/static/     # フロントエンド (HTML/CSS/JS)
sql/            # スキーマ・マイグレーション
```

## ライセンス

Private
