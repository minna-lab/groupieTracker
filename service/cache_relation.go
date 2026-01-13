package service

import (
	"groupie-tracker/modele"
	"sync"
)

type CacheRelations struct {
	mu   sync.RWMutex
	data map[int]modele.Relation
}

func NouveauCacheRelations() *CacheRelations {
	return &CacheRelations{
		data: make(map[int]modele.Relation),
	}
}

func (c *CacheRelations) Get(id int) (modele.Relation, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	rel, ok := c.data[id]
	return rel, ok
}

func (c *CacheRelations) Set(id int, rel modele.Relation) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[id] = rel
}
