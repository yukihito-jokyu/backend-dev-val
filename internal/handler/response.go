package handler

import (
	"net/http"

	"backend-dev-val/internal/domain"

	"github.com/gin-gonic/gin"
)

func respondError(c *gin.Context, err error) {
	switch err.(type) {
	case *domain.ValidationError:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case *domain.NotFoundError:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
