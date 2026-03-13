# Authentication

## 認証方式

リクエストごとに以下の優先順位で認証を試行する:

### 1. APIキー認証

```
Authorization: Bearer <api-key>
```

- CLI・外部連携で使用
- api_keysテーブルの`key_hash`とSHA-256で照合
- Web UIの設定画面から発行。生キーは発行時に1度だけ表示

### 2. セッション認証

```
Cookie: konbu_session=<user_id>:<hmac_signature>
```

- Web UIで使用
- ログイン時にHMAC-SHA256で署名したセッションCookieを発行
- `SESSION_SECRET` 環境変数が署名キー
- 有効期限: 30日
- HttpOnly, SameSite=Lax

### 3. 開発モード

```
DEV_USER=dev@local
```

- 環境変数設定時、そのメールアドレスで自動ログイン
- 認証ヘッダーもCookieも不要
- ローカル開発専用。本番では使わない

## 初回セットアップ

1. サーバー起動時、ユーザーが0人の場合は「未セットアップ」状態
2. `GET /api/v1/auth/setup-status` で状態を確認
3. Web UIがセットアップ画面を表示し、管理者アカウント作成を促す
4. `POST /api/v1/auth/register` でアカウント作成。最初のユーザーが管理者になる

## パスワード

- bcryptでハッシュ化して`users.password_hash`に保存
- 変更は `POST /api/v1/auth/change-password` (旧パスワード必須)

## ユーザー設定

- `users.user_settings` (JSONB) にユーザー固有の設定を格納
- 現在の設定項目: `first_day_of_week`, ウィジェット順序 等
- `GET/PUT /api/v1/auth/settings` で操作
