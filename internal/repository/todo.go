package repository

import (
	"context"
	"database/sql"

	"backend-dev-val/internal/domain"
)

type todoRepository struct {
	db *sql.DB
}

func NewTodoRepository(db *sql.DB) domain.TodoRepository {
	return &todoRepository{db: db}
}

func (r *todoRepository) FindAll(ctx context.Context) ([]*domain.Todo, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, title, description, completed, created_at, updated_at FROM todos ORDER BY id",
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var todos []*domain.Todo
	for rows.Next() {
		todo, err := scanTodo(rows)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}
	return todos, rows.Err()
}

func (r *todoRepository) FindByID(ctx context.Context, id int64) (*domain.Todo, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT id, title, description, completed, created_at, updated_at FROM todos WHERE id = $1",
		id,
	)
	todo, err := scanTodo(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &domain.NotFoundError{ID: id}
		}
		return nil, err
	}
	return todo, nil
}

func (r *todoRepository) Create(ctx context.Context, todo *domain.Todo) (*domain.Todo, error) {
	err := r.db.QueryRowContext(ctx,
		"INSERT INTO todos (title, description, completed) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at",
		todo.Title, todo.Description, todo.Completed,
	).Scan(&todo.ID, &todo.CreatedAt, &todo.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return todo, nil
}

func (r *todoRepository) Update(ctx context.Context, todo *domain.Todo) (*domain.Todo, error) {
	err := r.db.QueryRowContext(ctx,
		"UPDATE todos SET title = $1, description = $2, completed = $3, updated_at = NOW() WHERE id = $4 RETURNING updated_at",
		todo.Title, todo.Description, todo.Completed, todo.ID,
	).Scan(&todo.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &domain.NotFoundError{ID: todo.ID}
		}
		return nil, err
	}
	return todo, nil
}

func (r *todoRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM todos WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return &domain.NotFoundError{ID: id}
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTodo(s scanner) (*domain.Todo, error) {
	var todo domain.Todo
	err := s.Scan(&todo.ID, &todo.Title, &todo.Description, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
	return &todo, err
}
