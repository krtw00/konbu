# #25 テーブル型メモ — CSV的簡易テーブル機能

## 概要
- **目的**: 血圧記録・家計簿・在庫管理など、表形式で管理したいデータを扱える機能を提供する
- **スコープ**: テーブル型メモの完全実装（バックエンド API → フロントエンド UI → CLI）
- **前提条件**: DB スキーマ（`memo_rows`, `memos.table_columns`）は作成済み。バックエンド〜フロントの実装がほぼゼロ
- **制約**: 1ユーザーあたり最大1000テーブル × 数百行のヘビーユースケースに対応

## 要件

### 機能要件

#### 列管理
- 右端に常に空列を1つ表示 — ヘッダーに列名を入力すると列が追加され、次の空列が出現（行の追加と同じ UX）
- 列の削除・名前変更（列ヘッダーメニュー）
- 列に型なし（全テキスト）— 自由入力、ソート時は数値判別で自動切り替え
- 列の並び替え（左右ドラッグ or 上下操作）
- `table_columns` に列定義を JSONB で保存: `[{"id": "col_xxx", "name": "日付"}, {"id": "col_yyy", "name": "最高"}]`

#### 行管理
- 入力済み行 + 最後尾に空行1つを常時表示（入力すると自動で行追加、次の空行が出現）
- 行の削除（チェックボックス選択 → 一括削除、または行メニュー）
- 行のソート（列ヘッダークリックで昇順/降順トグル）
- `row_data` に JSONB で保存: `{"col_xxx": "3/17", "col_yyy": "120"}`

#### セル編集
- セルクリックで直接インライン編集（input 表示）
- Tab で右のセルへ、Enter で下のセルへ移動
- Escape で編集キャンセル

#### CSV インポート/エクスポート
- CSV エクスポート: 現在のテーブルを CSV ファイルでダウンロード
- CSV インポート: CSV ファイルをアップロード → 列自動検出 → 行を一括追加
- ヘッダー行の自動認識

#### 一覧表示
- メモ一覧で TBL バッジ表示（既存）
- テーブル型メモの作成ボタン（メモ一覧ヘッダー）

### 非機能要件
- **ページネーション**: 行は API で 100 件ずつ取得（無限スクロール）
- **パフォーマンス**: 1000 テーブル × 数百行でもレスポンス 200ms 以内
- **自動保存**: セル編集後 1 秒のデバウンスで自動保存（行単位）

## 設計

### データモデル

#### table_columns（memos.table_columns JSONB）
```json
[
  {"id": "col_a1b2", "name": "日付"},
  {"id": "col_c3d4", "name": "最高"},
  {"id": "col_e5f6", "name": "最低"},
  {"id": "col_g7h8", "name": "メモ"}
]
```

- `id`: フロントで生成する短い一意 ID（`col_` + nanoid 8文字）
- 列の順序は配列のインデックスで決定

#### memo_rows（既存テーブル）
```
id         UUID PK
memo_id    UUID FK → memos
row_data   JSONB  {"col_a1b2": "3/17", "col_c3d4": "120", ...}
sort_order INTEGER
created_at TIMESTAMPTZ
deleted_at TIMESTAMPTZ
```

### API エンドポイント

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/api/v1/memos/:id/rows?limit=100&offset=0&sort=col_xxx&order=asc` | 行一覧（ページネーション） |
| POST | `/api/v1/memos/:id/rows` | 行追加 `{"row_data": {...}}` |
| PUT | `/api/v1/memos/:id/rows/:rowId` | 行更新 `{"row_data": {...}}` |
| DELETE | `/api/v1/memos/:id/rows/:rowId` | 行削除（論理削除） |
| POST | `/api/v1/memos/:id/rows/batch` | 行一括追加（CSV インポート用） `{"rows": [{...}, ...]}` |
| GET | `/api/v1/memos/:id/rows/export` | CSV エクスポート（Content-Type: text/csv） |

列定義の更新は既存の `PUT /api/v1/memos/:id` で `table_columns` を送信。

### レスポンス形式

```json
// GET /api/v1/memos/:id/rows
{
  "data": [
    {"id": "uuid", "row_data": {"col_a1b2": "3/17", "col_c3d4": "120"}, "sort_order": 0, "created_at": "..."}
  ],
  "total": 245,
  "limit": 100,
  "offset": 0
}
```

### ソート

- API 側で完結（ページネーションと整合させるため DB ソート必須）
- パラメータ: `sort` に列 ID、`order` に `asc` / `desc`
- SQL: `ORDER BY (row_data->>$sort) ASC` — デフォルトはテキストソート
- 数値ソート: `ORDER BY (CASE WHEN row_data->>$sort ~ '^-?[0-9.]+$' THEN (row_data->>$sort)::numeric ELSE NULL END) ASC NULLS LAST` — 正規表現で数値判定し安全に CAST

### 認可

- 行操作の全エンドポイントで、対象メモが `user_id = 現在のユーザー` であることを確認（既存の memo handler と同じパターン）

### batch API 制限

- `POST /rows/batch` は 1 リクエストあたり最大 500 行に制限（CSV インポートの実用上限）

### メモ詳細レスポンス

- `GET /memos/:id` でテーブル型の場合、`table_columns` と `row_count` を含める（rows 自体は `/rows` エンドポイントで取得）

### フロントエンド

#### MemoEditPage の分岐
```
memo.type === 'markdown' → 既存の Monaco Editor
memo.type === 'table'    → TableEditor コンポーネント
```

#### TableEditor コンポーネント構成
```
TableEditor
├── ColumnHeader（列ヘッダー：列名表示、ソートボタン、列メニュー）
├── TableRow × N（データ行：セル群）
│   └── TableCell（インライン編集可能なセル）
├── NewRow（最後尾の空行：入力で自動追加）
├── NewColumn（右端の空列ヘッダー：入力で自動列追加）
└── TableToolbar（CSV インポート/エクスポート、行一括削除）
```

#### 操作フロー

**セル編集**:
```
セルクリック → input 表示 → 入力 → 1秒デバウンス → PUT /rows/:id → 自動保存
Tab → 右のセルへ / Enter → 下のセルへ / Escape → キャンセル
```

**新規行追加**:
```
最後尾の空行のセルに入力 → POST /rows → 新行追加 → 次の空行が出現
```

**列追加**:
```
右端の空列ヘッダーに列名を入力 → Enter → PUT /memos/:id (table_columns更新) → 新列出現 + 次の空列が右端に
```

**CSV インポート**:
```
ツールバー「インポート」→ ファイル選択 → ヘッダー行プレビュー
→ 列マッピング確認 → POST /rows/batch → invalidateCache
```

### パフォーマンス対策

- **行のページネーション**: 初回 100 件取得、スクロールで追加読み込み
- **バーチャルスクロール**: 表示行のみ DOM に描画（数百行対応）— 初期はシンプルなページネーション、必要に応じて導入
- **行単位の保存**: セル編集は行単位で PUT（テーブル全体を送らない）
- **インデックス活用**: `idx_memo_rows_memo_id` で memo_id + sort_order のクエリ高速化
- **一覧の軽量化**: メモ一覧取得時に rows は含めない（行数のみカウント）

## タスク分解

### Phase 1: バックエンド
- [ ] model: MemoRow 構造体、ListRowsParams 型定義
- [ ] repository: memo_rows CRUD（Create, List, Update, SoftDelete, BatchCreate）
- [ ] repository: CSV エクスポートクエリ
- [ ] service: MemoRowService（CRUD + バッチ + エクスポート）
- [ ] handler: `/memos/:id/rows` エンドポイント群
- [ ] handler: CSV エクスポートレスポンス（text/csv）
- [ ] ルーティング登録（cmd/server/main.go）

### Phase 2: フロントエンド基礎
- [ ] 型定義: MemoRow, TableColumn interfaces
- [ ] API クライアント: rows CRUD + batch + export メソッド
- [ ] MemoEditPage: type 分岐（markdown → Editor、table → TableEditor）
- [ ] MemosPage: テーブル型メモ作成ボタン

### Phase 3: TableEditor UI
- [ ] TableEditor: 基本レイアウト（ヘッダー + 行 + 空行）
- [ ] TableCell: インライン編集（クリック→input、Tab/Enter/Escape）
- [ ] ColumnHeader: ソートボタン（昇順/降順トグル）
- [ ] NewRow: 入力で自動行追加
- [ ] NewColumn: 右端空列ヘッダーで入力→列追加
- [ ] 列の削除・名前変更（列ヘッダーメニュー）
- [ ] 行の削除（行メニュー or チェックボックス一括）
- [ ] 自動保存（セル編集 → 1秒デバウンス → PUT）

### Phase 4: CSV 対応
- [ ] CSV エクスポートボタン（ツールバー）
- [ ] CSV インポートダイアログ（ファイル選択 → プレビュー → 一括追加）

### Phase 5: CLI / MCP
- [ ] CLI: `konbu memo` でテーブル型対応（行表示、行追加）
- [ ] MCP: テーブル行操作ツール追加

### Phase 6: i18n
- [ ] en.json / ja.json にテーブル関連の翻訳キー追加

## リスク・懸念事項

- **大量行のパフォーマンス**: 数百行 × 1000テーブルは memo_rows テーブルが数十万行に。`idx_memo_rows_memo_id` インデックスでカバーできるが、JSONB 内ソートは重い可能性。必要に応じて `row_data->>key` の GIN インデックス追加
- **JSONB ソートの限界**: PostgreSQL の `row_data->>key` はテキスト型。数値ソートには `CAST` が必要で、非数値データが混在するとエラー → `NULLIF` + `regexp` でガード
- **バーチャルスクロール**: 初期実装ではシンプルなページネーション。UX が悪ければ後から `@tanstack/virtual` 等を導入
- **列 ID の衝突**: nanoid 8文字で実用上問題ないが、念のため既存列 ID との重複チェックを入れる

## 未決事項

- 列の並び替え UI（ドラッグ or 上下ボタン）— 初期は上下ボタンで十分かも
- 行の並び替え（手動 drag or sort_order 固定）— ソートで十分か、手動並び替えも必要か
- セル内の改行対応（textarea にするか、1行固定か）
- 将来的な列集計機能（合計・平均・最大・最小の表示）の予約設計
