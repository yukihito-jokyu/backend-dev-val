package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"backend-dev-val/internal/domain"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		envOrDefault("DB_HOST", "localhost"),
		envOrDefault("DB_PORT", "5432"),
		envOrDefault("DB_USER", "postgres"),
		envOrDefault("DB_PASSWORD", "postgres"),
		envOrDefault("DB_NAME", "backend_dev_val_test"),
		envOrDefault("DB_SSLMODE", "disable"),
	)

	var err error
	testDB, err = sql.Open("postgres", dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open test db: %v\n", err)
		os.Exit(1)
	}

	if err := testDB.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to ping test db: %v\n", err)
		os.Exit(1)
	}

	if _, err := testDB.Exec(`
		CREATE TABLE IF NOT EXISTS todos (
			id BIGSERIAL PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			completed BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create todos table: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	_ = testDB.Close()
	os.Exit(code)
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func truncateTodos(t *testing.T) {
	t.Helper()
	_, err := testDB.Exec("TRUNCATE todos RESTART IDENTITY CASCADE")
	require.NoError(t, err)
}

func insertTestTodo(t *testing.T, title, description string, completed bool) *domain.Todo {
	t.Helper()
	var todo domain.Todo
	err := testDB.QueryRowContext(
		context.Background(),
		"INSERT INTO todos (title, description, completed) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at",
		title, description, completed,
	).Scan(&todo.ID, &todo.CreatedAt, &todo.UpdatedAt)
	require.NoError(t, err)
	todo.Title = title
	todo.Description = description
	todo.Completed = completed
	return &todo
}

func TestTodoRepository_FindAll(t *testing.T) {
	truncateTodos(t)
	repo := NewTodoRepository(testDB)
	ctx := context.Background()

	t.Run("正常系: 空のリストを取得", func(t *testing.T) {
		got, err := repo.FindAll(ctx)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("正常系: 複数のTodoを取得", func(t *testing.T) {
		insertTestTodo(t, "todo1", "desc1", false)
		insertTestTodo(t, "todo2", "desc2", true)

		got, err := repo.FindAll(ctx)
		require.NoError(t, err)
		require.Len(t, got, 2)

		assert.Equal(t, "todo1", got[0].Title)
		assert.Equal(t, "desc1", got[0].Description)
		assert.False(t, got[0].Completed)

		assert.Equal(t, "todo2", got[1].Title)
		assert.Equal(t, "desc2", got[1].Description)
		assert.True(t, got[1].Completed)
	})
}

func TestTodoRepository_FindByID(t *testing.T) {
	truncateTodos(t)
	repo := NewTodoRepository(testDB)
	ctx := context.Background()

	t.Run("正常系: 存在するTodoを取得", func(t *testing.T) {
		inserted := insertTestTodo(t, "todo1", "desc1", false)

		got, err := repo.FindByID(ctx, inserted.ID)
		require.NoError(t, err)
		assert.Equal(t, inserted.ID, got.ID)
		assert.Equal(t, "todo1", got.Title)
		assert.Equal(t, "desc1", got.Description)
		assert.False(t, got.Completed)
		assert.NotZero(t, got.CreatedAt)
		assert.NotZero(t, got.UpdatedAt)
	})

	t.Run("異常系: 存在しないTodoでNotFoundError", func(t *testing.T) {
		got, err := repo.FindByID(ctx, 999)
		assert.Nil(t, got)
		require.Error(t, err)
		assert.IsType(t, &domain.NotFoundError{}, err)
	})
}

func TestTodoRepository_Create(t *testing.T) {
	truncateTodos(t)
	repo := NewTodoRepository(testDB)
	ctx := context.Background()

	t.Run("正常系: タイトルのみで作成", func(t *testing.T) {
		todo := &domain.Todo{Title: "new todo", Description: "", Completed: false}

		got, err := repo.Create(ctx, todo)
		require.NoError(t, err)
		assert.NotZero(t, got.ID)
		assert.Equal(t, "new todo", got.Title)
		assert.Empty(t, got.Description)
		assert.False(t, got.Completed)
		assert.NotZero(t, got.CreatedAt)
		assert.NotZero(t, got.UpdatedAt)
	})

	t.Run("正常系: タイトルと説明で作成", func(t *testing.T) {
		todo := &domain.Todo{Title: "new todo", Description: "description", Completed: true}

		got, err := repo.Create(ctx, todo)
		require.NoError(t, err)
		assert.NotZero(t, got.ID)
		assert.Equal(t, "description", got.Description)
		assert.True(t, got.Completed)
	})
}

func TestTodoRepository_Update(t *testing.T) {
	truncateTodos(t)
	repo := NewTodoRepository(testDB)
	ctx := context.Background()

	t.Run("正常系: 全フィールドを更新", func(t *testing.T) {
		inserted := insertTestTodo(t, "old", "old desc", false)
		originalCreatedAt := inserted.CreatedAt

		inserted.Title = "updated"
		inserted.Description = "new desc"
		inserted.Completed = true

		got, err := repo.Update(ctx, inserted)
		require.NoError(t, err)
		assert.Equal(t, "updated", got.Title)
		assert.Equal(t, "new desc", got.Description)
		assert.True(t, got.Completed)
		assert.Equal(t, originalCreatedAt, got.CreatedAt)
		assert.True(t, !got.UpdatedAt.Before(originalCreatedAt))
	})

	t.Run("異常系: 存在しないTodoでNotFoundError", func(t *testing.T) {
		todo := &domain.Todo{ID: 999, Title: "not found", Description: "", Completed: false}

		got, err := repo.Update(ctx, todo)
		assert.Nil(t, got)
		require.Error(t, err)
		assert.IsType(t, &domain.NotFoundError{}, err)
	})
}

func TestTodoRepository_Delete(t *testing.T) {
	truncateTodos(t)
	repo := NewTodoRepository(testDB)
	ctx := context.Background()

	t.Run("正常系: 削除成功", func(t *testing.T) {
		inserted := insertTestTodo(t, "to delete", "desc", false)

		err := repo.Delete(ctx, inserted.ID)
		require.NoError(t, err)

		got, err := repo.FindByID(ctx, inserted.ID)
		assert.Nil(t, got)
		assert.IsType(t, &domain.NotFoundError{}, err)
	})

	t.Run("異常系: 存在しないTodoでNotFoundError", func(t *testing.T) {
		err := repo.Delete(ctx, 999)
		require.Error(t, err)
		assert.IsType(t, &domain.NotFoundError{}, err)
	})
}
