package api

import (
	"cis-engine/internal/search"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	searchService *search.Service
}

func NewHandler(ss *search.Service) *Handler {
	return &Handler{searchService: ss}
}

func NewRouter(h *Handler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	apiV1 := router.Group("/api/v1")
	{
		apiV1.GET("/search", h.searchHandler)
	}

	return router
}

func (h *Handler) searchHandler(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Параметр 'q' не может быть пустым"})
		return
	}

	results, err := h.searchService.Search(c.Request.Context(), query)
	if err != nil {
		log.Printf("ERROR: search service failed for query '%s': %v", query, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка сервера"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"query":   query,
		"results": results,
	})
}
