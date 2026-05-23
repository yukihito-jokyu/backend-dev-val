# 詳細設計書

## 1. データベース設計

### 1.1 ER図

```
todos            todo_tags        tags
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│ id          │  │ todo_id     │  │ id          │
│ title       │  │ tag_id      │  │ name        │
│ description │  └─────────────┘  │ created_at  │
│ completed   │                   │ updated_at  │
│ due_date    │                   └─────────────┘
│ deleted_at  │
│ created_at  │
│ updated_at  │
└─────────────┘
```

### 1.2 マイグレーション

#### 00002_add_due_date_to_todos.sql

```sql
-- +goose Up
ALTER TABLE todos ADD COLUMN due_date TIMESTAMP;

-- +goose Down
ALTER TABLE todos DROP COLUMN due_date;
```

#### 00003_add_soft_delete_to_todos.sql

```sql
-- +goose Up
ALTER TABLE todos ADD COLUMN deleted_at TIMESTAMP;

-- +goose Down
ALTER TABLE todos DROP COLUMN deleted_at;
```

#### 00004_create_tags.sql

```sql
-- +goose Up
CREATE TABLE tags (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE todo_tags (
    todo_id BIGINT NOT NULL REFERENCES todos(id) ON DELETE CASCADE,
    tag_id BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (todo_id, tag_id)
);

-- +goose Down
DROP TABLE todo_tags;
DROP TABLE tags;
```

### 1.3 テーブル定義

#### todos テーブル

| カラム      | 型           | 制約                   | 備考                    |
| ----------- | ------------ | ---------------------- | ----------------------- |
| id          | BIGSERIAL    | PRIMARY KEY            |                         |
| title       | VARCHAR(255) | NOT NULL               |                         |
| description | TEXT         | NOT NULL DEFAULT ''    |                         |
| completed   | BOOLEAN      | NOT NULL DEFAULT FALSE |                         |
| due_date    | TIMESTAMP    | NULLABLE               | 期限日時                |
| deleted_at  | TIMESTAMP    | NULLABLE               | 削除日時（NULL=未削除） |
| created_at  | TIMESTAMP    | NOT NULL DEFAULT NOW() |                         |
| updated_at  | TIMESTAMP    | NOT NULL DEFAULT NOW() |                         |

#### tags テーブル

| カラム     | 型           | 制約                   | 備考 |
| ---------- | ------------ | ---------------------- | ---- |
| id         | BIGSERIAL    | PRIMARY KEY            |      |
| name       | VARCHAR(255) | NOT NULL UNIQUE        |      |
| created_at | TIMESTAMP    | NOT NULL DEFAULT NOW() |      |
| updated_at | TIMESTAMP    | NOT NULL DEFAULT NOW() |      |

#### todo_tags テーブル

| カラム  | 型     | 制約                          | 備考   |
| ------- | ------ | ----------------------------- | ------ |
| todo_id | BIGINT | NOT NULL REFERENCES todos(id) | 複合PK |
| tag_id  | BIGINT | NOT NULL REFERENCES tags(id)  | 複合PK |

---

## 2. API設計

### 2.1 エンドポイント一覧

#### Todo API

| メソッド | パス                       | 概要                                               |
| -------- | -------------------------- | -------------------------------------------------- |
| GET      | `/api/todos`               | 一覧取得（ページネーション・フィルタ・ソート対応） |
| GET      | `/api/todos/:id`           | 個別取得                                           |
| POST     | `/api/todos`               | 新規作成                                           |
| PUT      | `/api/todos/:id`           | 更新                                               |
| DELETE   | `/api/todos/:id`           | ソフトデリート                                     |
| GET      | `/api/todos/trash`         | 削除済み一覧取得                                   |
| POST     | `/api/todos/:id/restore`   | 復元                                               |
| DELETE   | `/api/todos/:id/permanent` | 物理削除                                           |

#### Tag API

| メソッド | パス            | 概要         |
| -------- | --------------- | ------------ |
| GET      | `/api/tags`     | タグ一覧取得 |
| POST     | `/api/tags`     | タグ作成     |
| PUT      | `/api/tags/:id` | タグ更新     |
| DELETE   | `/api/tags/:id` | タグ削除     |

### 2.2 リクエスト・レスポンス定義

#### GET /api/todos

**クエリパラメータ**

| パラメータ | 型              | デフォルト | 説明                                                |
| ---------- | --------------- | ---------- | --------------------------------------------------- |
| page       | int             | 1          | ページ番号                                          |
| limit      | int             | 20         | 1ページあたりの件数（最大100）                      |
| completed  | bool            | -          | 完了状態フィルタ                                    |
| due_before | string(RFC3339) | -          | 期限 ≤ 指定日                                       |
| due_after  | string(RFC3339) | -          | 期限 ≥ 指定日                                       |
| tag_id     | int64           | -          | タグIDフィルタ                                      |
| sort       | string          | id         | ソートフィールド（id, title, due_date, created_at） |
| order      | string          | asc        | ソート順（asc, desc）                               |

**レスポンス 200**

```json
{
  "data": [
    {
      "id": 1,
      "title": "タスクA",
      "description": "説明文",
      "completed": false,
      "due_date": "2025-04-01T00:00:00Z",
      "is_overdue": true,
      "tags": [{ "id": 1, "name": "urgent" }],
      "created_at": "2025-03-01T10:00:00Z",
      "updated_at": "2025-03-01T10:00:00Z"
    }
  ],
  "meta": {
    "total": 50,
    "page": 1,
    "limit": 20,
    "total_pages": 3
  }
}
```

#### POST /api/todos

**リクエストボディ**

```json
{
  "title": "タスクA",
  "description": "説明文",
  "due_date": "2025-04-01T00:00:00Z",
  "tag_ids": [1, 2]
}
```

**レスポンス 201**

```json
{
  "id": 1,
  "title": "タスクA",
  "description": "説明文",
  "completed": false,
  "due_date": "2025-04-01T00:00:00Z",
  "is_overdue": false,
  "tags": [
    { "id": 1, "name": "urgent" },
    { "id": 2, "name": "work" }
  ],
  "created_at": "2025-03-01T10:00:00Z",
  "updated_at": "2025-03-01T10:00:00Z"
}
```

#### PUT /api/todos/:id

**リクエストボディ**

```json
{
  "title": "更新タイトル",
  "description": "更新説明",
  "completed": true,
  "due_date": "2025-05-01T00:00:00Z",
  "tag_ids": [1, 3]
}
```

- `tag_ids` が指定された場合、既存のタグ付けを全て置き換える。
- `tag_ids` が未指定の場合、タグ付けは変更しない。

#### GET /api/todos/trash

**レスポンス 200**

一覧取得と同じ形式。削除済み（`deleted_at IS NOT NULL`）のTodoのみ返す。

#### POST /api/todos/:id/restore

**レスポンス 200**

復元後のTodoオブジェクト。未削除のTodoに対する実行は `NotFoundError`。

#### DELETE /api/todos/:id/permanent

**レスポンス 204 No Content**

削除済みのTodoのみ物理削除可能。未削除のTodoに対する実行は `NotFoundError`。

#### POST /api/tags

**リクエストボディ**

```json
{
  "name": "urgent"
}
```

**レスポンス 201**

```json
{
  "id": 1,
  "name": "urgent",
  "created_at": "2025-03-01T10:00:00Z",
  "updated_at": "2025-03-01T10:00:00Z"
}
```

#### PUT /api/tags/:id

**リクエストボディ**

```json
{
  "name": "updated-name"
}
```

#### DELETE /api/tags/:id

**レスポンス 204 No Content**

紐づく `todo_tags` はCASCADEで自動削除。

### 2.3 エラーレスポンス

| ステータス | エラー型              | 発生条件                                                       |
| ---------- | --------------------- | -------------------------------------------------------------- |
| 400        | `ValidationError`     | バリデーション違反（タイトル空、タグ名重複、不正なsort値など） |
| 404        | `NotFoundError`       | リソース未検出（Todo ID不在、タグID不在）                      |
| 500        | Internal Server Error | 上記以外のサーバーエラー                                       |

---

## 3. ドメインモデル設計

### 3.1 エンティティ

#### Todo

```go
type Todo struct {
    ID          int64     `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Completed   bool      `json:"completed"`
    DueDate     *time.Time `json:"due_date"`
    DeletedAt   *time.Time `json:"-"`
    IsOverdue   bool      `json:"is_overdue"`
    Tags        []Tag     `json:"tags"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

- `DueDate`: ポインタ型（NULL許容）。未設定時は `nil`。
- `DeletedAt`: ポインタ型。未削除時は `nil`。JSON出力しない。
- `IsOverdue`: 計算フィールド。`DueDate != nil && !Completed && DueDate < now` の場合 `true`。
- `Tags`: 関連するタグのスライス。

#### Tag

```go
type Tag struct {
    ID        int64     `json:"id"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### 3.2 リポジトリインターフェース

#### TodoRepository

```go
type TodoRepository interface {
    FindAll(ctx context.Context, filter TodoFilter) (*TodoListResult, error)
    FindByID(ctx context.Context, id int64) (*Todo, error)
    FindDeletedAll(ctx context.Context, filter TodoFilter) (*TodoListResult, error)
    FindDeletedByID(ctx context.Context, id int64) (*Todo, error)
    Create(ctx context.Context, todo *Todo) (*Todo, error)
    Update(ctx context.Context, todo *Todo) (*Todo, error)
    SoftDelete(ctx context.Context, id int64) error
    Restore(ctx context.Context, id int64) error
    PermanentDelete(ctx context.Context, id int64) error
    ReplaceTags(ctx context.Context, todoID int64, tagIDs []int64) error
}
```

#### TagRepository

```go
type TagRepository interface {
    FindAll(ctx context.Context) ([]*Tag, error)
    FindByID(ctx context.Context, id int64) (*Tag, error)
    FindByIDs(ctx context.Context, ids []int64) ([]*Tag, error)
    Create(ctx context.Context, tag *Tag) (*Tag, error)
    Update(ctx context.Context, tag *Tag) (*Tag, error)
    Delete(ctx context.Context, id int64) error
    ExistsByName(ctx context.Context, name string, excludeID *int64) (bool, error)
}
```

### 3.3 フィルタ・結果型

```go
type TodoFilter struct {
    Page       int
    Limit      int
    Completed  *bool
    DueBefore  *time.Time
    DueAfter   *time.Time
    TagID      *int64
    Sort       string
    Order      string
}

type TodoListResult struct {
    Todos      []*Todo
    Total      int64
    Page       int
    Limit      int
    TotalPages int
}
```

### 3.4 ドメインエラー

既存の `NotFoundError`、`ValidationError` に加え、変更なし。

---

## 4. ユースケース設計

### 4.1 TodoUsecase

```go
type TodoUsecase struct {
    repo  domain.TodoRepository
    tagRepo domain.TagRepository
}
```

#### メソッド一覧

| メソッド          | 入力                       | 出力              | ロジック                                   |
| ----------------- | -------------------------- | ----------------- | ------------------------------------------ |
| `List`            | `TodoFilter`               | `*TodoListResult` | `FindAll` を呼び出し、`is_overdue` を計算  |
| `GetByID`         | `int64`                    | `*Todo`           | `FindByID` → `is_overdue` 計算             |
| `Create`          | `CreateTodoInput`          | `*Todo`           | バリデーション → `tag_ids` 存在確認 → 作成 |
| `Update`          | `int64`, `UpdateTodoInput` | `*Todo`           | バリデーション → `tag_ids` 存在確認 → 更新 |
| `Delete`          | `int64`                    | error             | `SoftDelete` を呼び出し                    |
| `ListTrash`       | `TodoFilter`               | `*TodoListResult` | `FindDeletedAll` → `is_overdue` 計算       |
| `Restore`         | `int64`                    | `*Todo`           | `Restore` → `is_overdue` 計算              |
| `PermanentDelete` | `int64`                    | error             | `PermanentDelete` を呼び出し               |

#### 入力DTO

```go
type CreateTodoInput struct {
    Title       string
    Description string
    DueDate     *time.Time
    TagIDs      []int64
}

type UpdateTodoInput struct {
    Title       *string
    Description *string
    Completed   *bool
    DueDate     **time.Time  // nil=変更なし、pointer-to-nil=クリア
    TagIDs      []int64      // nil=変更なし、空スライス=全解除
}
```

### 4.2 TagUsecase

```go
type TagUsecase struct {
    repo domain.TagRepository
}
```

#### メソッド一覧

| メソッド | 入力                      | 出力     | ロジック                |
| -------- | ------------------------- | -------- | ----------------------- |
| `List`   | -                         | `[]*Tag` | `FindAll` を呼び出し    |
| `Create` | `CreateTagInput`          | `*Tag`   | 名前重複チェック → 作成 |
| `Update` | `int64`, `UpdateTagInput` | `*Tag`   | 名前重複チェック → 更新 |
| `Delete` | `int64`                   | error    | `Delete` を呼び出し     |

---

## 5. リポジトリ実装設計

### 5.1 動的クエリビルディング

ページネーション・フィルタリングに対応するため、SQLを動的に構築する。

```sql
-- ベースクエリ
SELECT id, title, description, completed, due_date, deleted_at, created_at, updated_at
FROM todos
WHERE deleted_at IS NULL
-- 動的WHERE句（フィルタ条件に応じて追加）
  AND (completed = $1 OR $1 IS NULL)
  AND (due_date <= $2 OR $2 IS NULL)
  AND (due_date >= $3 OR $3 IS NULL)
-- ソート（ホワイトリストでSQLインジェクション対策）
ORDER BY id ASC
-- ページネーション
LIMIT $4 OFFSET $5
```

### 5.2 COUNTクエリ

ページネーションのメタ情報取得のため、同条件でCOUNTを実行する。

```sql
SELECT COUNT(*) FROM todos
WHERE deleted_at IS NULL
-- 同じフィルタ条件
```

### 5.3 ソートのホワイトリスト

SQLインジェクションを防ぐため、ソートフィールドはホワイトリストで検証する。

```go
var allowedSortFields = map[string]string{
    "id":         "id",
    "title":      "title",
    "due_date":   "due_date",
    "created_at": "created_at",
}
```

### 5.4 タグの取得

Todo取得時にタグ情報も取得する。N+1問題を回避するため、IDs収集後に一括取得する。

```sql
-- タグ一括取得
SELECT t.id, t.name, t.created_at, t.updated_at, tt.todo_id
FROM tags t
JOIN todo_tags tt ON t.id = tt.tag_id
WHERE tt.todo_id = ANY($1)
```

### 5.5 トランザクション

Todo作成・更新時のタグ付け操作はトランザクション内で実行する。

```
BEGIN
  → INSERT/UPDATE todos
  → DELETE FROM todo_tags WHERE todo_id = $1
  → INSERT INTO todo_tags (todo_id, tag_id) VALUES ($1, $2), ...
COMMIT
```

---

## 6. ハンドラー設計

### 6.1 TodoHandler 追加メソッド

| メソッド          | ルート                            | 処理                                           |
| ----------------- | --------------------------------- | ---------------------------------------------- |
| `ListTrash`       | `GET /api/todos/trash`            | クエリパラメータからフィルタ生成 → `ListTrash` |
| `Restore`         | `POST /api/todos/:id/restore`     | ID取得 → `Restore`                             |
| `PermanentDelete` | `DELETE /api/todos/:id/permanent` | ID取得 → `PermanentDelete`                     |

### 6.2 TagHandler

新規作成。`TodoHandler` と同じパターンで実装。

### 6.3 ルーティング

```go
func SetupRouter(th *TodoHandler, tagH *TagHandler) *gin.Engine {
    r := gin.Default()
    api := r.Group("/api")
    {
        todos := api.Group("/todos")
        {
            todos.GET("", th.List)
            todos.GET("/trash", th.ListTrash)
            todos.GET("/:id", th.GetByID)
            todos.POST("", th.Create)
            todos.PUT("/:id", th.Update)
            todos.DELETE("/:id", th.Delete)
            todos.POST("/:id/restore", th.Restore)
            todos.DELETE("/:id/permanent", th.PermanentDelete)
        }
        tags := api.Group("/tags")
        {
            tags.GET("", tagH.List)
            tags.POST("", tagH.Create)
            tags.PUT("/:id", tagH.Update)
            tags.DELETE("/:id", tagH.Delete)
        }
    }
    return r
}
```

**注意**: `/trash`、`/restore`、`/permanent` は `/:id` より前に登録し、ルーティングの競合を防ぐ。

### 6.4 is_overdue の計算

usecase層で計算し、エンティティに設定してhandlerに返す。

```go
func calcOverdue(todo *domain.Todo) {
    todo.IsOverdue = todo.DueDate != nil && !todo.Completed && todo.DueDate.Before(time.Now())
}
```

---

## 7. DI（Wire）設計

```go
func InitializeApp(db *sql.DB) *gin.Engine {
    wire.Build(
        repository.NewTodoRepository,
        repository.NewTagRepository,
        usecase.NewTodoUsecase,
        usecase.NewTagUsecase,
        handler.NewTodoHandler,
        handler.NewTagHandler,
        handler.SetupRouter,
    )
    return nil
}
```

`TagRepository`、`TagUsecase`、`TagHandler` を追加。
