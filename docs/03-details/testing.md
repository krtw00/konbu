---
depends_on:
  - ./ui.md
  - ./api.md
tags: [details, testing, quality]
ai_summary: "konbuのテスト責務と unit / integration / E2E の割り当て"
---

# テスト戦略

> **Status**: Active | 最終更新: 2026-03-18

本ドキュメントは、konbu のテストを `unit / integration / E2E` にどう割り当てるかを定義する。

---

## 方針

konbu の回帰は、純粋ロジックよりも `画面 + 状態 + API + routing` の境界で起きやすい。

そのため、同じ挙動を全レイヤーで重複して保証するのではなく、責務ごとに最適なテスト層を決める。

- `Unit`
  純粋ロジック、変換、整形、入力正規化を保証する
- `Integration`
  コンポーネント、フォーム、ダイアログ、非同期状態遷移を保証する
- `E2E`
  実際のユーザーフローが成立することを保証する

---

## レイヤー定義

### Unit

対象:

- 日付変換
- 公開 URL 組み立て
- config / validation
- 検索整形や AI 補助の純粋ロジック

ルール:

- 依存をできるだけ持たない
- I/O や DOM に寄せない
- 失敗時に原因がすぐわかる粒度で書く

### Integration

対象:

- Login / Setup の submit lock
- 公開リンクダイアログ
- イベントフォームの状態遷移
- カレンダー管理ダイアログ

ルール:

- UI コンポーネントと store / API 呼び出しの協調を見る
- モックは API 境界やブラウザ API に限定する
- 文言そのものより、操作と状態遷移を重視する

### E2E

対象:

- 認証セットアップ
- 終日予定作成
- カレンダー名変更
- 共有リンク発行
- 公開カレンダー表示切り替え
- 左ナビ遷移でシェルが消えないこと
- フィードバック送信

ルール:

- ユーザーにとって重要な導線だけを固定する
- 細かい見た目や文言は E2E に寄せすぎない
- 不具合再現が取れたものは優先して回帰化する

---

## 責務マトリクス

| 機能 / 責務 | Unit | Integration | E2E |
|---|---|---|---|
| 日付変換・終日変換 | `web/frontend/src/lib/date.test.ts` | -- | 終日予定作成 |
| runtime URL 解決 | `web/frontend/src/lib/runtime.test.ts` | -- | 公開ページ到達 |
| Login / Setup 多重送信防止 | -- | `LoginPage.test.tsx`, `SetupPage.test.tsx` | 認証セットアップ |
| 共有リンク作成 UI | -- | `PublicShareDialog.test.tsx` | 公開リンク発行 |
| カレンダー操作 | -- | 追加対象 | `calendar.spec.ts` |
| ナビゲーション安定性 | -- | -- | `navigation.spec.ts` |
| フィードバック送信 | -- | 追加対象 | `feedback.spec.ts` |
| Go の純粋ロジック | `config_test.go`, `search_service_test.go`, `chat_service_test.go` | -- | -- |

---

## 現在のテスト基盤

Frontend:

- `Vitest + Testing Library`
- `Playwright`

Backend:

- `go test ./...`

CI:

- Go: `go vet`, `go build`, `go test`
- Frontend: `npm run lint`, `npm test`, `npm run build`

---

## 優先順

1. 不具合修正時に、同じ現象を最低 1 本の回帰テストへ落とす
2. 純粋ロジックは unit で吸収し、E2E に持ち込まない
3. 画面の壊れやすい導線は integration か E2E のどちらかで固定する
4. E2E は critical flow に限定し、件数より安定性を優先する

---

## 次に増やすもの

- カレンダー管理ダイアログの integration test
- 公開 memo / event / calendar の E2E
- 登録後の onboarding 導線の E2E
- backend の認可と公開機能の service test

---

## 実行コマンド

```bash
go test ./...
cd web/frontend && npm test
cd web/frontend && npm run test:e2e
```
