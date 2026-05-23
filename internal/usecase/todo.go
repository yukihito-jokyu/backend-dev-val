package usecase

import (
	"context"

	"backend-dev-val/internal/domain"
)

type TodoUsecase struct {
	repo domain.TodoRepository
}

func NewTodoUsecase(repo domain.TodoRepository) *TodoUsecase {
	return &TodoUsecase{repo: repo}
}

func (u *TodoUsecase) List(ctx context.Context) ([]*domain.Todo, error) {
	return u.repo.FindAll(ctx)
}

func (u *TodoUsecase) GetByID(ctx context.Context, id int64) (*domain.Todo, error) {
	return u.repo.FindByID(ctx, id)
}

type CreateTodoInput struct {
	Title       string
	Description string
}

func (u *TodoUsecase) Create(ctx context.Context, input CreateTodoInput) (*domain.Todo, error) {
	if input.Title == "" {
		return nil, &domain.ValidationError{Field: "title", Message: "title is required"}
	}
	todo := &domain.Todo{
		Title:       input.Title,
		Description: input.Description,
		Completed:   false,
	}
	return u.repo.Create(ctx, todo)
}

type UpdateTodoInput struct {
	Title       *string
	Description *string
	Completed   *bool
}

func (u *TodoUsecase) Update(ctx context.Context, id int64, input UpdateTodoInput) (*domain.Todo, error) {
	todo, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.Title != nil {
		if *input.Title == "" {
			return nil, &domain.ValidationError{Field: "title", Message: "title must not be empty"}
		}
		todo.Title = *input.Title
	}
	if input.Description != nil {
		todo.Description = *input.Description
	}
	if input.Completed != nil {
		todo.Completed = *input.Completed
	}
	return u.repo.Update(ctx, todo)
}

func (u *TodoUsecase) Delete(ctx context.Context, id int64) error {
	return u.repo.Delete(ctx, id)
}
