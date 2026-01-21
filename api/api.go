package api

import (
	"encoding/json"
	"fmt"
	"groupie-tracker/modele"
	"net/http"
)

// RecupererJSON effectue une requête HTTP GET et décode la réponse JSON dans la cible fournie
// Gère les erreurs HTTP et de décodage de manière uniforme
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
