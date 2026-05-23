# AGENTS.md

## 開発環境コマンド

Task (Taskfile) をタスクランナーとして使用する。

### コマンド一覧

| コマンド | 用途 | 実行例 |
|---|---|---|
| `task build` | アプリケーションのビルド | `task build` |
| `task run` | アプリケーションの起動 | `task run` |
| `task test` | テスト実行（カバレッジ付き） | `task test` |
| `task lint` | golangci-lintによる静的解析 | `task lint` |
| `task fmt` | goimportsによるコードフォーマット | `task fmt` |
| `task db-up` | PostgreSQL起動（docker compose） | `task db-up` |
| `task db-down` | PostgreSQL停止 | `task db-down` |
| `task db-migrate` | gooseマイグレーション実行 | `task db-migrate` |
| `task db-rollback` | gooseマイグレーション巻き戻し | `task db-rollback` |
| `task db-status` | マイグレーション状態確認 | `task db-status` |
| `task wire` | Wire DIコード生成 | `task wire` |
| `task setup` | 初期セットアップ（依存インストール + DB起動 + マイグレーション） | `task setup` |

## プロジェクト共通ルール

### Go

- Go 1.24。モジュールパスは `backend-dev-val`。
- `database/sql` を直接使用し、ORMは使用しない。
- PostgreSQLドライバは `github.com/lib/pq` を使用する。

### アーキテクチャ

クリーンアーキテクチャ（レイヤー別）を採用。依存の向きは handler → usecase → domain ← repository。

```
internal/
├── domain/        # エンティティ + リポジトリインターフェース + 独自エラー型
├── usecase/       # ビジネスロジック（domainのインターフェースに依存）
├── handler/       # Gin HTTPハンドラー（プレゼンテーション層）
├── repository/    # リポジトリ実装（PostgreSQL）
└── di/            # Wire DI（wire.go + wire_gen.go）
```

#### 依存のルール

- **domain** は他のパッケージに依存しない（純粋なGo型のみ）。
- **usecase** は `domain` のインターフェースにのみ依存する（実装に依存しない）。
- **handler** は `usecase` に依存する。`repository` や `database/sql` には直接触れない。
- **repository** は `domain` のインターフェースを実装する。
- **di** が全てを組み立てる。

### エラー処理

- ドメインエラーは `internal/domain/errors.go` に定義する。
- 現在のエラー型: `NotFoundError`, `ValidationError`。
- 新しいエラー型が必要な場合は `domain/errors.go` に追加する。
- handler層（`respondError`）でエラー型に応じてHTTPステータスをマッピングする:
  - `ValidationError` → 400 Bad Request
  - `NotFoundError` → 404 Not Found
  - その他 → 500 Internal Server Error

### ロギング

`log/slog`（標準ライブラリ）を使用する。外部ロガーは使用しない。

#### 初期化

`cmd/server/main.go` で JSONハンドラ + `slog.LevelInfo` で初期化し、`slog.SetDefault()` でグローバルロガーとして設定する。

#### ログレベルの使い分け

| レベル | 用途 | 例 |
|---|---|---|
| `Error` | 処理の失敗。調査が必要 | DB接続エラー、予期しないパニック |
| `Info` | 正常な業務処理の記録 | サーバー起動、DB接続成功 |
| `Debug` | 開発・デバッグ用の詳細情報 | SQLクエリ、関数の引数・戻り値 |

#### 構造化ログ

コンテキスト情報は `slog.String` 等の型付き引数で渡す:

```go
slog.Info("starting server", "addr", addr)
slog.Error("failed to ping database", "error", err)
```

#### 機密情報の取り扱い

- ログ出力にDB接続文字列のパスワードを含めてはならない。
- ユーザーの個人情報をログに出力してはならない。

### 設定

環境変数を `os.Getenv` で直接読み取る。外部設定ライブラリは使用しない。

| 環境変数 | デフォルト値 | 用途 |
|---|---|---|
| `DB_HOST` | `localhost` | PostgreSQLホスト |
| `DB_PORT` | `5432` | PostgreSQLポート |
| `DB_USER` | `postgres` | DBユーザー |
| `DB_PASSWORD` | `postgres` | DBパスワード |
| `DB_NAME` | `backend_dev_val` | DB名 |
| `DB_SSLMODE` | `disable` | SSLモード |
| `PORT` | `8080` | サーバーポート |

### API

- REST API。エンドポイントは `/api` プレフィックス付き。
- JSONリクエスト/レスポンス。
- エラーレスポンスは `{"error": "message"}` 形式。

### マイグレーション

- `migrations/` ディレクトリに配置。
- ファイル名は連番プレフィックス: `00001_create_todos.sql`, `00002_...`。
- gooseフォーマット（`-- +goose Up` / `-- +goose Down`）。

### DI（Wire）

- `internal/di/wire.go` にWire定義を記述（`//go:build wireinject` タグ付き）。
- `internal/di/wire_gen.go` はWireが生成するコード。手動編集しない。
- 新しい依存を追加した場合は `task wire` で再生成する。

### テスト

#### フレームワーク・規約

- テストライブラリは **stretchr/testify** v1.11.0 を使用する。
- テストファイルはソースファイルと同じディレクトリに配置する（例: `internal/usecase/todo_test.go`）。
- テーブル駆動テスト（Table-Driven Tests）形式で実装する。
- C1レベル（分岐網羅）のテストカバレッジを目標とする。

#### テスト設計原則

- 公開関数・メソッドの入出力と副作用をテストする。内部実装の詳細はテストしない。
- 1つのテストケースにつき1つの観点（正常系・異常系・境界値）。境界値テストは必ず全て実装する。
- アサーションは `testify/assert` / `testify/require` を使用し、具体的な値で検証する。
- リポジトリ層のテストは実際のDBに接続して行う（モックを使わない）。テスト用DBは `backend_dev_val_test` を使用する。

#### テーブル駆動テストの書き方

```go
tests := []struct {
    name    string
    input   CreateTodoInput
    want    *domain.Todo
    wantErr bool
}{
    {
        name:    "正常系: タイトルのみで作成",
        input:   CreateTodoInput{Title: "test todo"},
        want:    &domain.Todo{Title: "test todo"},
        wantErr: false,
    },
    // ...
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        got, err := uc.Create(ctx, tt.input)
        if tt.wantErr {
            assert.Error(t, err)
            return
        }
        assert.NoError(t, err)
        assert.Equal(t, tt.want.Title, got.Title)
    })
}
```

#### モック戦略

- usecase層のテストでは、`domain.TodoRepository` インターフェースのモックを使用する。
- モックは testify/mock または手動でスタブを実装する。
- handler層のテストでは、`httptest.NewRecorder()` + Ginのコンテキストを使用する。

### コードフォーマット

- `goimports` でフォーマット統一。
- `task fmt` で一括適用。
- golangci-lint v2 にもgoimportsフォーマッタが統合されている。

### Linter

- golangci-lint v2 を使用。
- 有効なlinter: `errcheck`, `govet`, `staticcheck`, `unused`, `gosimple`, `ineffassign`, `typecheck`, `gocritic`, `revive`。
- `task lint` で実行。
