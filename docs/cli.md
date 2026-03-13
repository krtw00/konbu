# CLI

## 概要

konbu CLIはリモートのkonbuサーバーにREST API経由で接続するスタンドアロンクライアント。サーバーコード（handler, service, repository）には依存せず、`internal/client` パッケージのHTTPクライアントのみ使用する。

```
go install github.com/krtw00/konbu/cmd/konbu@latest
```

## 設計方針

- **APIクライアントとして独立** — サーバーとCLIは同じリポジトリだがバイナリは別。CLIにサーバーコードは含まれない
- **JSON出力対応** — `--json` フラグで全コマンドが機械可読なJSON出力に。AI連携やスクリプトでの利用を想定
- **短縮ID** — UUID先頭8文字で操作可能。list結果からコピペしてそのまま使える
- **環境変数ベース** — `KONBU_API` + `KONBU_API_KEY` で接続先を設定

## セットアップ

```bash
export KONBU_API=https://konbu.example.com
export KONBU_API_KEY=your-api-key
```

APIキーはWeb UIの「設定 > セキュリティ」で発行する。

## コマンド体系

```
konbu memo     list | show | add | edit | rm
konbu todo     list | show | add | edit | done | reopen | rm
konbu event    list | show | add | edit | rm
konbu tool     list | add | edit | rm
konbu tag      list | rm
konbu search   <query>
konbu api-key  list | create | rm
konbu export   json | markdown
konbu import   ical
```

## AI連携

CLIは人間だけでなくAIエージェントからの利用も想定している。

### AIエージェントが使う場合の想定フロー

```bash
# 1. データの取得（JSON出力）
konbu memo list --json
konbu todo list --json
konbu event list --json

# 2. 詳細の確認
konbu memo show <id> --json
konbu todo show <id> --json

# 3. データの操作
konbu memo add "議事録" -c "$(cat notes.md)"
konbu todo add "レビュー依頼" -d 2026-03-15
konbu todo done <id>

# 4. 横断検索
konbu search "デプロイ" --json
```

`--json` 出力は各リソースのAPIレスポンスと同じ構造を返すため、AIがパースしやすい。

### stdin対応

メモの内容はstdinから流し込める:

```bash
cat document.md | konbu memo add "ドキュメント" -c -
echo "会議メモ" | konbu memo edit <id> -c -
```

## 内部構成

```
cmd/konbu/main.go     cobraコマンド定義、フラグ処理、出力整形
internal/client/      HTTPクライアント（全APIエンドポイントに対応するメソッド）
```

client パッケージは `do(method, path, body)` を基盤に、各リソースのCRUDメソッドを提供する。認証はBearer tokenヘッダーで行う。
