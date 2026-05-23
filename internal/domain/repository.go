package domain

import "context"

type TodoRepository interface {
	FindAll(ctx context.Context) ([]*Todo, error)
	FindByID(ctx context.Context, id int64) (*Todo, error)
	Create(ctx context.Context, todo *Todo) (*Todo, error)
	Update(ctx context.Context, todo *Todo) (*Todo, error)
	Delete(ctx context.Context, id int64) error
}
