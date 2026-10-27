package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/edddddr/taskmanager/internal/domain"
)

	
func respondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	case errors.Is(err, domain.ErrNotValidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
	default:
		// Log the real cause; never leak internals to the client.
		log.Printf("internal error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}