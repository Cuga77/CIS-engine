package indexer

import (
	"cis-engine/internal/storage"
	"context"
	"log"
	"time"
)

type Indexer struct {
	storage  storage.Storer
	ticker   *time.Ticker
	doneChan chan bool
}

func NewIndexer(s storage.Storer, interval time.Duration) *Indexer {
	return &Indexer{
		storage:  s,
		ticker:   time.NewTicker(interval),
		doneChan: make(chan bool),
	}
}

func (i *Indexer) Start(ctx context.Context) {
	log.Println("Сервис индексации запущен.")
	for {
		select {
		case <-i.doneChan:
			log.Println("Сервис индексации остановлен.")
			return
		case <-i.ticker.C:
			err := i.indexNextPage(ctx)
			if err != nil {
				log.Printf("Ошибка при индексации страницы: %v", err)
			}
		}
	}
}

func (i *Indexer) Stop() {
	i.ticker.Stop()
	i.doneChan <- true
}

func (i *Indexer) indexNextPage(ctx context.Context) error {
	page, err := i.storage.GetNextPageToIndex(ctx)
	if err != nil {
		return err
	}

	if page == nil {
		return nil
	}

	log.Printf("Индексируется страница #%d (%s)", page.ID, page.URL)

	return i.storage.UpdatePageVector(ctx, page)
}
