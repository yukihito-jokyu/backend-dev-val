# タスク02: ソフトデリート

## 概要

Todoの削除を論理削除に変更し、削除済みデータの復元・一覧取得・物理削除を追加する。

---

## 前提条件

- タスク01（ページネーション・フィルタリング）が完了していること

---

## 実装ステップ

### 1. マイグレーション

- [ ] `migrations/00003_add_soft_delete_to_todos.sql` を作成
  - Up: `ALTER TABLE todos ADD COLUMN deleted_at TIMESTAMP;`
  - Down: `ALTER TABLE todos DROP COLUMN deleted_at;`
- [ ] `task db-migrate` を実行してマイグレーションを適用

### 2. ドメイン層の修正

- [ ] `internal/domain/todo.go` の `Todo` 構造体に `DeletedAt *time.Time` を追加
- [ ] `internal/domain/repository.go` にメソッドを追加
  - `FindDeletedAll(ctx, TodoFilter) (*TodoListResult, error)`
  - `FindDeletedByID(ctx, int64) (*Todo, error)`
  - `SoftDelete(ctx, int64) error`
  - `Restore(ctx, int64) error`
  - `PermanentDelete(ctx, int64) error`

### 3. リポジトリ層の実装

- [ ] `internal/repository/todo.go` を修正
  - `FindAll`: `WHERE deleted_at IS NULL` を追加
  - `FindByID`: `WHERE deleted_at IS NULL` を追加
  - `SoftDelete`: `UPDATE todos SET deleted_at = NOW() WHERE id = $1`
  - `Restore`: `UPDATE todos SET deleted_at = NULL WHERE id = $1`
  - `FindDeletedAll`: `WHERE deleted_at IS NOT NULL` でページネーション
  - `FindDeletedByID`: `WHERE deleted_at IS NOT NULL AND id = $1`
  - `PermanentDelete`: `DELETE FROM todos WHERE id = $1 AND deleted_at IS NOT NULL`
  - 既存 `Delete` メソッドの削除（`SoftDelete` に置き換え）

### 4. ユースケース層の実装

- [ ] `internal/usecase/todo.go` にメソッドを追加
  - `ListTrash(ctx, TodoFilter) (*TodoListResult, error)`
  - `Restore(ctx, int64) (*Todo, error)` — 未削除の場合は `NotFoundError`
  - `PermanentDelete(ctx, int64) error` — 未削除の場合は `NotFoundError`
  - 既存 `Delete` メソッドを `SoftDelete` 呼び出しに変更

### 5. ハンドラー層の実装

- [ ] `internal/handler/todo.go` にハンドラを追加
  - `ListTrash`: `GET /api/todos/trash` — クエリパラメータからフィルタ生成
  - `Restore`: `POST /api/todos/:id/restore` — 200で復元後のTodoを返す
  - `PermanentDelete`: `DELETE /api/todos/:id/permanent` — 204 No Content
- [ ] `internal/handler/router.go` にルートを追加
  - `/trash` は `/:id` より前に登録

### 6. テストの更新

- [ ] `internal/repository/todo_test.go`
  - ソフトデリート後の通常一覧に含まれないこと
  - 削除済み一覧に含まれること
  - 復元後の通常一覧に戻ること
  - 物理削除でデータが消えること
  - 未削除データの復元・物理削除で `NotFoundError`
- [ ] `internal/usecase/todo_test.go`
  - Delete がソフトデリートを呼ぶこと
  - Restore/PermanentDelete の正常系・異常系
- [ ] `internal/handler/todo_test.go`
  - trash/restore/permanent のエンドポイントテスト

---

## 成功基準

- `DELETE /api/todos/:id` で論理削除されること
- `GET /api/todos` に削除済みデータが含まれないこと
- `GET /api/todos/trash` で削除済み一覧が取得できること
- `POST /api/todos/:id/restore` で復元できること
- `DELETE /api/todos/:id/permanent` で物理削除できること
- 未削除データに対する restore/permanent で 404 が返ること
