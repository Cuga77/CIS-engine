package search

import (
	"cis-engine/internal/indexer"
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

	tokens := indexer.Tokenize(query)
	if len(tokens) == 0 {
		return []Result{}, nil
	}
	log.Printf("Поисковый запрос '%s' токенизирован в: %v", query, tokens)

	docIDCounts := make(map[int64]int)
	for _, token := range tokens {
		docIDs, err := s.storage.FindDocumentsByTerm(ctx, token)
		if err != nil {
			return nil, err
		}
		for _, id := range docIDs {
			docIDCounts[id]++
		}
	}

	var finalDocIDs []int64
	for id, count := range docIDCounts {
		if count == len(tokens) {
			finalDocIDs = append(finalDocIDs, id)
		}
	}

	if len(finalDocIDs) == 0 {
		log.Printf("Не найдено документов, содержащих все токены для запроса '%s'", query)
		return []Result{}, nil
	}

	pages, err := s.storage.GetPagesByIDs(ctx, finalDocIDs)
	if err != nil {
		return nil, err
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
