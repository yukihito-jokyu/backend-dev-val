# backend-dev-val

AI駆動開発ツールの検証用バックエンドプロジェクト。TodoアプリAPIを題材に、各種AI開発ツールでの使用感・開発効率を検証する。

## 技術スタック

| カテゴリ | 技術 |
|---|---|
| 言語 | Go 1.24 |
| フレームワーク | Gin |
| アーキテクチャ | クリーンアーキテクチャ（レイヤー別） |
| DB | PostgreSQL |
| DBアクセス | database/sql（ORMなし） |
| マイグレーション | goose |
| DI | google/wire |
| ログ | log/slog |
| テスト | stretchr/testify / テーブル駆動テスト |
| タスクランナー | Task (Taskfile) |
| リンター | golangci-lint v2 |
| フォーマッタ | goimports |

## セットアップ

```bash
# 依存関係のインストール + DB起動 + マイグレーション
task setup
```

または手動で実行:

```bash
go mod tidy
task db-up
sleep 2
task db-migrate
```

## 起動

```bash
task run
```

サーバーは `http://localhost:8080` で起動する。

## APIエンドポイント

| Method | Path | Description |
|---|---|---|
| GET | /api/todos | Todo一覧取得 |
| GET | /api/todos/:id | Todo個別取得 |
| POST | /api/todos | Todo作成 |
| PUT | /api/todos/:id | Todo更新 |
| DELETE | /api/todos/:id | Todo削除 |

### リクエスト例

```bash
# 作成
curl -X POST http://localhost:8080/api/todos \
  -H "Content-Type: application/json" \
  -d '{"title": "first todo", "description": "sample"}'

# 一覧取得
curl http://localhost:8080/api/todos

# 更新
curl -X PUT http://localhost:8080/api/todos/1 \
  -H "Content-Type: application/json" \
  -d '{"completed": true}'

# 削除
curl -X DELETE http://localhost:8080/api/todos/1
```

## タスク一覧

```bash
task build        # ビルド
task run          # 起動
task test         # テスト実行（カバレッジ付き）
task lint         # リント
task fmt          # フォーマット
task db-up        # PostgreSQL起動
task db-down      # PostgreSQL停止
task db-migrate   # マイグレーション実行
task db-rollback  # マイグレーション巻き戻し
task db-status    # マイグレーション状態確認
task wire         # Wire DIコード生成
```

## プロジェクト構成

```
backend-dev-val/
├── cmd/server/main.go          # エントリーポイント
├── internal/
│   ├── domain/                 # エンティティ + リポジトリIF + 独自エラー型
│   ├── usecase/                # ビジネスロジック
│   ├── handler/                # Gin HTTPハンドラー
│   ├── repository/             # PostgreSQLリポジトリ実装
│   └── di/                     # Wire DI
├── migrations/                 # gooseマイグレーション
├── docker-compose.yml
├── Taskfile.yml
├── .golangci.yml
└── .env.example
```

## 環境変数

| 変数 | デフォルト | 用途 |
|---|---|---|
| `DB_HOST` | `localhost` | PostgreSQLホスト |
| `DB_PORT` | `5432` | PostgreSQLポート |
| `DB_USER` | `postgres` | DBユーザー |
| `DB_PASSWORD` | `postgres` | DBパスワード |
| `DB_NAME` | `backend_dev_val` | DB名 |
| `DB_SSLMODE` | `disable` | SSLモード |
| `PORT` | `8080` | サーバーポート |
