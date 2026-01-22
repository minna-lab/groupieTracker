package api

import (
	"encoding/json"
	"fmt"
	"groupie-tracker/modele"
	"net/http"
)

func RecupererJSON(url string, cible interface{}) error {
	reponse, err := http.Get(url)
	if err != nil {
		return err
	}
	defer reponse.Body.Close()
	if reponse.StatusCode != http.StatusOK {
		return fmt.Errorf("erreur HTTP : %s", reponse.Status)
	}
	return json.NewDecoder(reponse.Body).Decode(cible)
}

// RecupererArtistes récupère la liste complète des artistes depuis l'API Groupie Tracker
// Retourne un tableau d'artistes avec leurs informations (nom, membres, année, etc.)
func RecupererArtistes() ([]modele.Artiste, error) {
	var artistes []modele.Artiste
	err := RecupererJSON(
		"https://groupietrackers.herokuapp.com/api/artists",
		&artistes,
	)
	return artistes, err
}

// RecupererRelation récupère les dates de concerts et lieux pour un artiste spécifique
// Prend l'ID de l'artiste et retourne une map associant chaque lieu à ses dates de concert
func RecupererRelation(id int) (modele.Relation, error) {
	var relation modele.Relation
	err := RecupererJSON(
		fmt.Sprintf("https://groupietrackers.herokuapp.com/api/relation/%d", id),
		&relation,
	)
	return relation, err
}

// RecupererLocations récupère les lieux de concerts disponibles depuis l'API
// Retourne une structure Lieux contenant tous les lieux de concerts
func RecupererLocations() ([]modele.Lieux, error) {
	var locations struct {
		Index []modele.Lieux `json:"index"`
	}
	err := RecupererJSON(
		"https://groupietrackers.herokuapp.com/api/locations",
		&locations,
	)
	return locations.Index, err
}

// RecupererDates récupère les dates de concerts disponibles depuis l'API
// Retourne un tableau de structures Dates contenant toutes les dates de concerts
func RecupererDates() ([]modele.Dates, error) {
	var dates struct {
		Index []modele.Dates `json:"index"`
	}
	err := RecupererJSON(
		"https://groupietrackers.herokuapp.com/api/dates",
		&dates,
	)
	return dates.Index, err
}
