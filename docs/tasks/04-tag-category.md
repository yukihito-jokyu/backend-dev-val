# タスク04: タグ・カテゴリ機能

## 概要

タグのCRUDと、Todoへのタグ付け機能（多対多リレーション）を追加する。

---

## 前提条件

- タスク01（ページネーション・フィルタリング）が完了していること
- タスク02（ソフトデリート）が完了していること

---

## 実装ステップ

### 1. マイグレーション

- [ ] `migrations/00004_create_tags.sql` を作成
  - Up: `tags` テーブル作成、`todo_tags` 中間テーブル作成
  - Down: `todo_tags` → `tags` の順でDROP
- [ ] `task db-migrate` を実行してマイグレーションを適用

### 2. ドメイン層の追加

- [ ] `internal/domain/tag.go` を作成
  - `Tag` 構造体（ID, Name, CreatedAt, UpdatedAt）
- [ ] `internal/domain/repository.go` に `TagRepository` インターフェースを追加
  - `FindAll(ctx) ([]*Tag, error)`
  - `FindByID(ctx, int64) (*Tag, error)`
  - `FindByIDs(ctx, []int64) ([]*Tag, error)`
  - `Create(ctx, *Tag) (*Tag, error)`
  - `Update(ctx, *Tag) (*Tag, error)`
  - `Delete(ctx, int64) error`
  - `ExistsByName(ctx, string, *int64) (bool, error)`
- [ ] `internal/domain/todo.go` の `Todo` 構造体に `Tags []Tag` を追加
- [ ] `internal/domain/repository.go` の `TodoRepository` にメソッドを追加
  - `ReplaceTags(ctx, todoID int64, tagIDs []int64) error`

### 3. リポジトリ層の実装

- [ ] `internal/repository/tag.go` を作成
  - `TagRepository` の全メソッドを実装
  - `ExistsByName`: 作成時は `excludeID=nil`、更新時は `excludeID=&id` で自身を除外
- [ ] `internal/repository/todo.go` を修正
  - `ReplaceTags`: トランザクション内で DELETE → INSERT を実行
  - `Create`/`Update`: トランザクション内でTodo作成/更新 + `ReplaceTags`
  - `FindAll`/`FindByID`: タグ情報も一括取得（N+1回避）
  - タグ取得クエリ: `SELECT t.*, tt.todo_id FROM tags t JOIN todo_tags tt ON t.id = tt.tag_id WHERE tt.todo_id = ANY($1)`

### 4. ユースケース層の実装

- [ ] `internal/usecase/tag.go` を作成
  - `TagUsecase` 構造体（`TagRepository` に依存）
  - `List`, `Create`, `Update`, `Delete` メソッド
  - `Create`/`Update` で名前重複チェック（`ExistsByName`）
  - 重複時は `ValidationError`
- [ ] `internal/usecase/todo.go` を修正
  - `TodoUsecase` に `tagRepo TagRepository` を追加
  - `CreateTodoInput` に `TagIDs []int64` を追加
  - `UpdateTodoInput` に `TagIDs []int64` を追加（nil=変更なし）
  - `Create`/`Update` で `tag_ids` の存在チェック（`FindByIDs`）
    - 指定IDの数と取得結果の数が一致しない場合は `NotFoundError`
  - `Create`/`Update` で `ReplaceTags` を呼び出す

### 5. ハンドラー層の実装

- [ ] `internal/handler/tag.go` を作成
  - `TagHandler` 構造体（`TagUsecase` に依存）
  - `List`, `Create`, `Update`, `Delete` メソッド
  - `createTagRequest`: `Name string`（binding:required）
  - `updateTagRequest`: `Name string`（binding:required）
- [ ] `internal/handler/todo.go` を修正
  - `createTodoRequest` に `TagIDs []int64` を追加
  - `updateTodoRequest` に `TagIDs []int64` を追加
  - レスポンスに `tags` 配列を含める
- [ ] `internal/handler/router.go` を修正
  - `/api/tags` ルートグループを追加
  - `SetupRouter` のシグネチャに `TagHandler` を追加

### 6. DI（Wire）の更新

- [ ] `internal/di/wire.go` を修正
  - `NewTagRepository`、`NewTagUsecase`、`NewTagHandler` を追加
- [ ] `task wire` を実行して `wire_gen.go` を再生成

### 7. テストの追加・更新

- [ ] `internal/repository/tag_test.go` を作成
  - CRUD正常系・異常系
  - 名前重複チェック（`ExistsByName`）
- [ ] `internal/repository/todo_test.go` を更新
  - タグ付きTodoの作成・更新・取得
  - タグ付けの置き換え
- [ ] `internal/usecase/tag_test.go` を作成
  - モックリポジトリでCRUDテスト
  - 名前重複で `ValidationError`
- [ ] `internal/usecase/todo_test.go` を更新
  - `tag_ids` 指定時の存在チェック（存在しないID → `NotFoundError`）
- [ ] `internal/handler/tag_test.go` を作成
  - タグCRUDのエンドポイントテスト
- [ ] `internal/handler/todo_test.go` を更新
  - `tag_ids` を含むリクエストのテスト
  - レスポンスに `tags` が含まれること

---

## 成功基準

- `POST /api/tags` でタグを作成できること
- `GET /api/tags` でタグ一覧を取得できること
- `PUT /api/tags/:id` でタグ名を更新できること
- `DELETE /api/tags/:id` でタグを削除できること（紐づく関連も削除）
- 同名タグの作成で `ValidationError`（400）が返ること
- `POST /api/todos` で `tag_ids` を指定してTodoを作成できること
- `PUT /api/todos/:id` で `tag_ids` を更新できること
- 存在しない `tag_ids` 指定で `NotFoundError`（404）が返ること
- Todoのレスポンスに `tags` 配列が含まれること
- `GET /api/todos?tag_id=1` でタグによるフィルタが動作すること
