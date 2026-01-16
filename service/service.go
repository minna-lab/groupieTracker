package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"groupie-tracker/api"
	"groupie-tracker/modele"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
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

// Geocodage avec cache
const fichierCacheGeocodage = "cache_geocodage.json"

type cacheGeocodage map[string]modele.Coordonnees
type reponseNominatim struct {
	Lat, Lon string `json:"lat,lon"`
}

func chargerCacheGeocodage() cacheGeocodage {
	var c cacheGeocodage
	if b, err := os.ReadFile(fichierCacheGeocodage); err == nil {
		json.Unmarshal(b, &c)
	}
	if c == nil {
		c = cacheGeocodage{}
	}
	return c
}

func sauverCacheGeocodage(c cacheGeocodage) {
	if b, err := json.MarshalIndent(c, "", "  "); err == nil {
		os.WriteFile(fichierCacheGeocodage, b, 0644)
	}
}

func normaliserLieu(lieu string) string {
	lieu = strings.ToLower(strings.TrimSpace(lieu))
	lieu = strings.ReplaceAll(strings.ReplaceAll(lieu, "-", ", "), "_", " ")
	return strings.Join(strings.Fields(lieu), " ")
}

func GeocoderLieu(lieu string, client *http.Client, cache cacheGeocodage) (modele.Coordonnees, bool, error) {
	lieu = normaliserLieu(lieu)
	if lieu == "" {
		return modele.Coordonnees{}, false, errors.New("lieu vide")
	}
	if v, ok := cache[lieu]; ok {
		return v, true, nil
	}

	req, _ := http.NewRequest("GET", "https://nominatim.openstreetmap.org/search?format=json&limit=1&q="+url.QueryEscape(lieu), nil)
	req.Header.Set("User-Agent", "groupie-tracker-fyne/1.0")

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return modele.Coordonnees{}, false, fmt.Errorf("erreur API: %v", err)
	}
	defer resp.Body.Close()

	var arr []struct {
		Lat string `json:"lat"`
		Lon string `json:"lon"`
	}
	if json.NewDecoder(resp.Body).Decode(&arr) != nil || len(arr) == 0 {
		return modele.Coordonnees{}, false, nil
	}

	var lat, lon float64
	fmt.Sscanf(arr[0].Lat, "%f", &lat)
	fmt.Sscanf(arr[0].Lon, "%f", &lon)

	coord := modele.Coordonnees{Lat: lat, Lon: lon}
	cache[lieu] = coord
	sauverCacheGeocodage(cache)
	time.Sleep(1100 * time.Millisecond)
	return coord, false, nil
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
