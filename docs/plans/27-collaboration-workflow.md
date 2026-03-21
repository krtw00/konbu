---
depends_on:
  - ../00-index.md
  - ../01-overview/summary.md
  - ../../CONTRIBUTING.md
tags: [plan, collaboration, issues, project]
ai_summary: "konbu のコラボレーター向けに issue 設計と GitHub Project 運用を整理する計画"
---

# Collaboration Workflow

## 概要

- 目的: `konbu` にコラボレーターが入っても、Issue の粒度と GitHub Project 上の状態管理がぶれない運用を定義する
- スコープ: Issue 種別、ラベル設計、Project フィールド、トリアージ手順、初期バックログの切り方
- 前提条件: 既存の `bug`, `feature`, `question` テンプレートと PR テンプレートは維持する
- 制約: 個人向け OSS のまま運用コストを増やしすぎない。状態管理は GitHub 標準機能で完結させる

## 要件

### 機能要件

- 外部ユーザー向けの問い合わせと、コラボレーター向けの実装タスクを分離する
- 各 Issue で最低限 `type`, `area`, `priority` が判断できるようにする
- 大きな計画は親 Issue と子 Task に分解し、Project 上で進捗を追えるようにする
- コラボレーターが「次に着手してよいもの」を Project の `Ready` から選べるようにする
- PR は原則 1 Issue に対応させ、Project 側でも追跡できるようにする
- 既存の設計ドキュメントや `docs/plans/*.md` から、親 Issue を起こしやすい構成にする

### 非機能要件

- 初見コントリビューターでも 5 分以内に運用ルールを理解できること
- ラベル数は増やしすぎず、1 Issue あたり 3 個前後で意味が伝わること
- 状態管理はラベルではなく Project の field を中心にし、二重管理を避けること

## 設計

### 1. Issue の役割分担

- `Bug Report`
  - 想定: 不具合報告、回帰報告
  - Project 追加時の既定値: `Status=Inbox`, `Priority=P1`
- `Feature Request`
  - 想定: 新機能提案、既存機能改善
  - 実装前に maintainer が task 化する
- `Question / Inquiry`
  - 想定: 利用方法、セルフホスト、API の相談
  - 原則 Project に入れず、議論のみで完了させる
- `Task / Implementation Slice`
  - 想定: コラボレーターが着手する具体作業
  - 必須項目: 目的、スコープ、受け入れ条件、検証方法

### 2. ラベル設計

Issue には原則として次の 3 系統を付ける。

#### Type

- `type:bug`
- `type:feature`
- `type:task`
- `type:docs`
- `type:refactor`
- `type:question`

#### Area

- `area:memo`
- `area:todo`
- `area:calendar`
- `area:search`
- `area:tool`
- `area:ai`
- `area:web`
- `area:api`
- `area:cli`
- `area:db`
- `area:auth`
- `area:infra`
- `area:docs`

#### Priority

- `prio:P0` - 本番障害、セキュリティ、データ破損
- `prio:P1` - 次の作業候補に入れる重要項目
- `prio:P2` - 通常の改善・保守
- `prio:P3` - 後回しでよいアイデア、探索的改善

#### 補助ラベル

- `good first issue`
- `help wanted`
- `blocked`
- `needs-design`

### 3. GitHub Project 設計

1 つの Project を `konbu roadmap` として運用し、Issue / PR をまとめて扱う。

#### 推奨フィールド

| Field | 型 | 用途 |
|---|---|---|
| `Status` | Single select | `Inbox`, `Backlog`, `Ready`, `In Progress`, `In Review`, `Blocked`, `Done` |
| `Priority` | Single select | `P0`, `P1`, `P2`, `P3` |
| `Area` | Single select | ラベルの `area:*` と同じ分類 |
| `Size` | Single select | `XS`, `S`, `M`, `L`, `XL` |
| `Target` | Text or iteration | `v0.x`, `Search`, `Publishing` など |
| `Docs/Plan` | Text | 対応する `docs/plans/*.md` や設計 doc へのリンク |

#### 推奨ビュー

- `Triage`
  - 条件: `Status=Inbox`
  - 用途: 新規 Issue の仕分け
- `Ready`
  - 条件: `Status=Ready`
  - 用途: コラボレーターが次に拾う作業一覧
- `In Flight`
  - 条件: `Status in (In Progress, In Review, Blocked)`
  - 用途: 現在進行中の把握
- `Bugs`
  - 条件: `type:bug`
  - 用途: 不具合優先管理
- `Docs and DX`
  - 条件: `area:docs` or `area:infra`
  - 用途: コントリビューター体験の改善

### 4. トリアージ手順

1. 新規 Issue は maintainer が `Inbox` で確認する
2. 種別を決めて `type:*` と `area:*` を付与する
3. すぐ着手できる粒度なら `Ready` に移す
4. 大きいテーマは親 Issue にして、子 Task に分解してから `Ready` を作る
5. 着手時に assignee を付けて `In Progress` へ移す
6. PR 作成時に Issue をリンクし、`In Review` にする
7. マージ後に `Done` へ移し、必要なら README / docs の追従 Issue を別で切る

### 5. Issue 粒度の基準

- `XS`: ドキュメントのみ、typo、軽微な UI 修正
- `S`: 1 PR で完結する単機能修正、DB 変更なし
- `M`: API と UI の両方に変更が入るが、設計判断が少ない
- `L`: DB マイグレーション、複数レイヤー変更、レビュー観点が多い
- `XL`: そのままでは広すぎる。親 Issue にして分割する

原則として、コラボレーターに割り当てるのは `S` か `M` までにする。

### 6. 既存計画からの初期バックログ

次の `docs/plans` はそのまま親 Issue の元ネタとして使える。

- `#24 search enhancement`
  - 候補 Task: バックエンド検索拡張、検索ページ UI、ホーム検索バー、trgm インデックス
- `#25 table memo`
  - 候補 Task: データモデル、API、テーブル UI、エクスポート整備
- `#26 share / publish rework`
  - 候補 Task: 用語整理、publish API、memo publish UI、public page 整備

これに加えて、コラボレーション開始時点では次のメンテナンストラックも並行で持つ。

- `Contributor experience`
  - `CONTRIBUTING.md`、Issue template、PR テンプレート、docs 導線
- `Self-hosting polish`
  - Docker / Firebase / Fly.io / Cloudflare 関連の導線整理
- `Docs drift check`
  - README, `docs/00-index.md`, API / UI / schema の整合確認

## タスク分解

- [ ] GitHub 上に `type:*`, `area:*`, `prio:*` ラベルを作成する
- [ ] GitHub Project `konbu roadmap` を作成し、`Status`, `Priority`, `Area`, `Size`, `Target`, `Docs/Plan` フィールドを追加する
- [ ] `Task / Implementation Slice` の Issue form を追加する
- [ ] `CONTRIBUTING.md` に Issue / Project 運用ルールへの導線を追加する
- [ ] 既存の `docs/plans/24-26` を親 Issue 化する
- [ ] 親 Issue を `S` / `M` サイズの子 Task に分解する
- [ ] `good first issue` と `help wanted` を最低 3 件ずつ用意する

## リスク・懸念事項

- ラベルを増やしすぎると maintainer しか扱えなくなる
- `Status` をラベルと Project の両方で管理するとすぐに崩れる
- `XL` のまま協業を始めると、途中で責任範囲が曖昧になる
- バグ報告と実装タスクを同じ粒度で扱うと、Project がノイズで埋まりやすい

## 未決事項

- GitHub Project は organization 配下か repository 配下か
- `Target` をリリース単位で持つか、テーマ単位で持つか
- `Question / Inquiry` を Discussions に分離するか
- 将来的に Sponsors 向け要望を別トラックで扱うか
