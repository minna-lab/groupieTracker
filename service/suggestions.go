package service

import (
	"groupie-tracker/modele"
	"strconv"
	"strings"
)

func ConstruireSuggestions(artistes []modele.Artiste, cache *CacheRelations) []modele.Suggestion {
	vu := make(map[string]bool)
	var res []modele.Suggestion

	ajouter := func(texte, typ string, id int) {
		cle := strings.ToLower(texte) + "|" + typ
		if texte == "" || vu[cle] {
			return
		}
		vu[cle] = true
		res = append(res, modele.Suggestion{Texte: texte, Type: typ, ID: id})
	}

	for _, a := range artistes {
		ajouter(a.Nom, "artiste", a.ID)

		for _, m := range a.Membres {
			ajouter(m, "membre", a.ID)
		}

		ajouter(strconv.Itoa(a.AnneeCreation), "année de création", a.ID)
		ajouter(a.PremierAlbum, "premier album", a.ID)

		// lieux si déjà en cache (sinon ils viendront quand tu indexeras)
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
