package crawler

import "sync"

type VisitedCache struct {
	mu    sync.Mutex
	items map[string]bool
}

func NewVisitedCache() *VisitedCache {
	return &VisitedCache{
		items: make(map[string]bool),
	}
}

func (vc *VisitedCache) AddIfNotExists(url string) bool {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	_, found := vc.items[url]
	if found {
		return false
	}

	vc.items[url] = true
	return true
}
