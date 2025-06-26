// internal/api/api.go
package api

import (
	"context"
	"log"
	"net/http"

	"cis-engine/internal/search"
	"cis-engine/internal/storage"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Searcher interface {
	Search(ctx context.Context, query string) ([]search.Result, error)
	ScheduleCrawl(ctx context.Context, url string) error
	GetStats(ctx context.Context) (*storage.Metrics, error)
}

type Handler struct {
	searchService Searcher
}

func NewHandler(s Searcher) *Handler {
	return &Handler{searchService: s}
}

func NewRouter(h *Handler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.Use(cors.Default())
	router.LoadHTMLGlob("frontend/*.html")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	apiV1 := router.Group("/api/v1")
	{
		apiV1.GET("/search", h.searchHandler)
		apiV1.POST("/crawl", h.crawlHandler)
		apiV1.GET("/status", h.statusHandler)
	}

	return router
}

func (h *Handler) searchHandler(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		return
	}

	results, err := h.searchService.Search(c.Request.Context(), query)
	if err != nil {
		log.Printf("ERROR: search service failed for query '%s': %v", query, err)
		c.String(http.StatusInternalServerError, "Ошибка сервера при поиске.")
		return
	}

	c.HTML(http.StatusOK, "results.html", gin.H{
		"Results": results,
	})
}

func (h *Handler) crawlHandler(c *gin.Context) {
	var request struct {
		URL string `json:"url"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса."})
		return
	}
	if request.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Поле 'url' не может быть пустым."})
		return
	}
	if err := h.searchService.ScheduleCrawl(c.Request.Context(), request.URL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось добавить URL в очередь"})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"message": "URL принят в очередь на сканирование."})
}

func (h *Handler) statusHandler(c *gin.Context) {
	stats, err := h.searchService.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить статистику"})
		return
	}
	c.JSON(http.StatusOK, stats)
}
