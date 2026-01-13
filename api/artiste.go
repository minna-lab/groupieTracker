package api

import (
	"fmt"
	"groupie-tracker/modele"
)

func RecupererArtistes() ([]modele.Artiste, error) {
	var artistes []modele.Artiste
	err := RecupererJSON(
		"https://groupietrackers.herokuapp.com/api/artists",
		&artistes,
	)
	return artistes, err
}

func RecupererRelation(id int) (modele.Relation, error) {
	var relation modele.Relation
	err := RecupererJSON(
		fmt.Sprintf("https://groupietrackers.herokuapp.com/api/relation/%d", id),
		&relation,
	)
	return relation, err
}
