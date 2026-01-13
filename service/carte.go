package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"groupie-tracker/modele"
)

// Données envoyées au HTML (markers)
type marker struct {
	Lieu  string  `json:"lieu"`
	Lat   float64 `json:"lat"`
	Lon   float64 `json:"lon"`
	Popup string  `json:"popup"`
}

// Construit les marqueurs à partir relation.DatesParLieu
func ConstruireMarkers(relation modele.Relation) ([]marker, error) {
	cache := chargerCacheGeocodage()
	client := &http.Client{}

	var res []marker
	for lieu, dates := range relation.DatesParLieu {
		coord, _, err := GeocoderLieu(lieu, client, cache)
		if err != nil {
			continue // simple: on ignore les lieux qui échouent
		}
		if coord.Lat == 0 && coord.Lon == 0 {
			continue
		}

		popup := fmt.Sprintf("<b>%s</b><br/>%s",
			strings.ReplaceAll(lieu, "-", " "),
			strings.Join(dates, "<br/>"),
		)

		res = append(res, marker{
			Lieu:  lieu,
			Lat:   coord.Lat,
			Lon:   coord.Lon,
			Popup: popup,
		})
	}
	return res, nil
}

// Génère un HTML Leaflet (CDN) + markers
// Leaflet quickstart :contentReference[oaicite:6]{index=6}
func GenererFichierCarteHTML(nomArtiste string, markers []marker) (string, error) {
	data, _ := json.Marshal(markers)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="fr">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
<title>Carte - %s</title>

<link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css"/>
<script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js"></script>

<style>
  body { margin:0; font-family: Arial, sans-serif; }
  #map { height: 100vh; width: 100vw; }
  .footer { position: absolute; bottom: 8px; left: 8px; background: rgba(255,255,255,0.85); padding: 6px 8px; border-radius: 6px; font-size: 12px; }
</style>
</head>
<body>
<div id="map"></div>
<div class="footer">Données carte © OpenStreetMap contributors</div>

<script>
  const markers = %s;

  const map = L.map('map');
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 18,
    attribution: '&copy; OpenStreetMap contributors'
  }).addTo(map);

  if (markers.length === 0) {
    map.setView([48.8566, 2.3522], 5); // fallback: Paris
  } else {
    const bounds = [];
    markers.forEach(m => {
      const mk = L.marker([m.lat, m.lon]).addTo(map);
      mk.bindPopup(m.popup);
      bounds.push([m.lat, m.lon]);
    });
    map.fitBounds(bounds, { padding: [30, 30] });
  }
</script>
</body>
</html>`, nomArtiste, string(data))

	// On écrit dans un fichier local (dans le dossier du projet)
	chemin := filepath.Join(".", "carte.html")
	if err := os.WriteFile(chemin, []byte(html), 0644); err != nil {
		return "", err
	}
	return chemin, nil
}
