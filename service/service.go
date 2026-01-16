package service

import (
	"encoding/json"
	"groupie-tracker/api"
	"groupie-tracker/modele"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Cache thread-safe pour les relations
type CacheRelations struct {
	mu   sync.RWMutex
	data map[int]modele.Relation
}

func NouveauCacheRelations() *CacheRelations {
	return &CacheRelations{data: make(map[int]modele.Relation)}
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

func RecupererRelationAvecCache(cache *CacheRelations, artisteID int) (modele.Relation, error) {
	if rel, ok := cache.Get(artisteID); ok {
		return rel, nil
	}
	rel, err := api.RecupererRelation(artisteID)
	if err == nil {
		cache.Set(artisteID, rel)
	}
	return rel, err
}

// Gestionnaire de favoris (singleton)
type GestionnaireFavoris struct {
	mutex   sync.RWMutex
	favoris map[int]modele.Artiste
	fichier string
}

var (
	instance *GestionnaireFavoris
	once     sync.Once
)

func ObtenirGestionnaireFavoris() *GestionnaireFavoris {
	once.Do(func() {
		instance = &GestionnaireFavoris{favoris: make(map[int]modele.Artiste), fichier: "favoris.json"}
		instance.charger()
	})
	return instance
}

func (g *GestionnaireFavoris) EstFavori(id int) bool {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	_, existe := g.favoris[id]
	return existe
}

func (g *GestionnaireFavoris) Basculer(artiste modele.Artiste) bool {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	_, existe := g.favoris[artiste.ID]
	if existe {
		delete(g.favoris, artiste.ID)
	} else {
		g.favoris[artiste.ID] = artiste
	}
	g.sauvegarder()
	return !existe
}

func (g *GestionnaireFavoris) ObtenirFavoris() []modele.Artiste {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	liste := make([]modele.Artiste, 0, len(g.favoris))
	for _, artiste := range g.favoris {
		liste = append(liste, artiste)
	}
	return liste
}

func (g *GestionnaireFavoris) sauvegarder() {
	if data, err := json.MarshalIndent(g.favoris, "", "  "); err == nil {
		os.WriteFile(g.fichier, data, 0644)
	}
}

func (g *GestionnaireFavoris) charger() {
	if data, err := os.ReadFile(g.fichier); err == nil {
		json.Unmarshal(data, &g.favoris)
	}
}

func normaliserLieu(lieu string) string {
	lieu = strings.ToLower(strings.TrimSpace(lieu))
	lieu = strings.ReplaceAll(strings.ReplaceAll(lieu, "-", ", "), "_", " ")
	return strings.Join(strings.Fields(lieu), " ")
}

// Construction des suggestions de recherche
func ConstruireSuggestions(artistes []modele.Artiste, cache *CacheRelations) []modele.Suggestion {
	vu := make(map[string]bool)
	var res []modele.Suggestion
	ajouter := func(texte, typ string, id int) {
		if cle := strings.ToLower(texte) + "|" + typ; texte != "" && !vu[cle] {
			vu[cle] = true
			res = append(res, modele.Suggestion{texte, typ, id})
		}
	}
	for _, a := range artistes {
		ajouter(a.Nom, "artiste", a.ID)
		for _, m := range a.Membres {
			ajouter(m, "membre", a.ID)
		}
		ajouter(strconv.Itoa(a.AnneeCreation), "année de création", a.ID)
		ajouter(a.PremierAlbum, "premier album", a.ID)
		if cache != nil {
			if rel, ok := cache.Get(a.ID); ok {
				for lieu := range rel.DatesParLieu {
					ajouter(lieu, "lieu", a.ID)
				}
			}
		}
	}
	return res
}
