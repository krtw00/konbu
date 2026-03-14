---
depends_on: []
tags: [overview, summary]
ai_summary: "konbuの1ページ概要 -- セルフホスト型個人ワークスペースの目的・主要機能・対象ユーザー・システム構成図"
---

# プロジェクト概要

> **Status**: Active | 最終更新: 2026-03-14

本ドキュメントは、konbu全体を1枚で把握するための概要を記載する。

---

## 一言で言うと

konbuは、メモ・ToDo・カレンダー・ツールランチャーを統合するパーソナルワークスペースである。セルフホスト（OSS・MIT）とクラウド版の2つの利用形態を提供する。

---

## 背景

| 項目 | 内容 |
|------|------|
| 現状の課題 | メモ・ToDo・カレンダー・ブックマークが別々のサービスに散らばっている。Notionは重く、Obsidianはローカル専用、Google Calendarはデータを手元に置けない |
| 解決アプローチ | 日常的に使う情報ツールを1つのアプリに統合し、REST APIを唯一の窓口として Web UI・CLI・外部連携すべてに対応する。セルフホストとクラウドの両方で利用可能 |

---

## 主要機能

| 機能 | 説明 |
|------|------|
| Memos | Markdown/テーブル型のメモ。CodeMirror 6エディタ、タグ付き |
| ToDo | 期限・タグ・メモ付きタスク管理。完了/未完了のステータス操作 |
| Calendar | 月表示カレンダー。終日/時間指定、繰り返し予定、iCalインポート |
| Tools | ブックマークランチャー。カテゴリ分類、favicon自動取得、ヘルスチェック |
| Cross-search | メモ・ToDo・予定をまたいだ横断全文検索（pg_bigmで日本語対応） |
| CLI | 全リソースのCRUD操作を備えたスタンドアロンCLIクライアント |
| Export/Import | JSON/Markdown ZIPエクスポート、iCalインポート |

---

## 対象ユーザー

| ユーザー種別 | 説明 | 主な利用シーン |
|--------------|------|----------------|
| クラウドユーザー | セットアップ不要で使いたい個人 | ブラウザからのメモ・タスク・予定管理 |
| セルフホスター | 自分のサーバーを持ち、データを手元に置きたい個人 | Docker等でのセルフホスト運用 |
| CLI/APIユーザー | ターミナルやスクリプトからデータにアクセスしたい開発者 | CLIでのメモ追加、AIエージェント連携 |

---

## 提供形態

| 形態 | ライセンス | 説明 |
|------|-----------|------|
| **Self-hosted** | OSS (MIT) | 全機能無料。Docker or ネイティブでセルフホスト |
| **Cloud** | SaaS | 無料で全機能利用可。GitHub Sponsors 支援者向けの追加特典あり |

## システム概観

```mermaid
flowchart TB
    subgraph Users[ユーザー]
        U1[Web UI<br/>React SPA]
        U2[CLI<br/>cobra/Go]
    end

    subgraph System[konbu]
        API[REST API Server<br/>Go / chi v5]
        DB[(PostgreSQL 16+<br/>pg_bigm)]
    end

    U1 -->|HTTP/JSON<br/>Session Cookie| API
    U2 -->|HTTP/JSON<br/>Bearer Token| API
    API -->|SQL| DB
```

---

## 関連ドキュメント

- [目的・解決する課題](./goals.md) - 課題一覧と成功基準の定義
- [スコープ・対象外](./scope.md) - 対象範囲とフェーズ分割
- [システム境界・外部連携](../02-architecture/context.md) - システム境界と外部システム定義
