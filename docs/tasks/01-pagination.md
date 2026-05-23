# タスク01: ページネーション・フィルタリング

## 概要

Todo一覧取得APIにページネーション、フィルタリング、ソート機能を追加する。

---

## 実装ステップ

### 1. ドメイン型の追加

- [ ] `internal/domain/todo.go` に `TodoFilter` 構造体を追加
- [ ] `internal/domain/todo.go` に `TodoListResult` 構造体を追加
- [ ] `internal/domain/repository.go` の `TodoRepository.FindAll` シグネチャを変更（`TodoFilter` を引数に追加、戻り値を `*TodoListResult` に変更）

### 2. リポジトリ層の実装

- [ ] `internal/repository/todo.go` の `FindAll` を修正
  - 動的WHERE句の構築（completed、due_before、due_after）
  - ソートフィールドのホワイトリスト検証
  - LIMIT/OFFSETによるページネーション
  - COUNTクエリで総件数を取得
  - フィルタ条件の値をログ出力（Debug レベル）

### 3. ユースケース層の実装

- [ ] `internal/usecase/todo.go` の `List` メソッドを修正
  - フィルタのデフォルト値設定（page=1, limit=20, sort=id, order=asc）
  - limitの上限チェック（最大100）
  - sort/orderのバリデーション（不正値は `ValidationError`）

### 4. ハンドラー層の実装

- [ ] `internal/handler/todo.go` の `List` メソッドを修正
  - クエリパラメータのパース（page, limit, completed, due_before, due_after, sort, order）
  - レスポンス形式を `{ "data": [...], "meta": {...} }` に変更

### 5. テストの更新

- [ ] `internal/repository/todo_test.go` — ページネーションのテスト追加
  - 正常系: ページ2の取得、limit指定、フィルタ条件の組み合わせ
  - 境界値: limit=0, limit=101, page=0
- [ ] `internal/usecase/todo_test.go` — フィルタデフォルト値・バリデーションのテスト追加
- [ ] `internal/handler/todo_test.go` — クエリパラメータのパース・レスポンス形式のテスト追加

---

## 成功基準

- `GET /api/todos` でページネーション、フィルタリング、ソートが動作すること
- レスポンスに `data` と `meta` が含まれること
- 不正なsort値で `ValidationError`（400）が返ること
- 既存のテストがすべてパスすること
