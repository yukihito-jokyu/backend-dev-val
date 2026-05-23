//go:build wireinject

package di

import (
	"database/sql"

	"backend-dev-val/internal/handler"
	"backend-dev-val/internal/repository"
	"backend-dev-val/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitializeApp(db *sql.DB) *gin.Engine {
	wire.Build(
		repository.NewTodoRepository,
		usecase.NewTodoUsecase,
		handler.NewTodoHandler,
		handler.SetupRouter,
	)
	return nil
}
