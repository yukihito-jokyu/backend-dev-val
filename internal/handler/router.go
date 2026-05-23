package handler

import "github.com/gin-gonic/gin"

func SetupRouter(th *TodoHandler) *gin.Engine {
	r := gin.Default()
	api := r.Group("/api")
	{
		api.GET("/todos", th.List)
		api.GET("/todos/:id", th.GetByID)
		api.POST("/todos", th.Create)
		api.PUT("/todos/:id", th.Update)
		api.DELETE("/todos/:id", th.Delete)
	}
	return r
}
