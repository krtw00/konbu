---
depends_on:
  - ../02-architecture/structure.md
tags: [details, flows, sequence, process]
ai_summary: "konbuの主要処理フロー -- 認証、リソースCRUD、横断検索のシーケンス図"
---

# 主要フロー

> **Status**: Active | 最終更新: 2026-03-14

本ドキュメントは、konbuの主要な処理フローを定義する。

---

## フロー一覧

| フローID | フロー名 | 説明 |
|----------|----------|------|
| F001 | 初期セットアップ | 初回起動時のアカウント作成 |
| F002 | ログイン認証 | Web UIからのセッション認証 |
| F003 | APIキー認証 | CLI/外部連携からのBearer認証 |
| F004 | リソースCRUD | メモ・ToDo・予定の作成（タグupsert含む） |
| F005 | 横断検索 | pg_trgmによる全リソース横断検索 |

---

## フロー詳細

### F001: 初期セットアップ

| 項目 | 内容 |
|------|------|
| 概要 | サーバー初回起動時、最初のユーザーを管理者として登録する |
| トリガー | ブラウザでkonbuにアクセス |
| アクター | 初回ユーザー |
| 前提条件 | usersテーブルが空 |
| 事後条件 | 管理者アカウントが作成され、セッションが発行される |

```mermaid
sequenceDiagram
    actor User as ユーザー
    participant FE as Web UI
    participant API as API Server
    participant DB as PostgreSQL

    User->>FE: アクセス
    FE->>API: GET /auth/setup-status
    API->>DB: SELECT count(*) FROM users
    DB-->>API: 0
    API-->>FE: {setup_required: true}
    FE-->>User: セットアップ画面表示
    User->>FE: メール・パスワード入力
    FE->>API: POST /auth/register
    API->>DB: INSERT INTO users (is_admin=true)
    DB-->>API: created
    API-->>FE: 200 + Set-Cookie: konbu_session
    FE-->>User: ダッシュボード表示
```

---

### F002: ログイン認証（Web UI）

| 項目 | 内容 |
|------|------|
| 概要 | メール+パスワードでログインし、HMAC署名セッションCookieを発行する |
| トリガー | ログイン画面でCredentials送信 |
| アクター | Webユーザー |

```mermaid
sequenceDiagram
    actor User as ユーザー
    participant FE as Web UI
    participant API as API Server
    participant DB as PostgreSQL

    User->>FE: メール・パスワード入力
    FE->>API: POST /auth/login
    API->>DB: SELECT * FROM users WHERE email = ?
    DB-->>API: user record
    API->>API: bcrypt.Compare(password, hash)
    API->>API: HMAC-SHA256署名Cookie生成
    API-->>FE: 200 + Set-Cookie: konbu_session
    FE-->>User: ダッシュボード表示
```

#### エラーケース

| エラー | 条件 | 対応 |
|--------|------|------|
| 認証失敗 | メールまたはパスワードが不一致 | 401 Unauthorized |
| ユーザー未登録 | メールアドレスが存在しない | 401 Unauthorized（同一メッセージ） |

---

### F003: APIキー認証（CLI）

| 項目 | 内容 |
|------|------|
| 概要 | Bearer tokenでAPIキーを照合し、ユーザーコンテキストを注入する |
| トリガー | CLIコマンド実行 |
| アクター | CLIユーザー / AIエージェント |

```mermaid
sequenceDiagram
    actor User as CLI
    participant API as API Server
    participant DB as PostgreSQL

    User->>API: GET /api/v1/memos<br/>Authorization: Bearer <api-key>
    API->>API: SHA-256(api-key)
    API->>DB: SELECT * FROM api_keys WHERE key_hash = ?
    DB-->>API: api_key record (user_id)
    API->>DB: UPDATE api_keys SET last_used_at = now()
    API->>API: ユーザーコンテキスト注入
    API->>DB: SELECT * FROM memos WHERE user_id = ?
    DB-->>API: memo records
    API-->>User: JSON response
```

---

### F004: リソースCRUD（メモ作成例）

| 項目 | 内容 |
|------|------|
| 概要 | メモを作成し、タグの暗黙的upsertを行う |
| トリガー | Web UIまたはCLIからメモ作成リクエスト |
| アクター | 認証済みユーザー |

```mermaid
sequenceDiagram
    actor User as ユーザー
    participant H as Handler
    participant S as Service
    participant R as Repository
    participant DB as PostgreSQL

    User->>H: POST /api/v1/memos<br/>{title, content, tags}
    H->>H: バリデーション
    H->>S: CreateMemo(ctx, req)
    S->>DB: BEGIN
    S->>R: InsertMemo(title, content)
    R->>DB: INSERT INTO memos
    DB-->>R: memo record
    loop 各タグ
        S->>R: UpsertTag(user_id, tag_name)
        R->>DB: INSERT ON CONFLICT DO NOTHING
        S->>R: InsertMemoTag(memo_id, tag_id)
        R->>DB: INSERT INTO memo_tags
    end
    S->>DB: COMMIT
    S-->>H: memo with tags
    H-->>User: 201 {data: {...}}
```

---

### F005: 横断検索

| 項目 | 内容 |
|------|------|
| 概要 | メモ・ToDo・予定をpg_trgm全文検索で横断的に検索する |
| トリガー | 検索クエリの入力 |
| アクター | 認証済みユーザー |

```mermaid
sequenceDiagram
    actor User as ユーザー
    participant API as API Server
    participant DB as PostgreSQL

    User->>API: GET /api/v1/search?q=デプロイ
    API->>DB: SELECT 'memo', id, title, snippet FROM memos<br/>WHERE (title ILIKE '%デプロイ%' OR content ILIKE '%デプロイ%')<br/>AND user_id = ? AND deleted_at IS NULL
    API->>DB: SELECT 'todo', id, title, snippet FROM todos<br/>WHERE (title ILIKE '%デプロイ%' OR description ILIKE '%デプロイ%')<br/>AND user_id = ? AND deleted_at IS NULL
    API->>DB: SELECT 'event', id, title, snippet FROM calendar_events<br/>WHERE (title ILIKE '%デプロイ%' OR description ILIKE '%デプロイ%')<br/>AND user_id = ? AND deleted_at IS NULL
    DB-->>API: 統合結果
    API-->>User: {data: [{type, id, title, snippet}, ...]}
```

---

## 関連ドキュメント

- [data-model.md](./data-model.md) - データモデル
- [api.md](./api.md) - API設計
- [ui.md](./ui.md) - UI設計
