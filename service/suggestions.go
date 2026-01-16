package service

import (
	"sort"
	"strconv"
	"strings"
	"sync"

	"groupie-tracker/modele"
)

type SuggestionsStore struct {
	mu   sync.RWMutex
	list []modele.Suggestion
}

func NouveauSuggestionsStore() *SuggestionsStore {
	return &SuggestionsStore{list: []modele.Suggestion{}}
}

func (s *SuggestionsStore) Get() []modele.Suggestion {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]modele.Suggestion, len(s.list))
	copy(out, s.list)
	return out
}

func (s *SuggestionsStore) Set(l []modele.Suggestion) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.list = l
}

// --- Helpers ---

func extraireAnnee(texte string) int {
	texte = strings.TrimSpace(texte)
	if len(texte) < 4 {
		return 0
	}
	parties := strings.FieldsFunc(texte, func(r rune) bool {
		return r == '-' || r == '/' || r == '.'
	})
	for _, p := range parties {
		if len(p) == 4 {
			if y, err := strconv.Atoi(p); err == nil {
				return y
			}
		}
	}
	if y, err := strconv.Atoi(texte[:4]); err == nil {
		return y
	}
	return 0
}

func cleanLieu(lieu string) string {
	lieu = strings.ReplaceAll(lieu, "_", " ")
	lieu = strings.ReplaceAll(lieu, "-", " ")
	return strings.TrimSpace(lieu)
}

// ConstruireSuggestions : génère les suggestions de l'énoncé.
// relations peut être nil -> pas de lieux.
func ConstruireSuggestions(artistes []modele.Artiste, relations map[int]modele.Relation) []modele.Suggestion {
	out := make([]modele.Suggestion, 0, len(artistes)*6)

	ajouter := func(texte, typ string, id int) {
		texte = strings.TrimSpace(texte)
		if texte == "" {
			return
		}
		out = append(out, modele.Suggestion{
			Texte: texte,
			Type:  typ,
			ID:    id,
		})
	}

	for _, a := range artistes {
		// artiste/groupe
		ajouter(a.Nom, "artiste/groupe", a.ID)

		// membres
		for _, m := range a.Membres {
			ajouter(m, "membre", a.ID)
		}

		// premier album (date brute + année)
		ajouter(a.PremierAlbum, "premier album", a.ID)
		if y := extraireAnnee(a.PremierAlbum); y > 0 {
			ajouter(strconv.Itoa(y), "premier album", a.ID)
		}

		// date de création (année)
		ajouter(strconv.Itoa(a.AnneeCreation), "date de création", a.ID)

		// lieux (si relations chargées)
		if relations != nil {
			if rel, ok := relations[a.ID]; ok {
				for lieu := range rel.DatesParLieu {
					ajouter(cleanLieu(lieu), "lieu", a.ID)
				}
			}
		}
	}

	// Tri stable
	sort.SliceStable(out, func(i, j int) bool {
		ti := strings.ToLower(out[i].Texte)
		tj := strings.ToLower(out[j].Texte)
		if ti == tj {
			return out[i].Type < out[j].Type
		}
		return ti < tj
	})

	return out
}
