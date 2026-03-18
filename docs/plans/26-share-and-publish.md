---
depends_on:
  - ../01-overview/principles.md
  - ./public-shares.md
tags: [plan, publishing, sharing]
ai_summary: "konbuの共有リンクと公開機能をshare/publishに分離する計画"
---

# Share / Publish Rework

## 概要

- 目的: konbu 内の情報を外部に出す手段を `share` と `publish` に分離し、用途ごとに一貫した公開体験を作る
- 背景: 現状の `public_shares` は token ベースの閲覧リンクであり、外部共有には有効だが、公開・配信・ブログ代替としては弱い
- スコープ: `memo`, `event`, `calendar`, `todo` の公開設計、公開URL体系、公開ページ責務、公開管理UI
- 非スコープ: 公開ページ上での編集、コメント、チームコラボ、認証付き限定配布、課金

## 要件

### 機能要件

- 公開機能を次の2モードに分離する
  - `share`: token ベースの unlisted な閲覧リンク
  - `publish`: slug ベースの public な公開ページ
- `share` は既存 `public_shares` をベースに継続し、ログイン不要の read-only URL として維持する
- `publish` は検索や配信で使える恒久URLを持つ
- `memo` は publish の最優先対象とし、ブログ/公開メモとして成立する表示を持つ
- `event` と `calendar` は外部告知・予定表として publish 可能にする
- `todo` は原則 share 寄りとし、publish は後段または限定運用にする
- `tool` は通常の個人用リンク集として扱い、publish 初期スコープから外す
- 公開状態は少なくとも `private / unlisted / public` を表現できる
- 公開ページは OGP タイトル・説明・URL の制御を行える
- 公開ページはログイン不要で安定して閲覧できる
- 公開状態の変更は UI から分かるようにする

### 非機能要件

- 既存の share URL は壊さない
- 現行 API / CLI / UI との責務を明確にし、意味の異なる公開方法を混在させない
- 軽量性を優先し、公開機能のために通常利用の導線を重くしない
- SEO を狙いすぎず、まずは「人に見せられる公開物」として成立させる

## 設計

### 1. モデル

- `share`
  - 用途: 相手にだけ見せる、配信で一時共有する、認証なしで参照させる
  - URL: token ベース
  - 特徴: 一覧化しない、推測困難、OGP は最低限
- `publish`
  - 用途: ブログ代替、公開プロフィール、公開リンク集、イベント告知
  - URL: slug ベース
  - 特徴: 一覧化できる、恒久URL、OGP を整える、公開面のテーマを持てる

### 2. 情報種別ごとの位置づけ

- `memo`
  - publish の中心
  - 長文、ノート、記事、配信メモに使う
- `event`
  - 単体イベント告知として publish 可能
- `calendar`
  - 継続イベント表、配信予定表として publish 可能
- `todo`
  - 基本は share 寄り
  - public roadmap や check list 的用途が見えたら後で publish を検討する
- `tool`
  - 現時点では個人用リンク集の性格が強く、publish / share の優先対象から外す

### 3. データ/URL 方針

- 既存 `public_shares` は `share` 用として継続する
- `publish` 用には別の公開メタデータを持つ
  - 候補: `published_resources` テーブル
- `publish` で最低限持つ値
  - `resource_type`
  - `resource_id`
  - `slug`
  - `title`
  - `description`
  - `published_at`
  - `updated_at`
  - `visibility`
- URL 例
  - share: `/public/:token`
  - publish: `/@/:slug` または `/:type/:slug`
- 初期案としては型ごとの衝突回避がしやすい `/:type/:slug` を優先する

### 4. UI 方針

- 各編集画面の公開導線は `共有` と `公開` を分けて表示する
- `共有`
  - 既存の公開リンクダイアログを発展させる
- `公開`
  - slug、タイトル、概要、公開状態を管理する専用 UI を持つ
- memo では `公開プレビュー` を優先的に提供する
- calendar / event では配信用に見やすい公開ページを重視する

### 5. API / CLI 方針

- API
  - `share` と `publish` を別エンドポイントに分ける
  - 公開状態変更、slug 更新、公開ページ取得 API を追加する
- CLI
  - `share` は既存 `public` コマンド系を維持または `share` に改名検討
  - `publish` は slug 管理、公開/非公開切り替え、URL 取得を扱えるようにする
- AI
  - 将来的に「このメモを公開用に整える」「公開説明文を作る」支援につなげる

## フェーズ

### Phase 1: 概念整理

- `public_shares` を share として再定義する
- UI 文言を `共有` に寄せる
- publish 用のモデルと URL 方針を決める
  - auth-side metadata API: `/api/v1/publishes/:resourceType/:id`
  - public slug lookup API: `/api/v1/published/:resourceType/:slug`

### Phase 2: Memo Publish

- memo の publish API / UI / 公開ページを作る
- slug と OGP を実装する
- 公開一覧またはプロフィール導線の最小版を用意する

### Phase 3: Event / Calendar Publish

- event / calendar の公開ページを配信・告知向けに整える

### Phase 4: 仕上げ

- 公開トップ / プロフィール導線
- 公開ページデザインの共通化
- OGP / sitemap / アクセス計測の検討

## タスク分解

- [ ] `share` と `publish` の用語・責務を docs と UI で分離する
- [x] publish 用データモデルと URL 仕様を決める
- [ ] memo publish の API を実装する
- [ ] memo publish の編集 UI と preview を実装する
- [ ] memo publish ページの OGP と表示を整える
- [ ] event / calendar publish を実装する
- [ ] CLI に publish 操作を追加する
- [ ] 公開一覧またはプロフィール導線を追加する
- [ ] docs / README / LP を publish 機能に追従させる
- [ ] tool の public share / publish を仕様から外し、UI・docs・CLI・backend cleanup 方針を整理する

## リスク・懸念事項

- `share` と `publish` の概念が UI 上で混ざると分かりにくくなる
- memo を publish の中心にしないと、公開機能が用途不明の寄せ集めになりやすい
- `todo` の public 化はプライバシー事故につながりやすい
- slug 更新や resource 削除時の URL 互換をどう扱うかは早めに決める必要がある
- OGP や公開一覧を入れ始めると、SPA だけでは限界が出る可能性がある

## 未決事項

- publish metadata API は `/:type/:slug` ベースの解決を前提に `GET /api/v1/published/:resourceType/:slug` を採用済み
- 公開ページ URL は `/:type/:slug` か `/@/:slug` か
- 公開一覧はユーザー単位プロフィールを持つか、まずは単体ページだけにするか
- `todo` の publish を初期スコープに含めるか
- publish metadata を resource ごとに持つか、共通テーブルに集約するか
- SEO をどこまで狙うか
