package service

import (
	"encoding/json"
	"os"
	"sync"

	"groupie-tracker/modele"
)

// GestionnaireFavoris gère la liste des artistes favoris
type GestionnaireFavoris struct {
	mutex   sync.RWMutex
	favoris map[int]modele.Artiste // ID -> Artiste
	fichier string
}

var (
	instance *GestionnaireFavoris
	once     sync.Once
)

// ObtenirGestionnaireFavoris retourne l'instance unique du gestionnaire
func ObtenirGestionnaireFavoris() *GestionnaireFavoris {
	once.Do(func() {
		instance = &GestionnaireFavoris{
			favoris: make(map[int]modele.Artiste),
			fichier: "favoris.json",
		}
		instance.charger()
	})
	return instance
}

// EstFavori vérifie si un artiste est dans les favoris
func (g *GestionnaireFavoris) EstFavori(id int) bool {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	_, existe := g.favoris[id]
	return existe
}

// Basculer ajoute ou retire un artiste des favoris
func (g *GestionnaireFavoris) Basculer(artiste modele.Artiste) bool {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if _, existe := g.favoris[artiste.ID]; existe {
		delete(g.favoris, artiste.ID)
		g.sauvegarder()
		return false // retiré
	}

	g.favoris[artiste.ID] = artiste
	g.sauvegarder()
	return true // ajouté
}

// ObtenirFavoris retourne la liste de tous les artistes favoris
func (g *GestionnaireFavoris) ObtenirFavoris() []modele.Artiste {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	liste := make([]modele.Artiste, 0, len(g.favoris))
	for _, artiste := range g.favoris {
		liste = append(liste, artiste)
	}
	return liste
}

// sauvegarder enregistre les favoris dans un fichier JSON
func (g *GestionnaireFavoris) sauvegarder() {
	data, err := json.MarshalIndent(g.favoris, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(g.fichier, data, 0644)
}

// charger récupère les favoris depuis le fichier JSON
func (g *GestionnaireFavoris) charger() {
	data, err := os.ReadFile(g.fichier)
	if err != nil {
		return // fichier n'existe pas encore
	}

	var favoris map[int]modele.Artiste
	if err := json.Unmarshal(data, &favoris); err != nil {
		return
	}

	g.favoris = favoris
}
