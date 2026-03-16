# #24 検索機能の大幅強化

## 概要
- **目的**: 検索対象をタグ・ブックマークに拡張し、曖昧検索・専用ページ・ホーム検索バーを追加する
- **スコープ**: バックエンド API 拡張 → フロントエンド検索ページ新設 → ホーム検索バー追加
- **前提条件**: pg_trgm 拡張は導入済み（0005マイグレーション）、GINインデックスも既存
- **制約**: Ctrl+K コマンドパレットは残す（クイックアクセス用）
- **方針**: 検索クエリは既存の `internal/repository/search.go` と同様に手書きSQL で実装（sqlc は使用しない）

## 要件

### 機能要件
- **検索対象の拡張**
  - `q` パラメータのキーワードが、タグ名にも ILIKE ヒットする（タグ経由で紐づくメモ・ToDo・イベントを返す）
  - ツール（ブックマーク）を名前・URL でキーワード検索できる（結果に `type: "tool"` として含まれる）
- **曖昧検索**
  - pg_trgm の `similarity()` で曖昧マッチし「もしかして」セクションを表示
- **専用検索ページ**（サイドバーに追加）
  - フィルタ: 種類（メモ/ToDo/イベント/ブックマーク）、タグ（完全一致で絞り込み）、日付範囲
  - ページネーション（offset/limit）
- **ホーム検索バー**: ホームページ上部に配置、入力で検索ページに遷移
- 既存の Ctrl+K コマンドパレットはそのまま維持

> **タグ検索 vs タグフィルタの区別**
> - **タグ検索**（`q` パラメータ）: キーワードがタグ名に部分一致 → そのタグが付いたアイテムも結果に含める
> - **タグフィルタ**（`tag` パラメータ）: 指定タグが付いたアイテムのみに結果を絞り込む（完全一致）

### 非機能要件
- 検索レスポンス 200ms 以内（既存インデックス活用）

## 設計

### API 変更

#### `GET /api/v1/search` の拡張

現在のパラメータ: `q`, `limit`

追加パラメータ:

| パラメータ | 型 | デフォルト | 説明 |
|---|---|---|---|
| `type` | string | (全種類) | `memo`, `todo`, `event`, `tool` のカンマ区切り |
| `tag` | string | — | タグ名で絞り込み |
| `from` | string | — | 日付範囲の開始（RFC3339）。対象: memo→`created_at`, todo→`due_date`, event→`start_at` |
| `to` | string | — | 日付範囲の終了（RFC3339）。同上。tool には日付フィルタ不適用 |
| `offset` | int | 0 | ページネーション用オフセット |

レスポンス拡張:

```json
{
  "data": [
    {
      "type": "memo",
      "id": "...",
      "title": "...",
      "snippet": "...",
      "tags": ["tag1", "tag2"],
      "updated_at": "..."
    }
  ],
  "total": 42,
  "suggestions": [
    {
      "type": "memo",
      "id": "...",
      "title": "...",
      "snippet": "...",
      "tags": [],
      "updated_at": "...",
      "similarity": 0.35
    }
  ]
}
```

- `data`: 完全一致・部分一致（ILIKE）の結果
- `total`: フィルタ条件に合致する全件数（ページネーション用）
- `suggestions`: `similarity() >= 0.3` かつ `data` に含まれない曖昧マッチ結果（最大5件）

### DB クエリ追加

#### タグ経由の検索
```sql
-- タグ名にマッチするメモを検索
SELECT DISTINCT m.id, m.title, m.content, m.created_at, m.updated_at
FROM memos m
JOIN memo_tags mt ON mt.memo_id = m.id
JOIN tags t ON t.id = mt.tag_id
WHERE m.user_id = $1 AND m.deleted_at IS NULL AND t.deleted_at IS NULL
  AND t.name ILIKE $2
ORDER BY m.updated_at DESC
LIMIT $3 OFFSET $4;
```

同様に `todo_tags`, `calendar_event_tags` にも適用。

#### ツール検索
```sql
SELECT id, name, url, icon, created_at
FROM tools
WHERE user_id = $1 AND deleted_at IS NULL
  AND (name ILIKE $2 OR url ILIKE $2)
ORDER BY sort_order
LIMIT $3 OFFSET $4;
```

> **tools テーブルの注意点**:
> - `updated_at` カラムが存在しない → SearchResult では `created_at` を `updated_at` として返す
> - trgm インデックスが未設定 → `idx_tools_name_trgm`, `idx_tools_url_trgm` を追加するマイグレーションが必要

#### 曖昧検索（suggestions）
```sql
SELECT id, title, similarity(title, $2) AS sim
FROM memos
WHERE user_id = $1 AND deleted_at IS NULL
  AND similarity(title, $2) > 0.3
  AND id NOT IN (/* 通常検索の結果ID */)
ORDER BY sim DESC
LIMIT 5;
```

### フロントエンド

#### Page 型に `search` を追加
```typescript
type Page = 'home' | 'memos' | ... | 'search'
```

#### SearchPage コンポーネント
- 検索バー（上部固定）
- フィルタサイドバー（種類チェックボックス、タグ選択、日付範囲）
- 結果リスト（アイコン + タイプラベル + タイトル + スニペット + タグ + 日時）
- 「もしかして」セクション（suggestions が存在する場合）
- ページネーション（前へ/次へ）
- モバイルではフィルタをドロワーに格納

#### サイドバー
- `navItems` に `{ page: 'search', icon: Search, labelKey: 'nav.search' }` を追加
- 既存のサイドバー下部の検索ボタン（CommandPalette 起動用）はそのまま維持

#### ホーム検索バー
- `HomePage` 上部に検索入力欄を配置
- 入力してEnter → `setPage('search')` + クエリを引き渡し
- Zustand ストアに `searchQuery` 状態を追加

#### API クライアント拡張
```typescript
search: (params: {
  q: string
  limit?: number
  offset?: number
  type?: string
  tag?: string
  from?: string
  to?: string
}) => request('GET', `/search?${new URLSearchParams(...)}`)
```

#### SearchResult 型拡張
```typescript
export interface SearchResult {
  type: 'memo' | 'todo' | 'event' | 'tool'
  id: string
  title: string
  snippet: string
  tags: string[]
  updated_at: string
  similarity?: number  // suggestions のみ
}

export interface SearchResponse {
  data: SearchResult[]
  total: number
  suggestions: SearchResult[]
}
```

### 統合ロジック（service 層）

検索は以下の順序で実行し、Go 側でマージする（UNION ALL は型が異なるため不可）:

1. **通常検索**（ILIKE）: memos, todos, events のタイトル・本文を検索
2. **タグ経由検索**（ILIKE）: タグ名にヒット → 紐づくアイテムを取得
3. **ツール検索**（ILIKE）: tools の名前・URL を検索
4. **マージ**: 1 と 2 の結果を `id` で重複排除（map[uuid.UUID]struct{} で管理）、3 を追加
5. **ソート**: `updated_at` 降順（tools は `created_at` を使用）
6. **ページネーション**: Go 側で全件マージ後に offset/limit でスライス
7. **曖昧検索**: `similarity()` で data に含まれない結果を最大5件取得

**total の算出**: 各テーブルの COUNT クエリを並列実行し合算。タグ経由の重複は概算値として許容（厳密な重複排除は UNION で別途取るとコスト高）。

### データフロー

```
ホーム検索バー / 検索ページ入力
  → setSearchQuery(q) + setPage('search')
  → SearchPage mount → API呼び出し
  → GET /api/v1/search?q=...&type=...&tag=...&offset=...
  → handler: パラメータパース
  → service:
      1. 通常 ILIKE 検索（memo/todo/event）
      2. タグ名 ILIKE → 紐づくアイテム取得
      3. ツール ILIKE 検索
      4. Go 側で重複排除・マージ・ソート
      5. offset/limit でスライス
      6. similarity() で suggestions 取得
  → レスポンス: { data, total, suggestions }
  → SearchPage: 結果表示 + もしかしてセクション + ページネーション
```

## タスク分解

### DB マイグレーション
- [ ] tools テーブルに trgm インデックス追加（`idx_tools_name_trgm`, `idx_tools_url_trgm`）

### バックエンド
- [ ] repository: タグ経由検索クエリ追加（memo/todo/event）— 手書きSQL
- [ ] repository: ツール検索クエリ追加 — 手書きSQL
- [ ] repository: 曖昧検索クエリ追加（similarity）— 手書きSQL
- [ ] repository: 各検索のカウントクエリ追加（total 用）— 手書きSQL
- [ ] model: SearchResult に tags フィールド追加、SearchResponse 型追加
- [ ] service: 検索パラメータ拡張（type, tag, from, to, offset）
- [ ] service: タグ検索・ツール検索の統合・重複排除ロジック
- [ ] service: 曖昧検索の実行と suggestions 生成
- [ ] handler: クエリパラメータのパース拡張
- [ ] handler: レスポンス形式を SearchResponse に変更

### CLI / MCP Server
- [ ] CLI: `konbu search` の出力を新レスポンス形式に対応（suggestions 表示追加）
- [ ] MCP Server: `search` ツールの戻り値を新形式に対応

### フロントエンド
- [ ] 型定義: SearchResult / SearchResponse の拡張
- [ ] API クライアント: search 関数のパラメータ拡張
- [ ] Zustand: Page 型に `search` 追加、`searchQuery` 状態追加
- [ ] SearchPage: 新規作成（検索バー + フィルタ + 結果 + もしかして + ページネーション）
- [ ] サイドバー: 検索ページへのナビ追加
- [ ] ホーム: 検索バー追加（Enter で SearchPage 遷移）
- [ ] i18n: 検索ページ用の翻訳キー追加（en/ja）
- [ ] CommandPalette: SearchResult 型変更への対応（レスポンスが `{ data }` ラッパーに変わるため）

## リスク・懸念事項
- **similarity() のパフォーマンス**: GIN インデックスは ILIKE には効くが similarity() には効かない。件数が多い場合は GiST インデックス（`gist_trgm_ops`）の追加を検討
- **レスポンス形式の破壊的変更**: 現在の配列レスポンスを `{ data, total, suggestions }` に変更するため、CommandPalette・CLI・MCP Server の3箇所の対応が必要
- **total の概算**: Go 側マージで重複排除するため、COUNT 合算の total は概算値になる。UX 上は「約 N 件」表示で許容

## 未決事項
- 検索結果のスニペットにキーワードハイライトを入れるか（フロントのみで対応可能）
- モバイルでのフィルタUIの詳細レイアウト（#22 と連携）
