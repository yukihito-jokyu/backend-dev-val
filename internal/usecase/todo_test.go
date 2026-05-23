package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend-dev-val/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockTodoRepository struct {
	mock.Mock
}

func (m *mockTodoRepository) FindAll(ctx context.Context) ([]*domain.Todo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Todo), args.Error(1)
}

func (m *mockTodoRepository) FindByID(ctx context.Context, id int64) (*domain.Todo, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Todo), args.Error(1)
}

func (m *mockTodoRepository) Create(ctx context.Context, todo *domain.Todo) (*domain.Todo, error) {
	args := m.Called(ctx, todo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Todo), args.Error(1)
}

func (m *mockTodoRepository) Update(ctx context.Context, todo *domain.Todo) (*domain.Todo, error) {
	args := m.Called(ctx, todo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Todo), args.Error(1)
}

func (m *mockTodoRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestTodoUsecase_List(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name    string
		mock    func(repo *mockTodoRepository)
		want    []*domain.Todo
		wantErr bool
	}{
		{
			name: "正常系: 複数のTodoを取得",
			mock: func(repo *mockTodoRepository) {
				repo.On("FindAll", ctx).Return([]*domain.Todo{
					{ID: 1, Title: "todo1", Completed: false, CreatedAt: now, UpdatedAt: now},
					{ID: 2, Title: "todo2", Completed: true, CreatedAt: now, UpdatedAt: now},
				}, nil)
			},
			want: []*domain.Todo{
				{ID: 1, Title: "todo1", Completed: false, CreatedAt: now, UpdatedAt: now},
				{ID: 2, Title: "todo2", Completed: true, CreatedAt: now, UpdatedAt: now},
			},
			wantErr: false,
		},
		{
			name: "正常系: 空のリストを取得",
			mock: func(repo *mockTodoRepository) {
				repo.On("FindAll", ctx).Return([]*domain.Todo{}, nil)
			},
			want:    []*domain.Todo{},
			wantErr: false,
		},
		{
			name: "異常系: リポジトリエラー",
			mock: func(repo *mockTodoRepository) {
				repo.On("FindAll", ctx).Return(nil, errors.New("db error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockTodoRepository)
			tt.mock(repo)
			uc := NewTodoUsecase(repo)

			got, err := uc.List(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			repo.AssertExpectations(t)
		})
	}
}

func TestTodoUsecase_GetByID(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name    string
		id      int64
		mock    func(repo *mockTodoRepository)
		want    *domain.Todo
		wantErr bool
	}{
		{
			name: "正常系: 存在するTodoを取得",
			id:   1,
			mock: func(repo *mockTodoRepository) {
				repo.On("FindByID", ctx, int64(1)).Return(&domain.Todo{
					ID: 1, Title: "todo1", Completed: false, CreatedAt: now, UpdatedAt: now,
				}, nil)
			},
			want: &domain.Todo{
				ID: 1, Title: "todo1", Completed: false, CreatedAt: now, UpdatedAt: now,
			},
			wantErr: false,
		},
		{
			name: "異常系: 存在しないTodo",
			id:   999,
			mock: func(repo *mockTodoRepository) {
				repo.On("FindByID", ctx, int64(999)).Return(nil, &domain.NotFoundError{ID: 999})
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "異常系: リポジトリエラー",
			id:   1,
			mock: func(repo *mockTodoRepository) {
				repo.On("FindByID", ctx, int64(1)).Return(nil, errors.New("db error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockTodoRepository)
			tt.mock(repo)
			uc := NewTodoUsecase(repo)

			got, err := uc.GetByID(ctx, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			repo.AssertExpectations(t)
		})
	}
}

func TestTodoUsecase_Create(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name    string
		input   CreateTodoInput
		mock    func(repo *mockTodoRepository)
		want    *domain.Todo
		wantErr error
	}{
		{
			name:  "正常系: タイトルのみで作成",
			input: CreateTodoInput{Title: "test todo"},
			mock: func(repo *mockTodoRepository) {
				repo.On("Create", ctx, &domain.Todo{
					Title: "test todo", Description: "", Completed: false,
				}).Return(&domain.Todo{
					ID: 1, Title: "test todo", Description: "", Completed: false, CreatedAt: now, UpdatedAt: now,
				}, nil)
			},
			want: &domain.Todo{
				ID: 1, Title: "test todo", Description: "", Completed: false, CreatedAt: now, UpdatedAt: now,
			},
			wantErr: nil,
		},
		{
			name:  "正常系: タイトルと説明で作成",
			input: CreateTodoInput{Title: "test todo", Description: "description"},
			mock: func(repo *mockTodoRepository) {
				repo.On("Create", ctx, &domain.Todo{
					Title: "test todo", Description: "description", Completed: false,
				}).Return(&domain.Todo{
					ID: 1, Title: "test todo", Description: "description", Completed: false, CreatedAt: now, UpdatedAt: now,
				}, nil)
			},
			want: &domain.Todo{
				ID: 1, Title: "test todo", Description: "description", Completed: false, CreatedAt: now, UpdatedAt: now,
			},
			wantErr: nil,
		},
		{
			name:    "異常系: 空のタイトル",
			input:   CreateTodoInput{Title: ""},
			mock:    func(repo *mockTodoRepository) {},
			want:    nil,
			wantErr: &domain.ValidationError{Field: "title", Message: "title is required"},
		},
		{
			name:  "異常系: リポジトリエラー",
			input: CreateTodoInput{Title: "test todo"},
			mock: func(repo *mockTodoRepository) {
				repo.On("Create", ctx, &domain.Todo{
					Title: "test todo", Description: "", Completed: false,
				}).Return(nil, errors.New("db error"))
			},
			want:    nil,
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockTodoRepository)
			tt.mock(repo)
			uc := NewTodoUsecase(repo)

			got, err := uc.Create(ctx, tt.input)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			repo.AssertExpectations(t)
		})
	}
}

func TestTodoUsecase_Update(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	updatedTitle := "updated"
	updatedDesc := "updated desc"
	completed := true
	emptyTitle := ""

	tests := []struct {
		name    string
		id      int64
		input   UpdateTodoInput
		mock    func(repo *mockTodoRepository)
		want    *domain.Todo
		wantErr error
	}{
		{
			name:  "正常系: タイトルを更新",
			id:    1,
			input: UpdateTodoInput{Title: &updatedTitle},
			mock: func(repo *mockTodoRepository) {
				repo.On("FindByID", ctx, int64(1)).Return(&domain.Todo{
					ID: 1, Title: "old", Description: "desc", Completed: false, CreatedAt: now, UpdatedAt: now,
				}, nil)
				repo.On("Update", ctx, &domain.Todo{
					ID: 1, Title: "updated", Description: "desc", Completed: false, CreatedAt: now, UpdatedAt: now,
				}).Return(&domain.Todo{
					ID: 1, Title: "updated", Description: "desc", Completed: false, CreatedAt: now, UpdatedAt: now,
				}, nil)
			},
			want: &domain.Todo{
				ID: 1, Title: "updated", Description: "desc", Completed: false, CreatedAt: now, UpdatedAt: now,
			},
			wantErr: nil,
		},
		{
			name:  "正常系: 説明を更新",
			id:    1,
			input: UpdateTodoInput{Description: &updatedDesc},
			mock: func(repo *mockTodoRepository) {
				repo.On("FindByID", ctx, int64(1)).Return(&domain.Todo{
					ID: 1, Title: "todo", Description: "old desc", Completed: false, CreatedAt: now, UpdatedAt: now,
				}, nil)
				repo.On("Update", ctx, &domain.Todo{
					ID: 1, Title: "todo", Description: "updated desc", Completed: false, CreatedAt: now, UpdatedAt: now,
				}).Return(&domain.Todo{
					ID: 1, Title: "todo", Description: "updated desc", Completed: false, CreatedAt: now, UpdatedAt: now,
				}, nil)
			},
			want: &domain.Todo{
				ID: 1, Title: "todo", Description: "updated desc", Completed: false, CreatedAt: now, UpdatedAt: now,
			},
			wantErr: nil,
		},
		{
			name:  "正常系: 完了状態を更新",
			id:    1,
			input: UpdateTodoInput{Completed: &completed},
			mock: func(repo *mockTodoRepository) {
				repo.On("FindByID", ctx, int64(1)).Return(&domain.Todo{
					ID: 1, Title: "todo", Description: "desc", Completed: false, CreatedAt: now, UpdatedAt: now,
				}, nil)
				repo.On("Update", ctx, &domain.Todo{
					ID: 1, Title: "todo", Description: "desc", Completed: true, CreatedAt: now, UpdatedAt: now,
				}).Return(&domain.Todo{
					ID: 1, Title: "todo", Description: "desc", Completed: true, CreatedAt: now, UpdatedAt: now,
				}, nil)
			},
			want: &domain.Todo{
				ID: 1, Title: "todo", Description: "desc", Completed: true, CreatedAt: now, UpdatedAt: now,
			},
			wantErr: nil,
		},
		{
			name:  "正常系: 全フィールドを更新",
			id:    1,
			input: UpdateTodoInput{Title: &updatedTitle, Description: &updatedDesc, Completed: &completed},
			mock: func(repo *mockTodoRepository) {
				repo.On("FindByID", ctx, int64(1)).Return(&domain.Todo{
					ID: 1, Title: "old", Description: "old desc", Completed: false, CreatedAt: now, UpdatedAt: now,
				}, nil)
				repo.On("Update", ctx, &domain.Todo{
					ID: 1, Title: "updated", Description: "updated desc", Completed: true, CreatedAt: now, UpdatedAt: now,
				}).Return(&domain.Todo{
					ID: 1, Title: "updated", Description: "updated desc", Completed: true, CreatedAt: now, UpdatedAt: now,
				}, nil)
			},
			want: &domain.Todo{
				ID: 1, Title: "updated", Description: "updated desc", Completed: true, CreatedAt: now, UpdatedAt: now,
			},
			wantErr: nil,
		},
		{
			name:  "正常系: フィールドなしで更新（変更なし）",
			id:    1,
			input: UpdateTodoInput{},
			mock: func(repo *mockTodoRepository) {
				repo.On("FindByID", ctx, int64(1)).Return(&domain.Todo{
					ID: 1, Title: "todo", Description: "desc", Completed: false, CreatedAt: now, UpdatedAt: now,
				}, nil)
				repo.On("Update", ctx, &domain.Todo{
					ID: 1, Title: "todo", Description: "desc", Completed: false, CreatedAt: now, UpdatedAt: now,
				}).Return(&domain.Todo{
					ID: 1, Title: "todo", Description: "desc", Completed: false, CreatedAt: now, UpdatedAt: now,
				}, nil)
			},
			want: &domain.Todo{
				ID: 1, Title: "todo", Description: "desc", Completed: false, CreatedAt: now, UpdatedAt: now,
			},
			wantErr: nil,
		},
		{
			name:  "異常系: 存在しないTodo",
			id:    999,
			input: UpdateTodoInput{Title: &updatedTitle},
			mock: func(repo *mockTodoRepository) {
				repo.On("FindByID", ctx, int64(999)).Return(nil, &domain.NotFoundError{ID: 999})
			},
			want:    nil,
			wantErr: &domain.NotFoundError{ID: 999},
		},
		{
			name:  "異常系: 空のタイトルでバリデーションエラー",
			id:    1,
			input: UpdateTodoInput{Title: &emptyTitle},
			mock: func(repo *mockTodoRepository) {
				repo.On("FindByID", ctx, int64(1)).Return(&domain.Todo{
					ID: 1, Title: "todo", Description: "desc", Completed: false, CreatedAt: now, UpdatedAt: now,
				}, nil)
			},
			want:    nil,
			wantErr: &domain.ValidationError{Field: "title", Message: "title must not be empty"},
		},
		{
			name:  "異常系: 更新時のリポジトリエラー",
			id:    1,
			input: UpdateTodoInput{Title: &updatedTitle},
			mock: func(repo *mockTodoRepository) {
				repo.On("FindByID", ctx, int64(1)).Return(&domain.Todo{
					ID: 1, Title: "old", Description: "desc", Completed: false, CreatedAt: now, UpdatedAt: now,
				}, nil)
				repo.On("Update", ctx, &domain.Todo{
					ID: 1, Title: "updated", Description: "desc", Completed: false, CreatedAt: now, UpdatedAt: now,
				}).Return(nil, errors.New("db error"))
			},
			want:    nil,
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockTodoRepository)
			tt.mock(repo)
			uc := NewTodoUsecase(repo)

			got, err := uc.Update(ctx, tt.id, tt.input)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			repo.AssertExpectations(t)
		})
	}
}

func TestTodoUsecase_Delete(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		id      int64
		mock    func(repo *mockTodoRepository)
		wantErr bool
	}{
		{
			name: "正常系: 削除成功",
			id:   1,
			mock: func(repo *mockTodoRepository) {
				repo.On("Delete", ctx, int64(1)).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "異常系: 存在しないTodo",
			id:   999,
			mock: func(repo *mockTodoRepository) {
				repo.On("Delete", ctx, int64(999)).Return(&domain.NotFoundError{ID: 999})
			},
			wantErr: true,
		},
		{
			name: "異常系: リポジトリエラー",
			id:   1,
			mock: func(repo *mockTodoRepository) {
				repo.On("Delete", ctx, int64(1)).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockTodoRepository)
			tt.mock(repo)
			uc := NewTodoUsecase(repo)

			err := uc.Delete(ctx, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			repo.AssertExpectations(t)
		})
	}
}
