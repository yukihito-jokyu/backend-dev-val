package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend-dev-val/internal/domain"
	"backend-dev-val/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) FindAll(ctx context.Context) ([]*domain.Todo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Todo), args.Error(1)
}

func (m *mockRepo) FindByID(ctx context.Context, id int64) (*domain.Todo, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Todo), args.Error(1)
}

func (m *mockRepo) Create(ctx context.Context, todo *domain.Todo) (*domain.Todo, error) {
	args := m.Called(ctx, todo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Todo), args.Error(1)
}

func (m *mockRepo) Update(ctx context.Context, todo *domain.Todo) (*domain.Todo, error) {
	args := m.Called(ctx, todo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Todo), args.Error(1)
}

func (m *mockRepo) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupRouter(repo *mockRepo) *gin.Engine {
	uc := usecase.NewTodoUsecase(repo)
	h := NewTodoHandler(uc)
	return SetupRouter(h)
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

func TestTodoHandler_List(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		mock       func(repo *mockRepo)
		wantStatus int
		wantBody   map[string]any
	}{
		{
			name: "正常系: Todo一覧を取得",
			mock: func(repo *mockRepo) {
				repo.On("FindAll", mock.Anything).Return([]*domain.Todo{
					{ID: 1, Title: "todo1", Description: "desc1", Completed: false, CreatedAt: now, UpdatedAt: now},
				}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "正常系: 空のリストを取得",
			mock: func(repo *mockRepo) {
				repo.On("FindAll", mock.Anything).Return([]*domain.Todo{}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "異常系: リポジトリエラーで500",
			mock: func(repo *mockRepo) {
				repo.On("FindAll", mock.Anything).Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockRepo)
			tt.mock(repo)
			router := setupRouter(repo)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			repo.AssertExpectations(t)
		})
	}
}

func TestTodoHandler_GetByID(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		id         string
		mock       func(repo *mockRepo)
		wantStatus int
	}{
		{
			name: "正常系: 存在するTodoを取得",
			id:   "1",
			mock: func(repo *mockRepo) {
				repo.On("FindByID", mock.Anything, int64(1)).Return(&domain.Todo{
					ID: 1, Title: "todo1", Description: "desc", Completed: false, CreatedAt: now, UpdatedAt: now,
				}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "異常系: 無効なIDで400",
			id:         "abc",
			mock:       func(repo *mockRepo) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "異常系: 存在しないTodoで404",
			id:   "999",
			mock: func(repo *mockRepo) {
				repo.On("FindByID", mock.Anything, int64(999)).Return(nil, &domain.NotFoundError{ID: 999})
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "異常系: リポジトリエラーで500",
			id:   "1",
			mock: func(repo *mockRepo) {
				repo.On("FindByID", mock.Anything, int64(1)).Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockRepo)
			tt.mock(repo)
			router := setupRouter(repo)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/todos/"+tt.id, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			repo.AssertExpectations(t)
		})
	}
}

func TestTodoHandler_Create(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		body       any
		mock       func(repo *mockRepo)
		wantStatus int
	}{
		{
			name: "正常系: Todoを作成",
			body: map[string]string{"title": "test todo", "description": "desc"},
			mock: func(repo *mockRepo) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Todo")).Return(&domain.Todo{
					ID: 1, Title: "test todo", Description: "desc", Completed: false, CreatedAt: now, UpdatedAt: now,
				}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "正常系: タイトルのみで作成",
			body: map[string]string{"title": "test todo"},
			mock: func(repo *mockRepo) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Todo")).Return(&domain.Todo{
					ID: 1, Title: "test todo", Description: "", Completed: false, CreatedAt: now, UpdatedAt: now,
				}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "異常系: 無効なJSONで400",
			body:       "invalid json",
			mock:       func(repo *mockRepo) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "異常系: タイトルなしで400",
			body:       map[string]string{"description": "desc"},
			mock:       func(repo *mockRepo) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "異常系: リポジトリエラーで500",
			body: map[string]string{"title": "test todo"},
			mock: func(repo *mockRepo) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Todo")).Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockRepo)
			tt.mock(repo)
			router := setupRouter(repo)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/todos", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			repo.AssertExpectations(t)
		})
	}
}

func TestTodoHandler_Update(t *testing.T) {
	now := time.Now()
	updatedTitle := "updated"

	tests := []struct {
		name       string
		id         string
		body       any
		mock       func(repo *mockRepo)
		wantStatus int
	}{
		{
			name: "正常系: Todoを更新",
			id:   "1",
			body: map[string]any{"title": "updated", "description": "new desc", "completed": true},
			mock: func(repo *mockRepo) {
				repo.On("FindByID", mock.Anything, int64(1)).Return(&domain.Todo{
					ID: 1, Title: "old", Description: "desc", Completed: false, CreatedAt: now, UpdatedAt: now,
				}, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Todo")).Return(&domain.Todo{
					ID: 1, Title: "updated", Description: "new desc", Completed: true, CreatedAt: now, UpdatedAt: now,
				}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "異常系: 無効なIDで400",
			id:         "abc",
			body:       map[string]any{"title": "updated"},
			mock:       func(repo *mockRepo) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "異常系: 無効なJSONで400",
			id:         "1",
			body:       "invalid json",
			mock:       func(repo *mockRepo) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "異常系: 存在しないTodoで404",
			id:   "999",
			body: map[string]any{"title": &updatedTitle},
			mock: func(repo *mockRepo) {
				repo.On("FindByID", mock.Anything, int64(999)).Return(nil, &domain.NotFoundError{ID: 999})
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "異常系: 空タイトルでバリデーションエラー400",
			id:   "1",
			body: map[string]any{"title": ""},
			mock: func(repo *mockRepo) {
				repo.On("FindByID", mock.Anything, int64(1)).Return(&domain.Todo{
					ID: 1, Title: "old", Description: "desc", Completed: false, CreatedAt: now, UpdatedAt: now,
				}, nil)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "異常系: リポジトリエラーで500",
			id:   "1",
			body: map[string]any{"title": "updated"},
			mock: func(repo *mockRepo) {
				repo.On("FindByID", mock.Anything, int64(1)).Return(&domain.Todo{
					ID: 1, Title: "old", Description: "desc", Completed: false, CreatedAt: now, UpdatedAt: now,
				}, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Todo")).Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockRepo)
			tt.mock(repo)
			router := setupRouter(repo)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/api/todos/"+tt.id, bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			repo.AssertExpectations(t)
		})
	}
}

func TestTodoHandler_Delete(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		mock       func(repo *mockRepo)
		wantStatus int
	}{
		{
			name: "正常系: Todoを削除",
			id:   "1",
			mock: func(repo *mockRepo) {
				repo.On("Delete", mock.Anything, int64(1)).Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "異常系: 無効なIDで400",
			id:         "abc",
			mock:       func(repo *mockRepo) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "異常系: 存在しないTodoで404",
			id:   "999",
			mock: func(repo *mockRepo) {
				repo.On("Delete", mock.Anything, int64(999)).Return(&domain.NotFoundError{ID: 999})
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "異常系: リポジトリエラーで500",
			id:   "1",
			mock: func(repo *mockRepo) {
				repo.On("Delete", mock.Anything, int64(1)).Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockRepo)
			tt.mock(repo)
			router := setupRouter(repo)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/api/todos/"+tt.id, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			repo.AssertExpectations(t)
		})
	}
}
