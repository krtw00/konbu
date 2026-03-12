# VPS 個人ツール基盤 — 構想メモ

> 作成: 2026-03-12
> ステータス: 構想段階（未着手）

---

## ビジョン

**「これ一つ開けば、個人に関する管理が全部できる」**

- 知識（メモ・ナレッジ）、行動（ToDo・カレンダー）、ツール（WebSSH等）を一箇所に集約
- ブラウザでもCLIでも同じデータに触れる
- 検索性と使いやすさに徹底的にこだわる
- 散らばったツール・ブックマーク・Obsidian を1つのポータルに一元化する
- **モバイルファースト**。出先のスマホから全機能が快適に使えること（PWA対応）

### 設計上の最重要テーマ: 知識の一元化と検索性

個人の知識・情報が複数ツールに分散している状態をなくす。
「あれどこに書いたっけ」をゼロにする。

- 全文検索がどこからでも即座に使える（Web UI のグローバル検索、CLI の `portal search`）
- メモ・ToDo・カレンダーを横断して検索できる
- タグによる構造化で、検索に頼らなくても辿れる導線を持つ

---

## 背景・動機

- 自宅サーバーは Tailscale 限定アクセスのため、外部PC（ネカフェ等）からアクセスできない
- Android からの Claude Code 操作に自作 WebSSH を使用中だが、現状は自宅サーバー上で Basic 認証のみ
- VPS に WebSSH を含む個人ツール群を置き、**ブラウザさえあればどこからでもアクセス可能**にしたい
- 会社の統一ID管理の個人版として、**Google OAuth による統一認証基盤**を構築したい

---

## 全体構成（案）

```
Internet
  │
  ▼
Traefik (VPS) ── TLS終端 + ForwardAuth
  │
  ├─ portal.example.com   → ポータル（リンク集・ダッシュボード）
  │                          ← ForwardAuth 適用
  │
  ├─ webssh.example.com   → WebSSH（自作）
  │                          ← ForwardAuth 適用
  │                          → Tailscale → 自宅Arch (sshd + tmux + Claude Code)
  │
  ├─ zip.example.com      → Zipline（画像/ファイルアップローダー）
  │                          ← ForwardAuth 適用 or 独自認証のみ（要判断）
  │
  └─ auth.example.com     → 認証サービス
```

### ネットワーク経路

```
[外部PC/Android ブラウザ]
    │ HTTPS
    ▼
[VPS: Traefik + 認証サービス + WebSSH + ポータル]
    │ Tailscale トンネル
    ▼
[自宅 Arch: sshd → tmux → Claude Code]
```

---

## 認証基盤（未決定・選択肢）

### 要件

- Google OAuth でログイン（普段 Google にログイン済みなら再認証ほぼ不要）
- 許可する Google アカウントをホワイトリスト指定
- ポータル・WebSSH 等の全サービスを統一認証で保護
- ネカフェPC 想定: PC 側にクレデンシャルが残らないこと

### 候補

| 候補 | 特徴 | 構成 | 適合度 |
|------|------|------|--------|
| **traefik-forward-auth** | Google OAuth 特化。環境変数だけで設定完了 | コンテナ1つ | ◎ シンプル最優先なら |
| **Authelia** | 軽量SSO。YAML設定。OIDC認定済み。TOTP/パスキー対応 | コンテナ1つ（20MB） | ◎ MFA追加したいなら |
| **Authentik** | フルIdP。管理Web UI。OAuth/OIDC/SAML/LDAP | PostgreSQL + Redis + コンテナ複数 | △ 個人利用にはオーバーキル |

### 判断ポイント（後で決める）

- Google の 2FA だけで十分か、追加の TOTP/パスキーが欲しいか
- Zipline を認証基盤に含めるか、独自認証のままにするか
- 将来ツールを増やす予定があるか（増やすなら Authelia が拡張しやすい）

---

## ポータル（自作方針）

### コンセプト

- **個人グループウェアのホーム画面**
- ログインしたらまずここが開く。日常的に使うダッシュボード
- 2つの領域で構成:
  - **ダッシュボード領域**: カレンダー、ToDo、メモ（ポータル内で直接表示・操作）
  - **ツールランチャー領域**: WebSSH、Zipline 等のカード（クリックで別タブに開く）
- 一般のWebブックマーク管理は対象外（自作ツール・個人サービスのURL一元化のみ）

### 既存OSSではなく自作する理由

- ToDo・メモのCRUDを自前で持つ必要がある（既存ダッシュボードOSSは表示専用ウィジェットのみ）
- カレンダーAPI連携 + 自前データ管理のハイブリッド構成は既存OSSで対応困難
- WebSSH と技術スタックを統一できる

### 画面構成（イメージ）

```
┌─────────────────────────────────────────┐
│  ポータル（認証済みホーム画面）             │
│                                         │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ │
│  │ カレンダー │ │  ToDo    │ │  メモ    │ │
│  │(API連携)  │ │(自前DB)  │ │(自前DB)  │ │
│  │ 表示+操作 │ │ CRUD     │ │ CRUD     │ │
│  └──────────┘ └──────────┘ └──────────┘ │
│                                         │
│  ── ツール ──────────────────────────── │
│  ┌────────┐ ┌────────┐ ┌────────┐       │
│  │ WebSSH │ │Zipline │ │ 将来   │       │
│  │→別タブ  │ │→別タブ  │ │→別タブ  │       │
│  └────────┘ └────────┘ └────────┘       │
└─────────────────────────────────────────┘
```

### データ管理方針

| 機能 | データの置き場 | 理由 |
|------|--------------|------|
| 個人カレンダー | 自前（PostgreSQL） | TimeTree APIは2023年末に終了済みで外部連携不可。個人用は自前で管理 |
| ToDo | 自前（PostgreSQL） | ポータル固有の機能として管理 |
| メモ | 自前（PostgreSQL） | ポータル固有の機能として管理。JSONB でテーブル型メモにも対応 |
| ツールリンク | 自前（JSON or YAML設定） | 静的な設定データ |

### カレンダーの棲み分け

| 用途 | ツール | ポータルでの扱い |
|------|--------|-----------------|
| 家族共有カレンダー | TimeTree（既存運用） | リンクのみ（別タブでTimeTree Webを開く） |
| 個人カレンダー | ポータル自前機能 | ダッシュボードに表示・操作（CRUD） |

### 技術スタック（案）

- バックエンド: Go（echo or chi）+ sqlc
- フロントエンド: 未選定（後で検討）
- データ: PostgreSQL（カレンダー・ToDo・メモ・ユーザー）+ 設定ファイル（ツールリンク）
- 検索: PostgreSQL pg_bigm or PGroonga（日本語全文検索）
- 認証: ForwardAuth に委譲（ポータル自体は認証機能を持たない）
- CLI: Go シングルバイナリ（cobra）
- デプロイ: Docker（scratch ベース、イメージ ~20MB）

### アーキテクチャ: Web UI + CLI 両対応

```
[ブラウザ]────→ ポータル Web UI ─┐
                                  ├→ Go REST API → PostgreSQL
[ターミナル]──→ CLI ─────────────┘
                 ↑                        ↑
          Claude Code              curl / bot / cron
          Slack bot
          SSH越し操作
```

**設計原則**:

1. REST API がデータ操作の唯一の窓口。Web UI も CLI も同じ API を通る
2. ポータルの責務は**保管と入力**。出力・加工は外部に委ねる（CLI + jq、Claude Code、bot等）
3. **入力は徹底的に作り込む**。思いついた瞬間に書き始められる速度、最小タイプ量のCLI、モバイルでの快適さ
4. **出力（API）も徹底的に作り込む**。一貫したJSON形式、全エンドポイント共通のフィルタ・ソート・ページネーション。機械可読性を最優先

| アクセス元 | 認証方式 |
|-----------|---------|
| ブラウザ（Web UI） | ForwardAuth Cookie |
| CLI（ローカル/SSH） | API キー（Bearer トークン） |
| bot（Slack等） | API キー（Bearer トークン） |

**CLI の使用イメージ**:

```bash
# メモ
portal memo add "Flask P5テスト結果: 合格" --tag flask,教育
portal memo search "Flask 設計"
portal memo list --tag 教育

# ToDo
portal todo add "DD_00レビュー" --due 2026-03-15
portal todo list --status open
portal todo done <id>

# カレンダー
portal cal add "教育D1開始" --date 2026-03-12
portal cal list --week

# 個人DB
portal db create 血圧 --cols "日付:date,上:number,下:number,脈拍:number,メモ:text"
portal db insert 血圧 --日付 2026-03-12 --上 128 --下 82 --脈拍 72
portal db query 血圧 --sort 日付:desc --limit 10

# 横断検索
portal search "Flask"
```

### メモ機能の要件（Obsidian 置き換え）

Obsidian の全機能再現ではなく、実際に使っている機能だけを移植:

| Obsidian の機能 | ポータルでの対応 | 備考 |
|----------------|-----------------|------|
| Markdown 編集 | Markdown 対応エディタ（プレビュー付き） | ライブラリは後で選定 |
| タグ | タグ CRUD + タグ絞り込み | `#flask` `#教育` 等 |
| 検索 | SQLite FTS5 全文検索 | タイトル + 本文 |
| フォルダ構造 | 不要（タグで代替） | Obsidian でも実質タグ中心なら |
| 双方向リンク | 不要（初期スコープ外） | 必要になったら後で追加 |
| グラフビュー | 不要 | |

### 個人DB → メモに統合（テーブル型メモ）

血圧・読書記録など**構造化データ**は、独立機能ではなく**メモの一種（type="table"）** として統合する。
計算・リレーション等の処理を持たないなら、メモと同じCRUD・検索・タグの仕組みに乗せられる。

**データモデル（memosテーブル1つで統一）**:

```
memos: id, user_id, title, type, body, tags, created_at, updated_at

type = "markdown" → body に Markdown テキスト
type = "table"    → body に JSONB { columns: [...], rows: [...] }
```

**メリット**:
- 同じREST API・同じCLIサブコマンドで扱える
- FTS5全文検索がMarkdownメモもテーブルデータも横断する
- タグも共通（`#log/血圧` でテーブル型メモを絞り込み）
- Web UIだけ type に応じてエディタ/テーブルビューを出し分ける

**CLIイメージ**:

```bash
# テーブル型メモの作成
portal memo add "血圧" --type table --cols "日付:date,上:number,下:number,脈拍:number"

# テーブルに行追加
portal memo row add 血圧 --日付 2026-03-12 --上 128 --下 82 --脈拍 72

# 検索は型を区別しない
portal memo search "血圧"
portal memo list --tag log/血圧
```

**スコープ外**: リレーション、ロールアップ、数式、ビュー切替（Notion的な高度機能）

### フェーズ分け

| Phase | 内容 | 備考 |
|-------|------|------|
| P1 | ツールランチャー（カード表示 + 別タブ遷移）+ REST API 基盤 | API キー認証含む |
| P2 | メモ（Markdown + テーブル型 CRUD + タグ + 全文検索）+ CLI | Obsidian 置き換え + 個人DB |
| P3 | ToDo（CRUD + 期限 + 完了/未完了フィルタ）+ CLI | |
| P4 | 個人カレンダー（CRUD + 月/週表示）+ CLI | |

※ メモを ToDo より先にしたのは、Obsidian 置き換えとして日常的なインパクトが最大のため

### 未決定事項

- [ ] PostgreSQL スキーマ設計（メモ・ToDo・カレンダー・タグ・ユーザー）
- [ ] Markdown エディタライブラリ選定（CodeMirror / Monaco / SimpleMDE 等）
- [ ] CLI のサブコマンド設計（cobra）
- [ ] UIデザイン（WebSSH のダークテーマと統一するか）
- [ ] MD Task Vault の構想との統合（ToDo + メモ = ほぼ MD Task Vault）
- [ ] 複数ユーザー対応の設計（user_id、権限管理、管理者指定）
- [ ] 日本語全文検索 extension の選定（pg_bigm vs PGroonga）

### 参考: Memos（usememos/memos）

全部自作する方針だが、メモ機能の先行OSSとして Memos の設計を参考にする。

**Memos の概要**: Go + React 製、SQLite デフォルト（PostgreSQL対応あり）、REST + gRPC API、GitHub 57K+ stars

**参考にすべき点**:
- データモデル（メモの構造、タグの持ち方）
- API 設計（REST エンドポイントの粒度、フィルタ・ページネーション）
- UX 判断（即入力→保存の摩擦ゼロ設計、タイムライン型表示）
- Safari IME ハンドリング（WebSSH と同じ課題を解決済み）

**Memos にあって構想にないもの（取り込み検討）**:
- Webhook（メモ変更時の外部通知）
- RSS フィード（公開メモの配信）

**Memos に足りなくて構想にあるもの（差別化ポイント）**:
- ToDo・カレンダーの本格実装
- ツールランチャー（ポータル機能）
- テーブル型メモ（個人DB）
- 横断検索（メモ・ToDo・カレンダーをまたぐ）
- CLI ファーストの設計（入出力の作り込み）
- ForwardAuth 前提の統一認証基盤連携

---

## WebSSH 改修事項

### 認証の変更

| 項目 | 現状 | 変更後 |
|------|------|--------|
| 認証方式 | Basic 認証（環境変数） | ForwardAuth（統一認証基盤に委譲）|
| auth.js | Basic 認証ミドルウェア | 削除 or ForwardAuth ヘッダー検証に置換 |
| WebSocket 認証 | HTTP Upgrade 時に Basic 認証チェック | ForwardAuth Cookie ベースに変更 |
| .env | WEBSSH_USER / WEBSSH_PASS | 不要になる |

### 認証基盤と独立した改修（やる場合）

WebSSH に直接セキュリティを入れる場合（ブックマーク直アクセス対策）:

- セッション認証 + TOTP を WebSSH 自体に実装
- ForwardAuth と併用可（多層防御）
- 認証基盤がダウンしても WebSSH 単体で守れる

### その他（将来検討）

- セッション再接続（切断後の自動復帰）
- 複数タブ（同時複数SSHセッション）
- ntfy 連携（Claude Code のタスク完了通知）

---

## デプロイ構成（Docker Compose）

```yaml
# 構成イメージ（実装時に詳細化）
services:
  traefik:
    image: traefik:v3
    # TLS + ForwardAuth ミドルウェア定義

  auth:
    # traefik-forward-auth or Authelia
    # Google OAuth 設定

  postgres:
    image: postgres:16-alpine
    volumes:
      - pgdata:/var/lib/postgresql/data

  portal:
    build: ./portal  # Go バイナリ（scratch ベース）
    depends_on:
      - postgres
    labels:
      - "traefik.http.routers.portal.middlewares=forwardauth"

  webssh:
    build: ./webssh
    labels:
      - "traefik.http.routers.webssh.middlewares=forwardauth"
    # Tailscale 経由で自宅 sshd に接続

  zipline:
    # ForwardAuth 適用するか要判断
```

---

## 作業順序（案）

1. **認証基盤の選定**（traefik-forward-auth vs Authelia）
2. **VPS に Traefik + 認証サービスをデプロイ**
3. **WebSSH を VPS に移設**（Basic 認証削除 → ForwardAuth 適用）
4. **ポータル P1 をデプロイ**（ツールランチャー + REST API 基盤）
5. **Zipline の認証方針を決定**
6. **ポータル P2**（メモ: Markdown CRUD + タグ + 全文検索 + CLI）← Obsidian 置き換えの核
7. **ポータル P3〜P4**（ToDo → カレンダー + CLI）
8. **WebSSH の追加改修**（セッション再接続、複数タブ等）

---

## 未決定事項

- [ ] 認証基盤の選定（Google OAuth のみ vs 追加MFA）
- [ ] ポータルの詳細設計（画面・API・データモデル）
- [ ] Zipline を統一認証に含めるか
- [ ] WebSSH 自体にも独自認証を持たせるか（多層防御）
- [ ] VPS のスペック・プロバイダー（KAGOYA VPS 既存？追加？）
- [ ] ドメイン構成（サブドメイン設計）

---

## OSS化について

- **OSS は目指す。ただし個人ツールとして先に開発する**
- 自分が使えないものを公開しても意味がない。まず自分が日常的に使い倒せる状態にする
- 既存OSSにない独自ポジション:「セルフホスト個人グループウェア + Web UI/CLI両対応 + REST API ファースト」
- OSS化を閉ざさないために最低限意識すること:
  - 個人設定のハードコードを避ける（環境変数 or 設定ファイルに寄せる）
  - README / CLAUDE.md をちゃんと書く（WebSSH と同様）

---

## リポジトリ情報

| 項目 | 内容 |
|------|------|
| オーナー | krtw00（個人アカウント） |
| リポジトリ名 | **konbu** |
| ライセンス | MIT |
| ドキュメント言語 | 日本語 |
| 技術スタック | Go / PostgreSQL / Docker |
| 設計ドキュメント | WebSSH と同じ Templarc 形式（docs/） |
| CI/CD | 未定（後で検討） |

### プロダクト名: konbu

- GitHub / npm ともに空き確認済み（2026-03-12 時点）
- 覚えやすく、CLIで打ちやすい（5文字）
- 和テイストでユニーク
- CLI: `konbu memo add`, `konbu todo list`, `konbu search` 等
