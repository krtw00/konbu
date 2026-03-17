---
depends_on: []
tags: [navigation, index]
ai_summary: "konbu設計ドキュメントへのナビゲーションハブ"
---

# 設計ドキュメントインデックス

> **Status**: Active | 最終更新: 2026-03-14

本ドキュメントは、konbuの設計ドキュメント全体のナビゲーションを提供する。

---

## ドキュメント構造

```mermaid
flowchart LR
    A[01-overview<br/>全体像] --> B[02-architecture<br/>設計]
    B --> C[03-details<br/>詳細]
    B --> D[04-hosting<br/>提供形態]
```

| レベル | 目的 | 対象読者 |
|--------|------|----------|
| **01-overview** | 何を作るか、なぜ作るか | 初見・思い出し用 |
| **02-architecture** | どう構成するか | 設計理解 |
| **03-details** | 具体的な仕様 | 実装時参照 |
| **04-hosting** | どう提供するか | 運営・インフラ |

---

## ドキュメント一覧

### 01 - Overview（全体像）

| ドキュメント | 説明 |
|--------------|------|
| [summary.md](./01-overview/summary.md) | プロジェクト概要（1枚で全体把握） |
| [goals.md](./01-overview/goals.md) | 目的・解決する課題 |
| [principles.md](./01-overview/principles.md) | 設計原則・判断基準 |
| [scope.md](./01-overview/scope.md) | スコープ・対象外 |

### 02 - Architecture（設計）

| ドキュメント | 説明 |
|--------------|------|
| [context.md](./02-architecture/context.md) | システム境界・外部連携 |
| [structure.md](./02-architecture/structure.md) | 主要コンポーネント構成 |
| [tech-stack.md](./02-architecture/tech-stack.md) | 技術スタック |

### 03 - Details（詳細）

| ドキュメント | 説明 |
|--------------|------|
| [data-model.md](./03-details/data-model.md) | データモデル・ER図 |
| [api.md](./03-details/api.md) | API設計 |
| [ui.md](./03-details/ui.md) | UI設計 |
| [flows.md](./03-details/flows.md) | 主要フロー・シーケンス |

### 04 - Hosting（提供形態）

| ドキュメント | 説明 |
|--------------|------|
| [hosting.md](./04-hosting/hosting.md) | 提供形態・料金モデル・インフラ構成 |

### 参照

| ドキュメント | 説明 |
|--------------|------|
| [schema.sql](./schema.sql) | DDL参照用（全マイグレーション統合版） |

---

## 読み方ガイド

### 初めて読む場合

1. [summary.md](./01-overview/summary.md) - プロジェクト概要を把握
2. [goals.md](./01-overview/goals.md) - 目的を理解
3. [principles.md](./01-overview/principles.md) - 設計原則を確認
4. [context.md](./02-architecture/context.md) - システム境界を確認

### 設計を理解したい場合

1. [structure.md](./02-architecture/structure.md) - コンポーネント構成
2. [tech-stack.md](./02-architecture/tech-stack.md) - 技術選定理由

### 実装時に参照する場合

1. [data-model.md](./03-details/data-model.md) - データ構造
2. [api.md](./03-details/api.md) - APIエンドポイント仕様
3. [flows.md](./03-details/flows.md) - 処理フロー
