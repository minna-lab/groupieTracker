package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"groupie-tracker/modele"
)

const fichierCacheGeocodage = "cache_geocodage.json"

// Cache simple : "paris, france" -> Coordonnees
type cacheGeocodage map[string]modele.Coordonnees

func chargerCacheGeocodage() cacheGeocodage {
	b, err := os.ReadFile(fichierCacheGeocodage)
	if err != nil {
		return cacheGeocodage{}
	}
	var c cacheGeocodage
	if json.Unmarshal(b, &c) != nil {
		return cacheGeocodage{}
	}
	return c
}

func sauverCacheGeocodage(c cacheGeocodage) {
	b, _ := json.MarshalIndent(c, "", "  ")
	_ = os.WriteFile(fichierCacheGeocodage, b, 0644)
}

// Normalise un lieu "paris-france" -> "paris, france"
func normaliserLieu(lieu string) string {
	lieu = strings.TrimSpace(strings.ToLower(lieu))
	lieu = strings.ReplaceAll(lieu, "-", ", ")
	lieu = strings.ReplaceAll(lieu, "_", " ")
	lieu = strings.Join(strings.Fields(lieu), " ")
	return lieu
}

type reponseNominatim struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

// GeocoderLieu :
// - respecte la politique Nominatim : User-Agent + 1 req/sec :contentReference[oaicite:2]{index=2}
func GeocoderLieu(lieu string, client *http.Client, cache cacheGeocodage) (modele.Coordonnees, bool, error) {
	lieu = normaliserLieu(lieu)
	if lieu == "" {
		return modele.Coordonnees{}, false, errors.New("lieu vide")
	}

	// 1) Cache
	if v, ok := cache[lieu]; ok {
		return v, true, nil
	}

	// 2) API Nominatim
	u := "https://nominatim.openstreetmap.org/search?format=json&limit=1&q=" + url.QueryEscape(lieu)

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return modele.Coordonnees{}, false, err
	}

	// User-Agent obligatoire (sinon blocage possible) :contentReference[oaicite:3]{index=3}
	req.Header.Set("User-Agent", "groupie-tracker-fyne/1.0 (contact: student-project)")

	resp, err := client.Do(req)
	if err != nil {
		return modele.Coordonnees{}, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return modele.Coordonnees{}, false, fmt.Errorf("nominatim status %d", resp.StatusCode)
	}

	var arr []reponseNominatim
	if err := json.NewDecoder(resp.Body).Decode(&arr); err != nil {
		return modele.Coordonnees{}, false, err
	}
	if len(arr) == 0 {
		return modele.Coordonnees{}, false, nil
	}

	lat, err1 := strconvParse(arr[0].Lat)
	lon, err2 := strconvParse(arr[0].Lon)
	if err1 != nil || err2 != nil {
		return modele.Coordonnees{}, false, errors.New("coordonnées invalides")
	}

	coord := modele.Coordonnees{Lat: lat, Lon: lon}
	cache[lieu] = coord
	sauverCacheGeocodage(cache)

	// 1 requête/seconde max :contentReference[oaicite:4]{index=4}
	time.Sleep(1100 * time.Millisecond)

	return coord, false, nil
}

func strconvParse(s string) (float64, error) {
	// évite d’importer partout
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}
