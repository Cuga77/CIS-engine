package search

import (
	"cis-engine/internal/storage"
	"context"
	"log"
)

type Service struct {
	storage storage.Storer
}

func NewService(s storage.Storer) *Service {
	return &Service{storage: s}
}

type Result struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

func (s *Service) Search(ctx context.Context, query string) ([]Result, error) {
	if query == "" {
		return []Result{}, nil
	}
	log.Printf("Searching for query: '%s'", query)

	pages, err := s.storage.SearchPages(ctx, query)
	if err != nil {
		return nil, err
	}

	if len(pages) == 0 {
		log.Printf("No results found for query: '%s'", query)
		return []Result{}, nil
	}

	results := make([]Result, 0, len(pages))
	for _, page := range pages {
		results = append(results, Result{
			URL:   page.URL,
			Title: page.Title,
		})
	}

	return results, nil
}

func (s *Service) ScheduleCrawl(ctx context.Context, url string) error {
	log.Printf("Received request to crawl URL: %s", url)

	page := &storage.Page{
		URL: url,
	}

	_, err := s.storage.StorePage(ctx, page)
	return err
}

func (s *Service) GetStats(ctx context.Context) (*storage.Metrics, error) {
	log.Println("Requesting system statistics")
	return s.storage.GetMetrics(ctx)
}