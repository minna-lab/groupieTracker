package service

import (
	"groupie-tracker/api"
	"groupie-tracker/modele"
)

// RecupererRelationAvecCache :
// 1) retourne depuis le cache si présent
// 2) sinon appelle l'API
// 3) enregistre dans le cache
func RecupererRelationAvecCache(cache *CacheRelations, artisteID int) (modele.Relation, error) {
	if rel, ok := cache.Get(artisteID); ok {
		return rel, nil
	}

	rel, err := api.RecupererRelation(artisteID)
	if err != nil {
		return modele.Relation{}, err
	}

	cache.Set(artisteID, rel)
	return rel, nil
}
