package server

import (
	"cis-engine/internal/search"
	"cis-engine/internal/storage"
	"context"
	"log"
	"net/http"
	"net/url"

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

	router.Use(errorHandler)

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' cannot be empty"})
		return
	}

	results, err := h.searchService.Search(c.Request.Context(), query)
	if err != nil {
		log.Printf("ERROR: search service failed for query '%s': %v", query, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format, expected JSON with 'url' field"})
		return
	}
	if request.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Поле 'url' не может быть пустым."})

		if _, err := url.ParseRequestURI(request.URL); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid URL format"})
			return
		}
		if err := h.searchService.ScheduleCrawl(c.Request.Context(), request.URL); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось добавить URL в очередь"})
			log.Printf("ERROR: failed to schedule crawl for url '%s': %v", request.URL, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add URL to the queue"})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"message": "URL принят в очередь на сканирование."})

		c.JSON(http.StatusAccepted, gin.H{"message": "URL accepted for scanning"})
	}
}

func (h *Handler) statusHandler(c *gin.Context) {
	stats, err := h.searchService.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get system statistics"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func errorHandler(c *gin.Context) {
	c.Next()

	if len(c.Errors) > 0 {
		// Log errors
		for _, e := range c.Errors {
			log.Printf("ERROR: %v", e.Err)
		}

		// Return a generic error message
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
	}
}
