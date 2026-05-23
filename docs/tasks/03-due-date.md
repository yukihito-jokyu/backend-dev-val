# タスク03: 期限管理（Due Date）

## 概要

Todoに期限（due_date）を設定可能にし、期限切れ状態を判定する機能を追加する。

---

## 前提条件

- タスク01（ページネーション・フィルタリング）が完了していること

---

## 実装ステップ

### 1. マイグレーション

- [ ] `migrations/00002_add_due_date_to_todos.sql` を作成
  - Up: `ALTER TABLE todos ADD COLUMN due_date TIMESTAMP;`
  - Down: `ALTER TABLE todos DROP COLUMN due_date;`
- [ ] `task db-migrate` を実行してマイグレーションを適用

### 2. ドメイン層の修正

- [ ] `internal/domain/todo.go` の `Todo` 構造体を修正
  - `DueDate *time.Time` を追加（NULL許容）
  - `IsOverdue bool` を追加（計算フィールド）
- [ ] `internal/domain/todo.go` に `CalcOverdue` 関数を追加
  - `DueDate != nil && !Completed && DueDate.Before(time.Now())` の場合 `true`

### 3. リポジトリ層の実装

- [ ] `internal/repository/todo.go` を修正
  - `Create`: `due_date` をINSERT対象に追加
  - `Update`: `due_date` をUPDATE対象に追加
  - `FindAll`/`FindByID`: SELECTに `due_date` を追加
  - フィルタに `due_before`、`due_after` の条件を追加（タスク01で実装済みの場合は確認のみ）

### 4. ユースケース層の実装

- [ ] `internal/usecase/todo.go` を修正
  - `CreateTodoInput` に `DueDate *time.Time` を追加
  - `UpdateTodoInput` に `DueDate **time.Time` を追加（nil=変更なし、pointer-to-nil=クリア）
  - `Create`/`Update` で `due_date` のRFC3339バリデーション（handler層で実施）
  - 取得時（List/GetByID/ListTrash/Restore）に `CalcOverdue` を呼び出す

### 5. ハンドラー層の実装

- [ ] `internal/handler/todo.go` を修正
  - `createTodoRequest` に `DueDate *string` を追加
  - `updateTodoRequest` に `DueDate *string` を追加
  - RFC3339パース失敗時は `ValidationError` を返す
  - レスポンスの `due_date` はRFC3339形式、null許容
  - レスポンスに `is_overdue` を含める

### 6. テストの更新

- [ ] `internal/repository/todo_test.go`
  - due_date付きの作成・更新・取得
  - due_date=null の作成・更新
- [ ] `internal/usecase/todo_test.go`
  - is_overdue の判定ロジック（期限前=false、期限後=true、完了済み=false、due_date未設定=false）
  - CalcOverdue の境界値テスト
- [ ] `internal/handler/todo_test.go`
  - 不正なdue_dateフォーマットで 400
  - レスポンスに due_date、is_overdue が含まれること

---

## 成功基準

- `POST /api/todos` で `due_date` を指定して作成できること
- `PUT /api/todos/:id` で `due_date` を更新・クリアできること
- 期限切れのTodoの `is_overdue` が `true` であること
- 完了済みTodoは期限切れでも `is_overdue` が `false` であること
- 不正な日時フォーマットで `ValidationError`（400）が返ること
