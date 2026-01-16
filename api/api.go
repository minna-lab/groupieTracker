package api

import (
	"encoding/json"
	"fmt"
	"groupie-tracker/modele"
	"net/http"
)

// Client HTTP générique pour récupérer les données JSON
func RecupererJSON(url string, cible interface{}) error {
	reponse, err := http.Get(url)
	if err != nil {
		return err
	}
	defer reponse.Body.Close()

	if reponse.StatusCode != http.StatusOK {
		return fmt.Errorf("erreur HTTP : %s", reponse.Status)
	}

	err = json.NewDecoder(reponse.Body).Decode(cible)
	if err != nil {
		return err
	}

	return nil
}

// Récupération des artistes
func RecupererArtistes() ([]modele.Artiste, error) {
	var artistes []modele.Artiste
	err := RecupererJSON(
		"https://groupietrackers.herokuapp.com/api/artists",
		&artistes,
	)
	return artistes, err
}

// Récupération des relations pour un artiste
func RecupererRelation(id int) (modele.Relation, error) {
	var relation modele.Relation
	err := RecupererJSON(
		fmt.Sprintf("https://groupietrackers.herokuapp.com/api/relation/%d", id),
		&relation,
	)
	return relation, err
}
